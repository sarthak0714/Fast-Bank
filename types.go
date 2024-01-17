package main

import (
	"math/rand"
	"time"
)

type Account struct {
	Id        int       `json:"id"`
	Fname     string    `json:"fname"`
	Lname     string    `json:"lanme"`
	AcNumber  int64     `json:"ac_number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAccountReq struct {
	Fname string `json:"fname"`
	Lname string `json:"lname"`
}

type TransferReq struct {
	ToAccount int `json:"to_account"`
	Amount    int `json:"amount"`
}

func NewAccount(fName, lName string) *Account {
	return &Account{
		Fname:     fName,
		Lname:     lName,
		AcNumber:  int64(rand.Intn(1000000)),
		CreatedAt: time.Now().UTC(),
	}

}
