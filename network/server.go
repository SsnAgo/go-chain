package network

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"go-chain/core"
	"go-chain/cryptoo"
	"go-chain/utils"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"github.com/samber/lo"
)

var (
	ErrNoSeedNodes = errors.New("no seed nodes provided")
)

type Server struct {
	opts         ServerOpts
	chain        *core.Blockchain
	mu           sync.RWMutex
	rpcCh        chan RPC
	quitCh       chan struct{}
	peerMap      map[net.Addr]*TCPPeer
	tcpTransport *TCPTransport
	priv         *cryptoo.PrivateKey
	pool         *core.TxPool
}

type ServerOpts struct {
	id               string
	listenAddr       string
	log              *log.Logger
	seedNodes        []string
	privKey          string
	allPoolLimit     uint32
	pendingPoolLimit uint32
}

type ServerOption func(*ServerOpts)

func WithListenAddr(addr string) ServerOption {
	return func(opts *ServerOpts) {
		opts.listenAddr = addr
	}
}

func WithLogger(logger *log.Logger) ServerOption {
	return func(opts *ServerOpts) {
		opts.log = logger
	}
}

func WithSeedNodes(nodes []string) ServerOption {
	return func(opts *ServerOpts) {
		opts.seedNodes = nodes
	}
}

func WithPrivKey(key string) ServerOption {
	return func(opts *ServerOpts) {
		opts.privKey = key
	}
}

func WithAllPoolLimit(size uint32) ServerOption {
	return func(opts *ServerOpts) {
		opts.allPoolLimit = size
	}
}

func WithPendingPoolLimit(size uint32) ServerOption {
	return func(opts *ServerOpts) {
		opts.pendingPoolLimit = size
	}
}

func NewServerOpts(options ...ServerOption) *ServerOpts {
	opts := &ServerOpts{}
	for _, option := range options {
		option(opts)
	}
	return opts
}

// NewServer 只是做一个初始化配置的操作 不做启动
func NewServer(opts ServerOpts) (s *Server, err error) {
	if opts.id == "" {
		opts.id = utils.RandID()
	}
	if opts.log == nil {
		opts.log = log.New(os.Stdout, "", log.LstdFlags)
	}
	if opts.listenAddr == "" {
		opts.listenAddr = ":9977"
	}
	if len(opts.seedNodes) == 0 {
		return nil, ErrNoSeedNodes
	}
	if opts.allPoolLimit == 0 {
		opts.allPoolLimit = 10000
	}
	if opts.pendingPoolLimit == 0 {
		opts.pendingPoolLimit = 4096
	}
	// 使用传递的私钥或者生成新私钥
	priv, err := cryptoo.UseOrGenPrivateKey(opts.privKey)
	if err != nil {
		return nil, err
	}

	tcpT, err := NewTCPTransport(opts.listenAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		opts:         opts,
		mu:           sync.RWMutex{},
		rpcCh:        make(chan RPC),
		quitCh:       make(chan struct{}),
		peerMap:      make(map[net.Addr]*TCPPeer),
		chain:        core.NewBlockchain(),
		tcpTransport: tcpT,
		priv:         priv,
		pool:         core.NewTxPool(int(opts.allPoolLimit), int(opts.pendingPoolLimit)),
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动TCP传输层
	if err := s.tcpTransport.Start(); err != nil {
		return fmt.Errorf("启动TCP传输层失败: %v", err)
	}

	// 启动接受循环
	go s.acceptLoop()

	// 连接种子节点
	for _, addr := range s.opts.seedNodes {
		if err := s.connectToNode(addr); err != nil {
			s.logf("连接种子节点 %s 失败: %v", addr, err)
		}
	}

	s.opts.log.Printf("服务器已在 %s 启动", s.opts.listenAddr)
	// 启动同步块协程
	go s.syncBlocksLoop()

	return nil
}

// logf 封装打印server日志的函数
func (s *Server) logf(format string, v ...interface{}) {
	if s.opts.log != nil {
		s.opts.log.Printf(format, v...)
	}
}

// connectToNode 连接到指定地址的节点
func (s *Server) connectToNode(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := NewTCPPeer(conn, true)
	s.mu.Lock()
	s.peerMap[conn.RemoteAddr()] = peer
	s.mu.Unlock()

	go s.handlePeer(peer)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		select {
		case peer := <-s.tcpTransport.peerCh:
			s.mu.Lock()
			s.peerMap[peer.conn.RemoteAddr()] = peer
			s.mu.Unlock()

			go s.handlePeer(peer)
		case rpc := <-s.rpcCh:
			// 处理接收到的RPC请求
			s.handleRPCRequest(rpc)
		case <-s.quitCh:
			return
		}
	}
}

func (s *Server) handlePeer(peer *TCPPeer) {
	defer func() {
		s.mu.Lock()
		delete(s.peerMap, peer.conn.RemoteAddr())
		s.mu.Unlock()
	}()

	peer.ReceiveLoop(s.rpcCh)
}

func (s *Server) handleRPCRequest(rpc RPC) {
	s.logf("收到来自 %s 的RPC请求", rpc.From)

	// 解码RPC请求
	var req Message
	if err := gob.NewDecoder(rpc.Payload).Decode(&req); err != nil {
		s.logf("解码RPC请求失败: %v", err)
		return
	}

	// 根据请求类型处理
	switch req.Type {
	case MessageTypeTx:
		go s.handleTxMessage(rpc.From, req.Body)
	case MessageTypeBlock:
		go s.handleBlockMessage(rpc.From, req.Body)
	case MessageTypeGetBlocks:
		go s.handleGetBlocksMessage(rpc.From, req.Body)
	case MessageTypeStatus:
		go s.handleStatusMessage(rpc.From, req.Body)
	case MessageTypeGetStatus:
		go s.handleGetStatusMessage(rpc.From)
	case MessageTypeBlocks:
		go s.handleBlocksMessage(rpc.From, req.Body)
	default:
		s.logf("未知的RPC请求类型: %v", req.Type)
	}
}

func (s *Server) handleTxMessage(from net.Addr, body []byte) {
	s.logf("处理来自 %s 的交易消息", from)
	tx := new(core.Transaction)
	b := bytes.NewBuffer(body)
	if err := tx.Decode(b); err != nil {
		s.logf("解析交易消息失败: %v", err)
		return
	}
	// TODO: 处理交易逻辑
	// 加入到池子里 然后广播这个交易
	if err := s.pool.Add([]*core.Transaction{tx}); err != nil {
		s.logf("加入交易到池子失败: %v", err)
		return
	}
	// 广播交易
	go s.broadcastTx(tx)
}

func (s *Server) handleBlockMessage(from net.Addr, body []byte) {
	s.logf("处理来自 %s 的区块消息", from)
	block := new(core.Block)
	b := bytes.NewBuffer(body)
	if err := block.Decode(b); err != nil {
		s.logf("解析区块消息失败: %v", err)
		return
	}
	// TODO: 处理区块逻辑
	// 传递过来一个区块
	// 校验并添加
	if err := s.chain.AddBlock(block); err != nil {
		s.logf("添加区块失败: %v", err)
		return
	}

	// 从交易池中移除已确认的交易
	s.pool.RemovePendingTxs(block.Transactions)

	// 广播新区块给其他节点
	go s.broadcastBlock(block)
}

func (s *Server) handleGetBlocksMessage(from net.Addr, body []byte) {
	s.logf("处理来自 %s 的获取区块消息", from)
	getBs := new(GetBlocksMessage)
	b := bytes.NewBuffer(body)
	if err := getBs.Decode(b); err != nil {
		s.logf("解析获取区块消息失败: %v", err)
		return
	}
	// TODO: 处理获取区块逻辑
	blocks := s.chain.GetRangeBlocks(getBs.From, getBs.To)
	Bs := new(BlocksMessage)
	Bs.Blocks = blocks
	bb := bytes.NewBuffer(nil)
	if err := Bs.Encode(bb); err != nil {
		s.logf("编码区块消息失败: %v", err)
		return
	}
	// 将数据发送回去
	s.send(from, bb.Bytes())
}

func (s *Server) handleStatusMessage(from net.Addr, body []byte) {
	s.logf("处理来自 %s 的状态消息", from)
	status := new(StatusMessage)
	b := bytes.NewBuffer(body)
	if err := status.Decode(b); err != nil {
		s.logf("解析状态消息失败: %v", err)
		return
	}
	// TODO: 处理状态消息逻辑
	// 判断这个区块是不是领先自己
	// todo 还需要判断版本号
	if status.CurrentHeight > s.chain.Height() {
		// 发送获取区块消息
		getBs := new(GetBlocksMessage)
		getBs.From = s.chain.Height() + 1
		getBs.To = status.CurrentHeight
		bb := bytes.NewBuffer(nil)
		if err := getBs.Encode(bb); err != nil {
			s.logf("编码获取区块消息失败: %v", err)
		}
		s.send(from, bb.Bytes())
	}
}

func (s *Server) handleGetStatusMessage(from net.Addr) {
	s.logf("处理来自 %s 的获取状态消息", from)
	// 将状态发送回去
	sm := StatusMessage{
		ID:            s.opts.id,
		CurrentHeight: s.chain.Height(),
	}
	bb := &bytes.Buffer{}
	if err := sm.Encode(bb); err != nil {
		s.logf("编码状态消息失败: %v", err)
		return
	}
	s.send(from, bb.Bytes())
}

func (s *Server) handleBlocksMessage(from net.Addr, body []byte) {
	s.logf("处理来自 %s 的区块列表消息", from)
	bm := new(BlocksMessage)
	b := bytes.NewBuffer(body)
	if err := bm.Decode(b); err != nil {
		s.logf("解析区块列表消息失败: %v", err)
		return
	}
	// TODO: 处理区块列表逻辑
	if len(bm.Blocks) == 0  || bm.FirstBlock().Height() > s.chain.Height() + 1 ||
	bm.LastBlock().Height() <= s.chain.Height() {
		s.logf("区块列表消息无效")
		return 
	}
	// 同步区块时考虑一种情况 那就是如果遇到更长的链 但是交集的有一部分是不同的 应该从最近的共同区块开始同步
	// 找到最近的共同区块
	startHeight := s.chain.Height()
	startBmIdx := len(bm.Blocks) - 1
	for i := len(bm.Blocks) - 1; i >= 0; i-- {
		block := bm.Blocks[i]
		if preb := s.chain.GetBlockByHash(block.Header.PrevBlockHash); preb != nil {
			startHeight = preb.Height()
			startBmIdx = i
			break
		}
	}

	s.mu.Lock()
	// 移除不需要的区块以及交易 同时包含交易回滚 放回池子 等操作
	s.RollBlockRange(startHeight + 1)
	for i := startBmIdx; i < len(bm.Blocks); i++ {
		block := bm.Blocks[i]
		if err := s.chain.AddBlock(block); err != nil {
			s.logf("添加区块失败: %v", err)
			return
		}
	}
	s.mu.Unlock()

}

func (s *Server) broadcastTx(tx *core.Transaction) {
	var buf bytes.Buffer
	if err := tx.Encode(&buf); err != nil {
		s.logf("编码交易失败: %v", err)
		return
	}
	s.boardcast(buf.Bytes())
}

func (s *Server) boardcast(data []byte) {
	for addr, peer := range s.peerMap {
		s.logf("向 %s 广播消息", addr)
		if err := peer.Send(data); err != nil {
			s.logf("向 %s 发送消息失败: %v", addr, err)
		}
	}
}

func (s *Server) broadcastBlock(block *core.Block) {
	var buf bytes.Buffer
	if err := block.Encode(&buf); err != nil {
		s.logf("编码区块失败: %v", err)
		return
	}
	s.boardcast(buf.Bytes())
}

func (s *Server) send(toAddr net.Addr, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	peer, ok := s.peerMap[toAddr]
	if !ok {
		s.logf("未找到节点 %s", toAddr)
		return nil
	}
	return peer.Send(data)
}

func (s *Server) syncBlocksLoop() error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			for addr, peer := range s.peerMap {
				s.logf("向 %s 发送 GetStatus 消息", addr)
				msg := &GetStatusMessage{}
				var buf bytes.Buffer
				if err := msg.Encode(&buf); err != nil {
					s.logf("编码 GetStatus 消息失败: %v", err)
					continue
				}
				if err := peer.Send(buf.Bytes()); err != nil {
					s.logf("向 %s 发送 GetStatus 消息失败: %v", addr, err)
				}
			}
			s.mu.RUnlock()
		case <-s.quitCh:
			return nil
		}
	}
}


// RemoveBlockRange 移除从from开始的区块，回滚交易，并返回被回滚的交易
func (s *Server) RollBlockRange(fromHeight uint32) {
	if fromHeight > s.chain.Height() {
		s.logf("移除区块范围的起始高度无效: %d > %d", fromHeight, s.chain.Height())
		return 
	}

	rolledBackTxs := []*core.Transaction{}

	removeBlocks := s.chain.GetRangeBlocks(fromHeight, s.chain.Height())

	for _, rmb := range lo.Reverse(removeBlocks) {
		for _, tx := range rmb.Transactions {
			err := s.chain.GetAccountState().Transfer(tx.To.Address(), tx.From.Address(), tx.Value)
			// 回滚的时候应该不会有这个错误 但是这里还是检查一下
			if err != nil {
				s.logf("回滚交易失败: %v", err)
				continue
			}
			// delete(bc.txStore, tx.CalHash())
			s.chain.DeleteTxs([]*core.Transaction{tx})
			rolledBackTxs = append(rolledBackTxs, tx)
		}
		// delete(bc.blockStore, rmb.GetDataHash())
		s.chain.DeleteBlockStore(rmb.GetDataHash())
	}
	// 从blocks中移除区块
	s.chain.RemoveBlocks(fromHeight)

	// 将回滚的交易重新放回池子里
	
	s.pool.Add(rolledBackTxs)


	s.logf("成功移除从高度 %d 开始的区块，共回滚 %d 笔交易", fromHeight, len(rolledBackTxs))
}