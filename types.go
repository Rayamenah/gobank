package main

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}

// createaccontrequest so we dont use account struct to send confidential details
type CreateAccountRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type Account struct {
	ID                int       `json:"id"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"createdAt"`
}

func (a *Account) ValidatePassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(pw)) == nil
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

func (a *Account) ChangePassword(password, newPassword string) (string, error) {
	//check if old password is correct
	if !a.ValidatePassword(password) {
		return "", fmt.Errorf("not authenticated")
	}

	encrptp, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	password = string(encrptp)
	return password, err
}

func (a *Account) validateEmail(email string) (*Account, error) {
	panic("unimplemented")
}
