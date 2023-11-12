package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) (*Account, error)
	TransferToAccount(int, int, int) (*Account, error)
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) UpdateAccount(acc *Account) (*Account, error) {
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
	return nil, err
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

func NewAccount(email, password string) (*Account, error) {
	encrptp, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		Email:             email,
		EncryptedPassword: string(encrptp),
		Number:            int64(rand.Intn(1000000000)),
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func (s *PostgresStore) TransferToAccount(fromAcc, toAcc, amount int) (*Account, error) {

	//subtract the amount from the user account
	q := `UPDATE account SET balance = balance - $1 WHERE number = $2`
	_, err := s.db.Exec(q, amount, fromAcc)
	if err != nil {
		return nil, err
	}

	query := `UPDATE account SET balance = balance + $1 WHERE number = $2`
	_, err = s.db.Exec(query, amount, toAcc)
	if err != nil {
		return nil, err
	}
	//should probably find a way to write 2 queries in a single statement

	//return the updated user balance
	updatedAccount, err := s.GetAccountByNumber(int(toAcc))
	if err != nil {
		return nil, err
	}

	return updatedAccount, nil
}

// func (s *PostgresStore) Withdraw(acc *Account, amount int64) (*Account, error) {
// 	query := `UPDATE account SET balance = $1 WHERE id = $2`
// 	newAmount := acc.Balance - amount
// 	_, err := s.db.Exec(query, newAmount, acc.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	updatedAccount, err := s.GetAccountByID(acc.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return updatedAccount, nil
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
	return nil, fmt.Errorf("account with number [%d] not found", number)

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
