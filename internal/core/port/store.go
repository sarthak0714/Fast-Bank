package port

import "github.com/sarthak014/Fast-Bank/internal/core/domain"

type StorageService interface {
	CreateAccount(*domain.Account) error
	DeleteAccount(int) error
	UpdateAccount(*domain.Account) error
	GetAccounts() ([]*domain.Account, error)
	GetAccountById(int) (*domain.Account, error)
	UpdateBalance(int, int64) error
	AddTransfer(*domain.TransferMessage) error
	GetTransferStatus(string) (string, error)
	UpdateTransferStatus(string, string) error
}

type MQService struct {
	
}
