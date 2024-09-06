package main

import (
	"fmt"
	"strconv"
)

type AccountService struct {
	store Storage
}

func NewAccountService(store Storage) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (s *AccountService) GetAccounts() ([]*Account, error) {
	return s.store.GetAccounts()
}

func (s *AccountService) GetAccountById(id string) (*Account, error) {
	accountId, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %v", err)
	}
	return s.store.GetAccountById(accountId)
}

func (s *AccountService) CreateAccount(req *CreateAccountReq) (*Account, error) {
	acc, err := NewAccount(req.Fname, req.Lname, req.Password)
	if err != nil {
		return nil, err
	}
	if err := s.store.CreateAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *AccountService) DeleteAccount(id string) error {
	accountId, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid account ID: %v", err)
	}
	return s.store.DeleteAccount(accountId)
}
