package test

import (
	"fmt"
	"go-chain/core"
	"go-chain/cryptoo"
	"testing"
)

func TestNewTxPool(t *testing.T) {
	allSize := 100
	pendingSize := 50
	pool := core.NewTxPool(allSize, pendingSize)

	if pool == nil {
		t.Fatal("创建交易池失败")
	}

	if pool.GetMaxAllSize() != allSize {
		t.Errorf("期望全部交易池大小为 %d，实际为 %d", allSize, pool.GetAllSize())
	}

	if pool.GetMaxPendingSize() != pendingSize {
		t.Errorf("期望待处理交易池大小为 %d，实际为 %d", pendingSize, pool.GetPendingSize())
	}
}

func TestAddTransaction(t *testing.T) {
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()
	
	pool := core.NewTxPool(100, 50)
	tx := core.NewTransaction(pv1, pb2, []byte("test"), 100, 0)

	err := pool.Add([]*core.Transaction{tx})
	if err != nil {
		t.Fatalf("添加交易失败：%v", err)
	}

	pendingTxs := pool.GetPendingTxs()
	if len(pendingTxs) != 1 {
		t.Errorf("期望待处理交易数量为 1，实际为 %d", len(pendingTxs))
	}

	if pendingTxs[0] != tx {
		t.Error("获取的待处理交易与添加的交易不匹配")
	}
}

func TestGetPendingTxsForPacking(t *testing.T) {
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	pool := core.NewTxPool(100, 50)
	tx1 := core.NewTransaction(pv1, pb2, []byte("test1"), 100, 0)
	tx2 := core.NewTransaction(pv1, pb2, []byte("test2"), 200, 1)

	pool.Add([]*core.Transaction{tx1, tx2})

	packedTxs, err := pool.GetPendingTxsForPacking()
	if err != nil {
		t.Fatalf("获取待打包交易失败：%v", err)
	}

	if len(packedTxs) != 2 {
		t.Errorf("期望待打包交易数量为 2，实际为 %d", len(packedTxs))
	}

	pendingTxs := pool.GetPendingTxs()
	if len(pendingTxs) != 0 {
		t.Errorf("打包后待处理交易池应为空，实际包含 %d 个交易", len(pendingTxs))
	}
}

func TestRemovePendingTxs(t *testing.T) {

	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	
	pool := core.NewTxPool(100, 50)
	tx1 := core.NewTransaction(pv1, pb2, []byte("test1"), 100, 0)
	tx2 := core.NewTransaction(pv1, pb2, []byte("test2"), 200, 1)
	tx3 := core.NewTransaction(pv1, pb2, []byte("test3"), 300, 2)

	pool.Add([]*core.Transaction{tx1, tx2, tx3})

	pool.RemovePendingTxs([]*core.Transaction{tx1, tx2})

	pendingTxs := pool.GetPendingTxs()
	if len(pendingTxs) != 1 {
		t.Errorf("期望待处理交易数量为 1，实际为 %d", len(pendingTxs))
	}

	if pendingTxs[0] != tx3 {
		t.Error("剩余的待处理交易应为 tx3")
	}
}

func TestTxPoolIntegration(t *testing.T) {
	pool := core.NewTxPool(100, 50)

	// 添加交易
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	tx1 := core.NewTransaction(pv1, pb2, []byte("test1"), 100, 0)
	tx2 := core.NewTransaction(pv1, pb2, []byte("test2"), 200, 1)
	tx3 := core.NewTransaction(pv1, pb2, []byte("test3"), 300, 2)

	pool.Add([]*core.Transaction{tx1, tx2, tx3})

	// 验证待处理交易
	pendingTxs := pool.GetPendingTxs()
	if len(pendingTxs) != 3 {
		t.Errorf("期望待处理交易数量为 3，实际为 %d", len(pendingTxs))
	}

	// 获取待打包交易
	packedTxs, err := pool.GetPendingTxsForPacking()
	if err != nil {
		t.Fatalf("获取待打包交易失败：%v", err)
	}
	if len(packedTxs) != 3 {
		t.Errorf("期望待打包交易数量为 3，实际为 %d", len(packedTxs))
	}

	// 验证待处理交易池已清空
	pendingTxs = pool.GetPendingTxs()
	if len(pendingTxs) != 0 {
		t.Errorf("打包后待处理交易池应为空，实际包含 %d 个交易", len(pendingTxs))
	}

	// 添加新交易
	tx4 := core.NewTransaction(pv1, pb2, []byte("test4"), 400, 3)
	pool.Add([]*core.Transaction{tx4})

	// 移除部分待处理交易
	pool.RemovePendingTxs([]*core.Transaction{tx4})

	// 验证最终状态
	pendingTxs = pool.GetPendingTxs()
	if len(pendingTxs) != 0 {
		t.Errorf("期望待处理交易数量为 0，实际为 %d", len(pendingTxs))
	}
}

func TestTxPoolEvictionAndSorting(t *testing.T) {
	// 创建一个小容量的交易池，以便于测试淘汰策略
	pool := core.NewTxPool(5, 5)

	// 生成测试用的私钥和公钥
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	// 创建6个交易，value从100到600
	txs := make([]*core.Transaction, 6)
	for i := 0; i < 6; i++ {
		txs[i] = core.NewTransaction(pv1, pb2, []byte(fmt.Sprintf("test%d", i+1)), uint64((i+1)*100), int64(i))
	}

	// 按照随机顺序添加交易
	randomOrder := []int{3, 0, 5, 2, 1, 4}
	for _, i := range randomOrder {
		pool.Add([]*core.Transaction{txs[i]})
	}

	// 验证池子中的交易数量
	allTxs := pool.GetAllTxs()
	if len(allTxs) != 5 {
		t.Errorf("期望交易池中有5个交易，实际有%d个", len(allTxs))
	}

	// 验证被淘汰的是value最小的交易
	if pool.Get(txs[0].CalHash()) != nil {
		t.Errorf("value最小的交易应该被淘汰")
	}

	// 验证value最大的5个交易都在池子中
	for i := 1; i < 6; i++ {
		if pool.Get(txs[i].CalHash()) == nil {
			t.Errorf("交易%d应该在池子中", i+1)
		}
	}
}
