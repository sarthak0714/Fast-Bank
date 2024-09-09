package domain

import "time"

type Account struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Fname     string    `json:"fname" gorm:"type:varchar(100);not null"`
	Lname     string    `json:"lname" gorm:"type:varchar(100);not null"`
	EPassword string    `json:"epassword" gorm:"type:varchar(255);not null"`
	AcNumber  int32     `json:"ac_number" gorm:"unique;not null"`
	Balance   int64     `json:"balance" gorm:"not null;default:1000"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:current_timestamp"`
}

type CreateAccountReq struct {
	Fname    string `json:"fname"`
	Lname    string `json:"lname"`
	Password string `json:"password"`
}
