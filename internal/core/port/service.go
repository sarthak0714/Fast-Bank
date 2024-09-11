package port

import (
	"github.com/labstack/echo/v4"
	"github.com/sarthak014/Fast-Bank/internal/core/domain"
)

type AccountService interface {
	Create(*domain.CreateAccountReq) (*domain.Account, error)
	Delete(string) error
	// Update(*domain.Account) error
	GetAll() ([]*domain.Account, error)
	GetById(string) (*domain.Account, error)
	GetByAccNo(int) (*domain.Account, error)
}

type TransactionService interface {
	PublishTransferMessage(domain.TransferMessage) error
	GetTransferStatus(string) (string, error)
	ExecuteTransfer(domain.TransferMessage) error
	AddTransferRecord(*domain.TransferMessage) error
	GetByAccNo(int) ([]*domain.TransferMessage, error)
	ProcessTransfers()
}

type AuthService interface {
	Validate(string) (*domain.JWTClaims, error)
	Middleware(echo.HandlerFunc) echo.HandlerFunc
	Generate(int) (string, error)
}
