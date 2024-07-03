package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	UpdateBalance(int, int64) error
}

type PGStore struct {
	db *sql.DB
}

func NewPGStore() (*PGStore, error) {
	cstr := "user=postgres dbname=postgres password=jomum sslmode=disable"
	db, err := sql.Open("postgres", cstr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PGStore{
		db: db,
	}, nil
}

func (s *PGStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PGStore) CreateAccountTable() error {
	q := `CREATE TABLE IF NOT EXISTS account (
        id SERIAL PRIMARY KEY,
        fname VARCHAR(50),
        lname VARCHAR(50),
        ac_number SERIAL,
        password VARCHAR(256),
        balance BIGINT,
        created_at TIMESTAMP
    )`
	_, err := s.db.Exec(q)
	return err
}

func (s *PGStore) CreateAccount(acc *Account) error {
	q := `INSERT INTO account (fname, lname, ac_number, password, balance, created_at) 
          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := s.db.QueryRow(q, acc.Fname, acc.Lname, acc.AcNumber, acc.EPassword, acc.Balance, acc.CreatedAt).Scan(&acc.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *PGStore) DeleteAccount(id int) error {
	q := "delete from account where id = $1"
	_, err := s.db.Query(q, id)
	return err
}

func (s *PGStore) UpdateAccount(acc *Account) error {
	q := `UPDATE account SET fname=$1, lname=$2, epassword=$3, ac_number=$4, balance=$5, created_at=$6 WHERE id=$7`
	_, err := s.db.Exec(q, acc.Fname, acc.Lname, acc.EPassword, acc.AcNumber, acc.Balance, acc.CreatedAt, acc.Id)
	if err != nil {
		return err
	}
	return nil

}

func (s *PGStore) UpdateBalance(id int, balance int64) error {
	q := `UPDATE account Set balance=$1 where id = $2`
	_, err := s.db.Exec(q, balance, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PGStore) GetAccountById(id int) (*Account, error) {
	q := "SELECT id, fname, lname, ac_number, password, balance, created_at FROM account WHERE id=$1"
	acc := new(Account)
	err := s.db.QueryRow(q, id).Scan(&acc.Id, &acc.Fname, &acc.Lname, &acc.AcNumber, &acc.EPassword, &acc.Balance, &acc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PGStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT id, fname, lname, ac_number, password, balance, created_at FROM account")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		acc, err := s.ScanAccounts(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *PGStore) ScanAccounts(rows *sql.Rows) (*Account, error) {
	acc := new(Account)
	err := rows.Scan(&acc.Id, &acc.Fname, &acc.Lname, &acc.AcNumber, &acc.EPassword, &acc.Balance, &acc.CreatedAt)
	return acc, err
}
