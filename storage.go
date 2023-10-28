package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
	UpdateAccount(*Account) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("error loading env file %v", err)
	}
	connStr := os.Getenv("POSTGRES_URL")
	// connStr := ""
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table if not exists account(
		id serial primary key,
		email varchar(100),
		username varchar(100),
		number serial,
		encrypted_password varchar(100),
		balance serial,
		created_at timestamp 
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `INSERT into account (first_name, last_name, number, encrypted_password, balance, created_at) values($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Query(
		query,
		acc.Email,
		acc.Username,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) UpdateAcount(acc *Account) (*Account, error) {
	query := `UPDATE account SET id, email, username, number, encrypted_password, balance, created_at = $1, $2, $3, $4, $5, $6, $7 where id == $1`

	rows, err := s.db.Query(
		query,
		acc.ID,
		acc.Email,
		acc.Username,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("couldn't update account", acc)
}

// func (s *PostgresStore) updateUsername(username string) error {
// 	query := `UPDATE account SET encrypted_password = $1, where id == $2`

// 	_, err := s.db.Query(
// 		query, username,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (s *PostgresStore) DeleteAccount(id int) error {

	//probably shouldnt "truly" delete an account in production
	_, err := s.db.Query("delete from account where id = $1", id)
	return err
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("accont with number [%d] not found", number)

}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

// controller function for getting accounts from db
func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.Email,
		&account.Username,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)
	return account, err
}
