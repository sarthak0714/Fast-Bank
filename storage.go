package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(int) error
	GetAccountById(int) (*Account, error)
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
	q := `create table if not exists account(
		id serial primary key,
		fname varchar(50),
		lname varchar(50),
		ac_number serial,
		balance serial,
		created_at timestamp
	)`
	_, err := s.db.Exec(q)
	return err
}

func (s *PGStore) CreateAccount(acc *Account) error {
	q := `insert into account(fname,lname,ac_number,balance,created_at) values($1,$2,$3,$4,$5)`
	res, err := s.db.Query(q, acc.Fname, acc.Lname, acc.AcNumber, acc.Balance, acc.CreatedAt)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", res)
	fmt.Println("called")

	return nil
}

func (s *PGStore) DeleteAccount(id int) error {
	q := `delete from account where id = ?`
	res, err := s.db.Query(q, id)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", res)
	return nil
}

func (s *PGStore) UpdateAccount(id int) error {
	return nil
}

func (s *PGStore) GetAccountById(id int) (*Account, error) {
	q := `select * from account where id =?`
	res, err := s.db.Query(q, id)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", res)
	acc := &Account{}
	for res.Next() {
		er := res.Scan(&acc.Id, &acc.Fname, &acc.Lname, &acc.AcNumber, &acc.Balance, &acc.CreatedAt)
		if er != nil {
			return nil, er
		}
	}
	return acc, nil
}
