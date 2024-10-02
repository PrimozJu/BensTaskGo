package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Storage interface {
	// Create a new account
	CreateAccount(*Account) error
	// Get an account by id
	DeleteAccount(int64) error
	// Get an account by id
	UpdateAccount(int64, *Account) error
	// Get an account by id
	GetAccountByID(int64) (*Account, error)
	// Get all accounts
	GetAccounts() ([]*Account, error)
}

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage() (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", "accounts.db")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Init() error {
	err := s.createAccountTable()
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) createAccountTable() error {
	_, err := s.db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, first_name TEXT, last_name TEXT, number INTEGER, balance INTEGER, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) CreateAccount(a *Account) error {
	fmt.Printf("Creating account: %+v\n", a)
	resp, err := s.db.Exec("INSERT INTO accounts (first_name, last_name, number, balance) VALUES (?, ?, ?, ?)", a.FirstName, a.LastName, a.Number, a.Balance)
	fmt.Printf("Response: %+v\n", resp)
	return err
}

func (s *SQLiteStorage) DeleteAccount(id int64) error {
	fmt.Printf("Deleting account: %d\n", id)
	_, err := s.db.Exec("DELETE FROM accounts WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) UpdateAccount(id int64, a *Account) error {
	_, err := s.db.Exec("UPDATE accounts SET first_name = ?, last_name = ?, number = ?, balance = ? WHERE id = ?", a.FirstName, a.LastName, a.Number, a.Balance, id)
	return err
}

func (s *SQLiteStorage) GetAccountByID(id int64) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("Account %d not found", id)
}

func (s *SQLiteStorage) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		a, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	a := new(Account)
	err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Number, &a.Balance, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}
