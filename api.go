package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	//route to /account with handleAccount parent function as the controller
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	//route to /account/{id} with handleGetAccountbyID parent function as the controller
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))

	//route to /transfer with handleTransferAccount function as the controller
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransferAccount))

	//route to /update with handleUpdateAccount function as the controller
	router.HandleFunc("/update", makeHTTPHandleFunc(s.handleUpdateAccount))

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	//fetch the account from db
	acc, err := s.store.GetAccountByNumber(int(req.Number))
	if err != nil {
		return err
	}

	//if its an incorrect password return error
	if !acc.ValidatePassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	//send token as a response
	token, err := createJWT(acc)
	if err != nil {
		return err
	}
	resp := LoginResponse{
		Token:  token,
		Number: acc.Number,
	}

	return WriteJSON(w, http.StatusOK, resp)

}

// parent function for creating an account or pulling an account from the server
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.HandleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.HandleCreateAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// parent function for pulling a specific account from the server or deleting it
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		//get the id of the account from db return error if it doesnt exist
		id, err := getID(r)
		if err != nil {
			return err
		}
		//get specific account from db usimg id
		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, account)
	}

	if r.Method == "DELETE" {
		//handle delete account
		return s.HandleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// function for updating user profile
func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	//initialize a new account to be stored in memory
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

// transfer from account to account
func (s *APIServer) handleTransferAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	req := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	defer r.Body.Close()

	//find account to be sent money by number
	acc, err := s.store.GetAccountByNumber(int(req.ToAccount))
	if err != nil {
		return err
	}
	//update the amount in the account
	updatedAcc, err := s.store.Transfer(acc, int64(req.Amount))
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, updatedAcc)

	// defer r.Body.Close()
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

// jwt middleware
func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}

		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}

		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %V", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// covert id to an int and return
func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	// convert int to string
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
