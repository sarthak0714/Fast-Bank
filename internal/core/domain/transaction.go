package domain

import "time"

type TransferReq struct {
	Amount int64 `json:"amount"`
}

type TransferMessage struct {
	TransferId string    `json:"transfer_id" gorm:"type:varchar(100);primaryKey"`
	SenderId   int       `json:"sender_id" gorm:"type:int;not null"`
	ToAccount  int       `json:"to_account" gorm:"type:int;not null"`
	Amount     int64     `json:"amount" gorm:"type:bigint;not null"`
	Status     string    `json:"status" gorm:"type:varchar(20);not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"type:timestamp;not null;default:current_timestamp;autoUpdateTime"`
}
