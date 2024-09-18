package core

import (
	"container/heap"
	"errors"
	"go-chain/types"
	"sync"

	"github.com/samber/lo"
)

var defaultSortFunc = func(a, b *Transaction) bool {
	return a.Value < b.Value
}

var (
	ErrTxAlreadyInPool = errors.New("交易已存在于交易池中")
	ErrPoolIsFull      = errors.New("交易池已满")
)

// TxPool 表示交易池
type TxPool struct {
	mu       sync.RWMutex
	all      *TxSortedStore
	pending  *TxSortedStore
	allSize  int
	pendingSize int
}

func (pool *TxPool) GetAllSize() int {
	return pool.all.Len()
}

func (pool *TxPool) GetMaxAllSize() int {
	return pool.all.GetMaxSize()
}


func (pool *TxPool) GetPendingSize() int {
	return pool.pending.Len()
}

func (pool *TxPool) GetMaxPendingSize() int {
	return pool.pending.GetMaxSize()
}

// NewTxPool 创建一个新的交易池

// NewTxPool 创建一个新的交易池
func NewTxPool(allSize, pendingSize int) *TxPool {
	return &TxPool{
		mu:       sync.RWMutex{},
		all:      NewTxSortedStore(allSize),
		pending:  NewTxSortedStore(pendingSize),
		allSize:  allSize,
		pendingSize: pendingSize,
	}
}

// Add 向交易池中添加交易
func (pool *TxPool) Add(txs []*Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	lo.ForEach(txs, func(tx *Transaction, _ int) {
		pool.all.Add(tx)
		pool.pending.Add(tx)
	})

	return nil
}


// GetPendingTxs 获取待处理的交易
func (pool *TxPool) GetPendingTxs() (pendings []*Transaction) {
	pendings = append(pendings, pool.pending.GetAll()...)
	return pendings
}

// GetPendingTxs 获取待处理的交易
func (pool *TxPool) GetAllTxs() (all []*Transaction) {
	all = append(all, pool.all.GetAll()...)
	return all
}

// 从pending里打包交易 并清空pending
func (pool *TxPool) GetPendingTxsForPacking() (packedTxs []*Transaction, err error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.pending.Len() == 0 {
		return nil, nil
	}
	packedTxs = pool.GetPendingTxs()
	pool.ClearPending()

	return packedTxs, nil
}

func (pool *TxPool) RemovePendingTxs(txs []*Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	// 从pending里删除交易
	for _, tx := range txs {
		pool.pending.Remove(tx)
	}
	newTxs := make([]*Transaction, 0, len(pool.pending.lookup))
	for _, tx := range pool.pending.lookup {
		newTxs = append(newTxs, tx)
	}
	pool.pending.ResetTree(newTxs)
}


func (pool *TxPool) Get(hash types.Hash) *Transaction {
	return pool.all.Get(hash)
}

func (pool *TxPool) ClearPending() {
	pool.pending.Clear()
}

func (pool *TxPool) ClearAll() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.all.Clear()
}


type TxSortedStore struct {
	mu sync.RWMutex
	lookup map[types.Hash]*Transaction
	txx *types.MinHeap[*Transaction]
	maxSize int
}

// NewTxSortedStore 创建一个新的TxSortedStore
func NewTxSortedStore(maxSize int) *TxSortedStore {
	return &TxSortedStore{
		mu: sync.RWMutex{},
		lookup: make(map[types.Hash]*Transaction),
		txx: types.NewMinHeap[*Transaction](make([]*Transaction, 0), defaultSortFunc),
		maxSize: maxSize,
	}
}

// Add 添加一个交易到存储中 维持前maxsize个value的交易
func (s *TxSortedStore) Add(tx *Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := tx.CalHash()
	// 每次检查长度 pop时 保持loopup和tree里面的交易是一致的
	if _, exists := s.lookup[hash]; exists {
		return 
	}
	// 长度大于的时候  顶掉
	if s.txx.Len() >= s.maxSize {
		// 如果比堆顶还小 那就不插入了
		
		peeked := s.txx.Peek().(*Transaction)
		if tx.Value <= peeked.Value {
			return 
		}
		poped := heap.Pop(s.txx).(*Transaction)
		delete(s.lookup, poped.CalHash())
	}
	s.lookup[hash] = tx
	heap.Push(s.txx, tx)
}

func (s *TxSortedStore) Get(hash types.Hash) *Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lookup[hash]
}

func (s *TxSortedStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.txx.Len()
}

func (s *TxSortedStore) GetAll() []*Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.txx.GetAll()
}

func (s *TxSortedStore) Remove(tx *Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lookup, tx.CalHash())
}

func (s *TxSortedStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 清空lookup和tree
	s.lookup = make(map[types.Hash]*Transaction)
	s.txx = types.NewMinHeap(make([]*Transaction, 0), defaultSortFunc)
}

func (s *TxSortedStore) ResetTree(txs []*Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.txx = types.NewMinHeap(txs, defaultSortFunc)
}

func (s *TxSortedStore) GetMaxSize() int {
	return s.maxSize
}