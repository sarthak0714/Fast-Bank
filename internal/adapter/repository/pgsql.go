package repository

import (
	"github.com/sarthak014/Fast-Bank/internal/core/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGStore struct {
	db *gorm.DB
}

func NewPGStore() (*PGStore, error) {
	dsn := "host=localhost user=postgres dbname=postgres password=jomum port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PGStore{
		db: db,
	}, nil
}

func (s *PGStore) Init() error {
	err := s.db.AutoMigrate(&domain.Account{})
	if err != nil {
		return err
	}
	return s.db.AutoMigrate(&domain.TransferMessage{})

}
func (s *PGStore) CreateAccount(acc *domain.Account) error {
	return s.db.Create(acc).Error
}

func (s *PGStore) DeleteAccount(id int) error {
	return s.db.Delete(&domain.Account{}, id).Error
}

func (s *PGStore) UpdateAccount(acc *domain.Account) error {
	return s.db.Save(acc).Error
}

func (s *PGStore) UpdateBalance(id int, balance int64) error {
	return s.db.Model(&domain.Account{}).Where("id = ?", id).Update("balance", balance).Error
}

func (s *PGStore) GetAccountById(id int) (*domain.Account, error) {
	var acc domain.Account
	err := s.db.First(&acc, id).Error
	return &acc, err
}

func (s *PGStore) GetAccounts() ([]*domain.Account, error) {
	var accounts []*domain.Account
	err := s.db.Find(&accounts).Error
	return accounts, err
}

func (s *PGStore) GetTransferStatus(trxid string) (string, error) {
	var trx domain.TransferMessage
	err := s.db.Where("transfer_id = ?", trxid).First(&trx).Error
	if err != nil {
		return "", err
	}
	return trx.Status, nil
}

func (s *PGStore) AddTransfer(transferMsg *domain.TransferMessage) error {
	return s.db.Create(transferMsg).Error
}

func (s *PGStore) UpdateTransferStatus(trxid, status string) error {
	return s.db.Model(&domain.TransferMessage{}).Where("transfer_id = ?", trxid).Update("status", status).Error
}
