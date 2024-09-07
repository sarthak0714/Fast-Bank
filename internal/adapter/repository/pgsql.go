package repository

import (
	"fmt"
	"time"

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
	err := s.db.AutoMigrate(&domain.Account{}, &domain.TransferMessage{})
	if err != nil {
		return err
	}
	return nil
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
	return s.db.Model(&domain.TransferMessage{}).Where("transfer_id = ?", trxid).Updates(map[string]interface{}{"status": status, "updated_at": time.Now().UTC()}).Error
}

func (s *PGStore) Transcation(senderAccount, recipientAccount *domain.Account, msg *domain.TransferMessage) error {
	tx := s.db.Begin()

	senderNewBalance := senderAccount.Balance - msg.Amount
	err := tx.Model(&domain.Account{}).Where("id = ?", msg.SenderId).Update("balance", senderNewBalance).Error //s.UpdateBalance(msg.SenderId, senderNewBalance)
	if err != nil {
		tx.Rollback()
		er := s.UpdateTransferStatus(msg.TransferId, "failed")
		if er != nil {
			return er
		}
		return fmt.Errorf("failed to update sender account: %v", err)
	}

	recipientNewBalance := recipientAccount.Balance + msg.Amount
	err = tx.Model(&domain.Account{}).Where("id = ?", recipientAccount.Id).Update("balance", recipientNewBalance).Error //.UpdateBalance(msg.ToAccount, recipientNewBalance)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update recipient account: %v", err)
	}
	tx.Commit()
	return nil
}
