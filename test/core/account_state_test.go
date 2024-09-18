package test

import (
	"go-chain/core"
	"go-chain/types"
	"testing"
)

func TestNewAccountState(t *testing.T) {
	as := core.NewAccountState()
	if as == nil {
		t.Fatal("NewAccountState应该返回一个非空的AccountState")
	}
}

func TestGetAccount(t *testing.T) {
	as := core.NewAccountState()
	addr := types.Address{0x1}
	account := &core.Account{Address: addr, Balance: 100}
	as.CreateAccount(addr, account)

	retrievedAccount := as.GetAccount(addr)
	if retrievedAccount != account {
		t.Errorf("GetAccount返回的账户与创建的账户不匹配")
	}

	nonExistentAddr := types.Address{0x2}
	if as.GetAccount(nonExistentAddr) != nil {
		t.Errorf("对于不存在的地址，GetAccount应该返回nil")
	}
}

func TestCreateAccount(t *testing.T) {
	as := core.NewAccountState()
	addr := types.Address{0x1}
	account := &core.Account{Address: addr, Balance: 100}

	as.CreateAccount(addr, account)
	if as.GetAccount(addr) != account {
		t.Errorf("CreateAccount未能正确创建账户")
	}

	// 测试重复创建
	as.CreateAccount(addr, &core.Account{Address: addr, Balance: 200})
	if as.GetAccount(addr).Balance != 100 {
		t.Errorf("CreateAccount不应覆盖已存在的账户")
	}
}

func TestGetBalance(t *testing.T) {
	as := core.NewAccountState()
	addr := types.Address{0x1}
	account := &core.Account{Address: addr, Balance: 100}
	as.CreateAccount(addr, account)

	balance := as.GetBalance(addr)
	if balance != 100 {
		t.Errorf("GetBalance返回的余额不正确，期望100，实际%d", balance)
	}

	nonExistentAddr := types.Address{0x2}
	if as.GetBalance(nonExistentAddr) != 0 {
		t.Errorf("对于不存在的地址，GetBalance应该返回0")
	}
}

func TestTransfer(t *testing.T) {
	as := core.NewAccountState()
	from := types.Address{0x1}
	to := types.Address{0x2}
	as.CreateAccount(from, &core.Account{Address: from, Balance: 100})
	as.CreateAccount(to, &core.Account{Address: to, Balance: 50})

	// 正常转账
	err := as.Transfer(from, to, 30)
	if err != nil {
		t.Errorf("正常转账应该成功：%v", err)
	}
	if as.GetBalance(from) != 70 || as.GetBalance(to) != 80 {
		t.Errorf("转账后余额不正确")
	}

	// 余额不足
	err = as.Transfer(from, to, 100)
	if err != core.InsufficientBalance {
		t.Errorf("余额不足时应返回InsufficientBalance错误")
	}

	// 从不存在的账户转账
	nonExistentAddr := types.Address{0x3}
	err = as.Transfer(nonExistentAddr, to, 10)
	if err != core.AccountNotExistsErr {
		t.Errorf("从不存在的账户转账应返回AccountNotExistsErr错误")
	}

	// 零地址转账
	err = as.Transfer(types.Address{}, to, 10)
	if err != core.NotZeroAddrErr {
		t.Errorf("从零地址转账应返回NotZeroAddrErr错误")
	}
}

func TestAccountStateIntegration(t *testing.T) {
	as := core.NewAccountState()
	addr1 := types.Address{0x1}
	addr2 := types.Address{0x2}
	addr3 := types.Address{0x3}

	// 创建账户
	as.CreateAccount(addr1, &core.Account{Address: addr1, Balance: 1000})
	as.CreateAccount(addr2, &core.Account{Address: addr2, Balance: 500})

	// 执行一系列转账操作
	if err := as.Transfer(addr1, addr2, 300); err != nil {
		t.Fatalf("转账失败：%v", err)
	}
	if err := as.Transfer(addr2, addr3, 100); err != nil {
		t.Fatalf("转账失败：%v", err)
	}
	if err := as.Transfer(addr1, addr3, 200); err != nil {
		t.Fatalf("转账失败：%v", err)
	}

	// 验证最终余额
	if as.GetBalance(addr1) != 500 {
		t.Errorf("addr1最终余额不正确，期望500，实际%d", as.GetBalance(addr1))
	}
	if as.GetBalance(addr2) != 700 {
		t.Errorf("addr2最终余额不正确，期望700，实际%d", as.GetBalance(addr2))
	}
	if as.GetBalance(addr3) != 300 {
		t.Errorf("addr3最终余额不正确，期望300，实际%d", as.GetBalance(addr3))
	}
}
