package test

import (
	"bytes"
	"fmt"
	"go-chain/core"
	"go-chain/cryptoo"
	"go-chain/types"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {

	prevHash := types.RandomHash()
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb1 := cryptoo.PublicKey(pv1.GetPublicKey())[:32]
	pb2 := cryptoo.PublicKey(pv2.GetPublicKey())[:32]


	fmt.Println("pv1", pv1)
	fmt.Println("pb1", pb1)
	fmt.Println("pv2", pv2)
	fmt.Println("pb2", pb2)
	

	transactions := []*core.Transaction{
		core.NewTransaction(pv1, pb2, []byte("transaction1"), 100, 0),
		core.NewTransaction(pv2, pb1, []byte("transaction2"), 200, 0),
	}
	height := uint32(1)

	block := core.NewBlock(prevHash, height, transactions)

	if block.GetPrevBlockHash() != prevHash {
		t.Errorf("期望前一个区块哈希为 %v，实际为 %v", prevHash, block.GetPrevBlockHash())
	}
	if block.Height() != height {
		t.Errorf("期望区块高度为 %d，实际为 %d", height, block.Height())
	}
	if len(block.Transactions) != len(transactions) {
		t.Errorf("期望交易数量为 %d，实际为 %d", len(transactions), len(block.Transactions))
	}
	if block.Timestamp() == 0 {
		t.Error("区块时间戳不应为零")
	}
}

func TestBlockEncodeDecode(t *testing.T) {

	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb1 := cryptoo.PublicKey(pv1.GetPublicKey())
	pb2 := cryptoo.PublicKey(pv2.GetPublicKey())

	t.Log("pv1", pv1)
	t.Log("pb1", pb1)
	t.Log("pv2", pv2)
	t.Log("pb2", pb2)

	originalBlock := core.NewBlock(types.RandomHash(), 1, []*core.Transaction{
		core.NewTransaction(pv1, pb2, []byte("transaction1"), 100, 0),
	})

	var buf bytes.Buffer
	if err := originalBlock.Encode(&buf); err != nil {
		t.Fatalf("编码区块失败：%v", err)
	}

	decodedBlock := &core.Block{}
	if err := decodedBlock.Decode(&buf); err != nil {
		t.Fatalf("解码区块失败：%v", err)
	}

	if originalBlock.GetDataHash() != decodedBlock.GetDataHash() {
		t.Errorf("解码后的区块哈希不匹配")
	}
	if originalBlock.GetPrevBlockHash() != decodedBlock.GetPrevBlockHash() {
		t.Errorf("解码后的前一个区块哈希不匹配")
	}
	if originalBlock.Height() != decodedBlock.Height() {
		t.Errorf("解码后的区块高度不匹配")
	}
	if !(originalBlock.Timestamp() == decodedBlock.Timestamp()) {
		t.Errorf("解码后的时间戳不匹配")
	}
	if len(originalBlock.Transactions) != len(decodedBlock.Transactions) {
		t.Errorf("解码后的交易数量不匹配")
	}
}

func TestBlockHash(t *testing.T) {
	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb1 := cryptoo.PublicKey(pv1.GetPublicKey())
	pb2 := cryptoo.PublicKey(pv2.GetPublicKey())

	t.Log("pv1", pv1)
	t.Log("pb1", pb1)
	t.Log("pv2", pv2)
	t.Log("pb2", pb2)
	block := core.NewBlock(types.RandomHash(), 1, []*core.Transaction{
		core.NewTransaction(pv1, pb2, []byte("transaction1"), 100, 0),
	})

	hash := block.GetDataHash()
	if hash.IsZero() {
		t.Error("区块哈希不应为空")
	}

	// 修改区块内容，确保哈希值变化 只能修改交易
	block.Transactions[0].Data = []byte("transaction11")

	newHash, err := block.CalculateDataHash()
	if err != nil {
		t.Errorf("计算区块哈希失败：%v", err)
	}
	if hash == newHash {
		t.Error("修改区块内容后，哈希值应该改变")
	}
}

func TestBlockIntegration(t *testing.T) {
	// 创建一系列区块，模拟一个小型区块链
	var blocks []*core.Block
	prevHash := types.Hash{}

	pv1, _ := cryptoo.GeneratePrivateKey()
	pv2, _ := cryptoo.GeneratePrivateKey()
	pb1 := cryptoo.PublicKey(pv1.GetPublicKey())
	pb2 := cryptoo.PublicKey(pv2.GetPublicKey())

	t.Log("pv1", pv1)
	t.Log("pb1", pb1)
	t.Log("pv2", pv2)
	t.Log("pb2", pb2)

	for i := uint32(0); i < 5; i++ {
		transactions := []*core.Transaction{
			core.NewTransaction(pv1, pb2, []byte("transaction1"), uint64(100*(i+1)), 0),
			core.NewTransaction(pv2, pb1, []byte("transaction2"), uint64(200*(i+1)), 0),
		}
		block := core.NewBlock(prevHash, i, transactions)
		blocks = append(blocks, block)
		prevHash = block.GetDataHash()

		// 等待一小段时间，确保时间戳不同
		time.Sleep(10 * time.Millisecond)
	}

	// 验证区块链的完整性
	for i := 1; i < len(blocks); i++ {
		if blocks[i].GetPrevBlockHash() != blocks[i-1].GetDataHash() {
			t.Errorf("区块 %d 的前一个区块哈希不正确", i)
		}
		if blocks[i].Height() != blocks[i-1].Height() + 1 {
			t.Errorf("区块 %d 的高度不正确", i)
		}
		if blocks[i].Timestamp() < blocks[i-1].Timestamp() {
			t.Errorf("区块 %d 的时间戳不正确", i)
		}
	}

	// 测试序列化和反序列化整个区块链
	var buf bytes.Buffer
	for _, block := range blocks {
		if err := block.Encode(&buf); err != nil {
			t.Fatalf("编码区块失败：%v", err)
		}
	}

	decodedBlocks := make([]*core.Block, 0)
	for i := 0; i < len(blocks); i++ {
		decodedBlock := &core.Block{}
		if err := decodedBlock.Decode(&buf); err != nil {
			t.Fatalf("解码区块失败：%v", err)
		}
		decodedBlocks = append(decodedBlocks, decodedBlock)
	}

	// 验证解码后的区块链
	for i := 0; i < len(blocks); i++ {
		if blocks[i].GetDataHash() != decodedBlocks[i].GetDataHash() {
			t.Errorf("解码后的区块 %d 哈希不匹配", i)
		}
	}
}
