package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	UpdateBalance(int, int64) error
	AddTransfer(*TransferMessage) error
	GetTransferStatus(string) (string, error)
	UpdateTransferStatus(string, string) error
}

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
	err := s.db.AutoMigrate(&Account{})
	if err != nil {
		return err
	}
	return s.db.AutoMigrate(&TransferMessage{})

}

func (s *PGStore) CreateAccount(acc *Account) error {
	return s.db.Create(acc).Error
}

func (s *PGStore) DeleteAccount(id int) error {
	return s.db.Delete(&Account{}, id).Error
}

func (s *PGStore) UpdateAccount(acc *Account) error {
	return s.db.Save(acc).Error
}

func (s *PGStore) UpdateBalance(id int, balance int64) error {
	return s.db.Model(&Account{}).Where("id = ?", id).Update("balance", balance).Error
}

func (s *PGStore) GetAccountById(id int) (*Account, error) {
	var acc Account
	err := s.db.First(&acc, id).Error
	return &acc, err
}

func (s *PGStore) GetAccounts() ([]*Account, error) {
	var accounts []*Account
	err := s.db.Find(&accounts).Error
	return accounts, err
}

func (s *PGStore) GetTransferStatus(trxid string) (string, error) {
	var trx TransferMessage
	err := s.db.Where("transfer_id = ?", trxid).First(&trx).Error
	if err != nil {
		return "", err
	}
	return trx.Status, nil
}

func (s *PGStore) AddTransfer(transferMsg *TransferMessage) error {
	return s.db.Create(transferMsg).Error
}

func (s *PGStore) UpdateTransferStatus(trxid, status string) error {
	return s.db.Model(&TransferMessage{}).Where("transfer_id = ?", trxid).Update("status", status).Error
}
