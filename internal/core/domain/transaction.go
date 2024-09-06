package domain

type TransferReq struct {
	ToAccount int   `json:"to_account"`
	Amount    int64 `json:"amount"`
}

type TransferMessage struct {
	TransferId string `json:"transfer_id"`
	SenderId   int    `json:"sender_id"`
	ToAccount  int    `json:"to_account"`
	Amount     int64  `json:"amount"`
	Status     string `josn:"status"`
}
