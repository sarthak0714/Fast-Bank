package main

import "math/rand"

type Account struct {
	Id       int    `json:"id"`
	Fname    string `json:"fname"`
	Lname    string `json:"lanme"`
	AcNumber int64  `json:"acnumber"`
	Balance  int64  `json:"balance"`
}

func NewAccount(fName, lName string) *Account {
	return &Account{
		Id:       rand.Intn(10000),
		Fname:    fName,
		Lname:    lName,
		AcNumber: rand.Int63(),
	}

}
