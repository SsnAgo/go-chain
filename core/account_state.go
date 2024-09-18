package core

import (
	"errors"
	"go-chain/types"
	"sync"
)

var (
	AccountNotExistsErr = errors.New("no such account")
	NotZeroAddrErr      = errors.New("not zero address")
	InsufficientBalance = errors.New("insufficient balance")
)

type Account struct {
	Address types.Address
	Balance uint64
}

type AccountState struct {
	mu       sync.RWMutex
	accounts map[types.Address]*Account
}

func NewAccountState() *AccountState {
	return &AccountState{
		mu:       sync.RWMutex{},
		accounts: make(map[types.Address]*Account),
	}
}

func (s *AccountState) GetAccount(address types.Address) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accounts[address]
}

func (s *AccountState) CreateAccount(address types.Address, account *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.accounts[address] == nil {
		s.accounts[address] = account
	}
}

func (s *AccountState) GetBalance(address types.Address) (balance uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	account := s.accounts[address]
	if account == nil {
		return 0
	}
	return account.Balance
}

func (s *AccountState) Transfer(from types.Address, to types.Address, amount uint64) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if from.IsZero() {
		return NotZeroAddrErr
	}
	fromAccount := s.accounts[from]
	if fromAccount == nil {
		return AccountNotExistsErr
	}
	if fromAccount.Balance < amount {
		return InsufficientBalance
	}
	if s.accounts[to] == nil {
		s.accounts[to] = &Account{
			Address: to,
		}
	}
	toAccount := s.accounts[to]
	toAccount.Balance += amount
	fromAccount.Balance -= amount
	return nil

}
