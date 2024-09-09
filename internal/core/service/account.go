package service

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/sarthak014/Fast-Bank/internal/core/domain"
	"github.com/sarthak014/Fast-Bank/internal/core/port"
	"golang.org/x/crypto/bcrypt"
)

type accountService struct {
	store port.StorageService
}

func NewAccountService(store port.StorageService) port.AccountService {
	return &accountService{
		store: store,
	}
}

func (s *accountService) GetAll() ([]*domain.Account, error) {
	return s.store.GetAccounts()
}

func (s *accountService) GetById(id string) (*domain.Account, error) {
	accountId, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %v", err)
	}
	return s.store.GetAccountById(accountId)
}

func (s *accountService) GetByAccNo(accNo int) (*domain.Account, error) {
	return s.store.GetAccountByAccNo(accNo)
}

func (s *accountService) Create(req *domain.CreateAccountReq) (*domain.Account, error) {
	acc, err := NewAccount(req.Fname, req.Lname, req.Password)
	if err != nil {
		return nil, err
	}
	if err := s.store.CreateAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *accountService) Delete(id string) error {
	accountId, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid account ID: %v", err)
	}
	return s.store.DeleteAccount(accountId)
}

func NewAccount(fName, lName, password string) (*domain.Account, error) {
	if fName == "" {
		return nil, fmt.Errorf("first name is required")
	}
	if lName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &domain.Account{
		Fname:     fName,
		Lname:     lName,
		EPassword: string(encpw),
		AcNumber:  rand.Int31n(100000),
		Balance:   1000,
		CreatedAt: time.Now().UTC(),
	}, nil

}
