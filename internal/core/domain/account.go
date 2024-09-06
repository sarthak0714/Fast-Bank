package domain

import "time"

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
