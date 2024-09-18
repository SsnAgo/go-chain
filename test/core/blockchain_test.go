package test

import (
	"testing"
	"go-chain/core"
	"go-chain/types"
	"go-chain/cryptoo"
)

func TestNewBlockchain(t *testing.T) {
	bc := core.NewBlockchain()
	if bc == nil {
		t.Fatal("创建区块链失败")
	}
	// 创建创世区块
	genesisBlock := core.NewBlock(types.Hash{}, 0, []*core.Transaction{})
	if genesisBlock == nil {
		t.Fatal("获取创世区块失败")
	}
	bc.AddBlockWithoutValidation(genesisBlock)
	
	// 验证创世区块
	if genesisBlock.Height() != 0 {
		t.Errorf("创世区块高度应为0，实际为%d", genesisBlock.Height())
	}
	
	if !genesisBlock.GetPrevBlockHash().IsZero() {
		t.Error("创世区块的前一个区块哈希应为零值")
	}
	
	if len(genesisBlock.Transactions) != 0 {
		t.Errorf("创世区块不应包含交易，实际包含%d笔交易", len(genesisBlock.Transactions))
	}
	if bc.Height() != 0 {
		t.Errorf("期望初始高度为0，实际为%d", bc.Height())
	}
}

func TestAddBlock(t *testing.T) {
	bc := core.NewBlockchain()
	
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()
	// 添加创世区块
	genesisBlock := core.NewBlock(types.Hash{}, 0, []*core.Transaction{})
	bc.AddBlockWithoutValidation(genesisBlock)

	tx := core.NewTransaction(pv1, pb2, []byte("test transaction"), 100, 0)
	block := core.NewBlock(genesisBlock.GetDataHash(), 1, []*core.Transaction{tx})

	err := bc.AddBlock(block)
	if err != nil {
		t.Fatalf("添加区块失败：%v", err)
	}

	if bc.Height() != 1 {
		t.Errorf("期望区块链高度为1，实际为%d", bc.Height())
	}
}

func TestGetBlock(t *testing.T) {
	bc := core.NewBlockchain()
	
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	bc.GetAccountState().CreateAccount(pv1.GetPublicKey().Address(), &core.Account{
		Address: pv1.GetPublicKey().Address(),
		Balance: 1000,
	})

	genesisBlock := core.NewBlock(types.Hash{}, 0, []*core.Transaction{})
	bc.AddBlockWithoutValidation(genesisBlock)

	tx := core.NewTransaction(pv1, pb2, []byte("test transaction"), 100, 0)
	block := core.NewBlock(genesisBlock.GetDataHash(), 1, []*core.Transaction{tx})

	err := bc.AddBlock(block)
	if err != nil {
		t.Fatalf("添加区块失败：%v", err)
	}

	retrievedBlock, err := bc.GetBlock(1)
	if err != nil {
		t.Fatalf("获取区块失败：%v", err)
	}

	if retrievedBlock.GetDataHash() != block.GetDataHash() {
		t.Error("获取的区块与添加的区块不匹配")
	}
}

func TestBlockchainIntegration(t *testing.T) {
	bc := core.NewBlockchain()
	
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb2 := pv2.GetPublicKey()

	bc.GetAccountState().CreateAccount(pv1.GetPublicKey().Address(), &core.Account{
		Address: pv1.GetPublicKey().Address(),
		Balance: 1000000,
	})

	genesisBlock := core.NewBlock(types.Hash{}, 0, []*core.Transaction{})
	bc.AddBlockWithoutValidation(genesisBlock)

	// 添加多个区块
	for i := 1; i <= 5; i++ {
		tx := core.NewTransaction(pv1, pb2, []byte("transaction"), uint64(100*i), int64(i-1))
		block := core.NewBlock(bc.GetLatestBlock().GetDataHash(), uint32(i), []*core.Transaction{tx})
		err := bc.AddBlock(block)
		if err != nil {
			t.Fatalf("添加第%d个区块失败：%v", i, err)
		}
	}

	// 验证区块链高度
	if bc.Height() != 5 {
		t.Errorf("期望区块链高度为5，实际为%d", bc.Height())
	}

	// 验证区块链完整性
	for i := 1; i <= 5; i++ {
		block, err := bc.GetBlock(uint32(i))
		if err != nil {
			t.Fatalf("获取第%d个区块失败：%v", i, err)
		}
		if block.Height() != uint32(i) {
			t.Errorf("第%d个区块高度不正确", i)
		}
		if i > 1 {
			prevBlock, _ := bc.GetBlock(uint32(i - 1))
			if block.GetPrevBlockHash() != prevBlock.GetDataHash() {
				t.Errorf("第%d个区块的前一个区块哈希不正确", i)
			}
		}
	}
}
