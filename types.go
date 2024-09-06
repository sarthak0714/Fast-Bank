package main

import (
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Id        int       `json:"id"`
	Fname     string    `json:"fname"`
	Lname     string    `json:"lanme"`
	EPassword string    `json:"epassword"`
	AcNumber  int64     `json:"ac_number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAccountReq struct {
	Fname    string `json:"fname"`
	Lname    string `json:"lname"`
	Password string `json:"password"`
}

type TransferReq struct {
	ToAccount int   `json:"to_account"`
	Amount    int64 `json:"amount"`
}

type JWTClaims struct {
	Id int `json:"id"`
	jwt.RegisteredClaims
}

type TransferMessage struct {
	TransferId string `json:"transfer_id"`
	SenderId   int    `json:"sender_id"`
	ToAccount  int    `json:"to_account"`
	Amount     int64  `json:"amount"`
	Status     string `josn:"status"`
}

func NewAccount(fName, lName, password string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		Fname:     fName,
		Lname:     lName,
		EPassword: string(encpw),
		AcNumber:  int64(rand.Intn(1000000)),
		Balance:   1000,
		CreatedAt: time.Now().UTC(),
	}, nil

}
