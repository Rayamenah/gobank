package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// create a new account
func (s *APIServer) HandleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateAccountRequest)
	//return error if the body doesnt exist
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	account, err := NewAccount(req.Email, req.Password)
	if err != nil {
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

// get all accounts in db
func (s *APIServer) HandleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

// delete account from db
func (s *APIServer) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

// transfer from account to account
func (s *APIServer) HandleTransferAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	req := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	defer r.Body.Close()
	//find the user account
	userAcc, err := s.store.GetAccountByNumber(int(req.FromAcccount))
	if err != nil {
		return err
	}
	//check if the amount to be transferred is more than what is avaliable in user account
	if userAcc.Balance < int64(req.Amount) {
		return err
	}

	//update the amount in the account
	updatedAcc, err := s.store.TransferToAccount(req.FromAcccount, req.ToAccount, req.Amount)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, updatedAcc)
}

// function for updating user profile
func (s *APIServer) HandleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	//initialize a new account with zero values to be stored in memory
	req := new(Account)
	//decode the request
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	// if there are no errors pass the req as a parameter for the UpdateAccount() fxn
	updateAcc, err := s.store.UpdateAccount(req)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, updateAcc)
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
