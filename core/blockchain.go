package core

import (
	"errors"
	"go-chain/inter"
	"go-chain/types"
	"go-chain/utils"
	"log"
	"os"
	"sync"
)

var (
	ErrBlockNotFound = errors.New("区块未找到")
	ErrChainNotFound = errors.New("链未找到")
)

// Blockchain 表示整个区块链
// 假设这里区块链并没有分叉
type Blockchain struct {
	logger       log.Logger
	mu           sync.RWMutex
	blocks       []*Block
	headers      []*BlockHeader
	blockStore   map[types.Hash]*Block
	txStore      map[types.Hash]*Transaction
	accountState *AccountState
	stateLock    sync.RWMutex
	validator    inter.Validator
}

// NewBlockchain 创建一个新的区块链
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		logger:       *log.New(os.Stdout, "Blockchain", log.LstdFlags),
		mu:           sync.RWMutex{},
		blocks:       make([]*Block, 0),
		headers:      make([]*BlockHeader, 0),
		blockStore:   make(map[types.Hash]*Block),
		txStore:      make(map[types.Hash]*Transaction),
		accountState: NewAccountState(),
		stateLock:    sync.RWMutex{},
		validator:    nil,
	}
	v := &BlockValidator{bc: bc}
	bc.validator = v

	return bc
}

// AddBlock 向区块链中添加一个新区块
// 区块必须先验证过
func (bc *Blockchain) AddBlock(block *Block) error {

	if !bc.validator.Validate(*block) {
		return errors.New("区块验证失败")
	}

	bc.addBlock(block)
	return nil
}

// AddBlock 向区块链中添加一个新区块
// 用于直接添加区块 在同步别的节点的block时，无需再验证每个区块
func (bc *Blockchain) AddBlockWithoutValidation(block *Block) {
	bc.addBlock(block)
}

// GetBlock 根据高度获取区块
func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height >= uint32(len(bc.blocks)) || height < 1 {
		bc.logger.Printf("区块高度 %d 超出范围", height)
		return nil, ErrBlockNotFound
	}
	block := bc.blocks[height]
	return block, nil
}

// GetBlock 根据高度获取区块
func (bc *Blockchain) GetBlockByHash(hash types.Hash) *Block {
	return bc.blockStore[hash]
}

// GetLatestBlock 获取最新的区块
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks[len(bc.blocks)-1]
}

func (bc *Blockchain) Height() uint32 {
	// 区块高度为区块数量减1  因为初始创世区块的高度是0 每个区块链必然有一个创世区块
	return uint32(len(bc.blocks) - 1)
}

// addBlock 将区块添加到区块链中
func (bc *Blockchain) addBlock(block *Block) {
	// 先执行这个区块的所有交易
	bc.stateLock.Lock()
	for i, tx := range block.Transactions {
		if err := bc.ExecuteTransaction(tx); err != nil {
			// 将这个交易删除 是通过与最后的Tx进行交换，然后删除最后的Tx
			block.Transactions[i] = block.Transactions[len(block.Transactions)-1]
			block.Transactions = block.Transactions[:len(block.Transactions)-1]
			continue
		}
	}
	bc.stateLock.Unlock()


	bc.mu.Lock()
	// 将区块添加到存储中
	bc.blocks = append(bc.blocks, block)
	bc.headers = append(bc.headers, block.Header)
	bc.blockStore[block.Header.DataHash] = block

	// 将交易也加到区块链中
	for _, tx := range block.Transactions {
		// todo 要执行区块中每个交易
		bc.txStore[tx.CalHash()] = tx
	}
	bc.mu.Unlock()

	bc.logger.Println(
		"msg", "new block",
		"hash", block.GetDataHash(),
		"height", block.Height(),
		"transactions", len(block.Transactions),
	)
}

// HasBlock 检查区块链中是否存在指定哈希的区块
func (bc *Blockchain) HasBlock(hash types.Hash) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	
	_, exists := bc.blockStore[hash]
	return exists
}


func (bc *Blockchain) getRangeBlocks(from uint32, to uint32) []*Block {
	return bc.blocks[utils.HeightToIndex(from): utils.HeightToIndex(to) + 1]
}

func (bc *Blockchain) GetRangeBlocks(from uint32, to uint32) []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if !bc.checkGetBlockRange(from, to) {
		return nil
	}

	return bc.getRangeBlocks(from, to)
}



func (bc *Blockchain) checkGetBlockRange(from uint32, to uint32) bool {
	if from > to {
		bc.logger.Printf("获取区块消息的from参数错误: %v > %v", from, to)
		return false
	}
	if from <= 0 || from > bc.Height() {
		bc.logger.Printf("获取区块消息的from参数错误: %v > %v", from, bc.Height())
		return false
	}
	if to > bc.Height() {
		bc.logger.Printf("获取区块消息的to参数错误: %v > %v", to, bc.Height())
		return false
	}
	return true
}

// ExecuteTransaction 执行交易并更新账户状态
func (bc *Blockchain) ExecuteTransaction(tx *Transaction) error {
	// 验证交易
	if !tx.Verify() {
		return errors.New("交易验证失败")
	}
	err := bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
	if err != nil {
		bc.logger.Printf("交易执行失败: %v", err)
		return err
	}

	// 记录交易
	bc.txStore[tx.CalHash()] = tx

	bc.logger.Printf("交易执行成功: 从 %s 转账 %d 到 %s", tx.From, tx.Value, tx.To)
	return nil
}




func (bc *Blockchain) GetAccountState() *AccountState {
	return bc.accountState
}


// 
func (bc *Blockchain) RemoveBlocks(toHeight uint32) {
	bc.blocks = bc.blocks[:toHeight]
}


func (bc *Blockchain) DeleteTxs(txs []*Transaction) {
	for _, tx := range txs {
		delete(bc.txStore, tx.CalHash())
	}
}

func (bc *Blockchain) DeleteBlockStore(hash types.Hash) {
	delete(bc.blockStore, hash)
}
