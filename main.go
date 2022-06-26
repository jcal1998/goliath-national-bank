package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type AccountAccess struct {
	AccountID int32 `json:"account_id"`
	Password  int32 `json:"password"`
}

type PersonalInfo struct {
	Name string `json:"name"`
	Cpf  int32  `json:"cpf"`
}

type Account struct {
	AccountAccess
	PersonalInfo
	CurrentBalance float32 `json:"current_balance"`
}

type RequestResponse struct {
	Status         string
	CurrentBalance float32
}

type CreateAccountResponse struct {
	RequestResponse
	PersonalInfo
	AccountID int32
}

type AccountChangeValueRequest struct {
	AccountAccess
	Value float32 `json:"value"`
}
type Destiny struct {
	AccountChangeValueRequest
	DestinyID int32 `json:"destiny_id"`
}

var accounts []Account

var leeAccount = Account{
	AccountAccess: AccountAccess{
		AccountID: 130230,
		Password:  12345,
	},
	PersonalInfo: PersonalInfo{
		Name: "Test User",
		Cpf:  1234567890,
	},
	CurrentBalance: 2500,
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var requestBody Account

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.Name == "" || requestBody.Cpf == 0 || requestBody.Password == 0 {
		http.Error(w, "Invalid Name, Cpf or Password", http.StatusBadRequest)
		return
	}

	for _, currentAccount := range accounts {
		if currentAccount.Cpf == requestBody.Cpf {
			http.Error(w, "Failed to create a user: User already exists, cpf in use", http.StatusBadRequest)
			return
		}
	}

	rand.Seed(time.Now().UnixNano())
	min := 100000
	max := 999999

	newAccount := Account{
		AccountAccess: AccountAccess{
			AccountID: int32(rand.Intn(max-min) + min),
			Password:  requestBody.Password,
		},
		PersonalInfo: PersonalInfo{
			Name: requestBody.Name,
			Cpf:  requestBody.Cpf,
		},
		CurrentBalance: 0,
	}

	accounts = append(accounts, newAccount)

	response := CreateAccountResponse{
		RequestResponse: RequestResponse{
			Status:         "Create account operation executed with success",
			CurrentBalance: newAccount.CurrentBalance,
		},
		AccountID: newAccount.AccountID,
		PersonalInfo: PersonalInfo{
			Name: newAccount.Name,
			Cpf:  newAccount.Cpf,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func validateCaller(account int32, password int32) (*Account, string) {
	for index, currentAccount := range accounts {
		if currentAccount.AccountID == account && currentAccount.Password == password {
			return &accounts[index], ""
		}
	}
	return &Account{}, "invalid caller or password"
}

func accountGet(w http.ResponseWriter, r *http.Request) {
	var requestBody Account

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	caller, error := validateCaller(requestBody.AccountID, requestBody.Password)
	if error != "" {
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	response := CreateAccountResponse{
		RequestResponse: RequestResponse{
			Status:         "Get account operation executed with success",
			CurrentBalance: caller.CurrentBalance,
		},
		AccountID: caller.AccountID,
		PersonalInfo: PersonalInfo{
			Name: caller.Name,
			Cpf:  caller.Cpf,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func changeAmount(account *Account, value float32, operation string) string {
	if value <= 0 {
		return "Value not valid"
	}

	switch operation {
	case "deposit":
		account.CurrentBalance += value
	case "withdraw":
		if value > account.CurrentBalance {
			return "Not enough money"
		}
		account.CurrentBalance -= value
	}
	return ""
}

func depositPost(w http.ResponseWriter, r *http.Request) {
	var requestBody AccountChangeValueRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	caller, error := validateCaller(requestBody.AccountID, requestBody.Password)
	if error != "" {
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	if operationError := changeAmount(caller, requestBody.Value, "deposit"); operationError != "" {
		http.Error(w, operationError, http.StatusUnprocessableEntity)
		return
	}

	response := RequestResponse{
		Status:         "Deposit operation executed with success",
		CurrentBalance: caller.CurrentBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func withdrawPost(w http.ResponseWriter, r *http.Request) {
	var requestBody AccountChangeValueRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	caller, error := validateCaller(requestBody.AccountID, requestBody.Password)
	if error != "" {
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	if operationError := changeAmount(caller, requestBody.Value, "withdraw"); operationError != "" {
		http.Error(w, operationError, http.StatusUnprocessableEntity)
		return
	}

	response := RequestResponse{
		Status:         "Withdraw operation executed with success",
		CurrentBalance: caller.CurrentBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func validateDestiny(account int32) (*Account, string) {
	for index, currentAccount := range accounts {
		if currentAccount.AccountID == account {
			return &accounts[index], ""
		}
	}
	return &Account{}, "the destiny account does not exists"
}

func transferPost(w http.ResponseWriter, r *http.Request) {
	var requestBody Destiny
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	caller, error := validateCaller(requestBody.AccountID, requestBody.Password)
	if error != "" {
		http.Error(w, error, http.StatusBadRequest)
		return
	}

	destiny, destinyError := validateDestiny(requestBody.DestinyID)
	if destinyError != "" {
		http.Error(w, destinyError, http.StatusBadRequest)
		return
	}

	if operationError := changeAmount(caller, requestBody.Value, "withdraw"); operationError != "" {
		http.Error(w, operationError, http.StatusUnprocessableEntity)
		return
	}

	if operationError := changeAmount(destiny, requestBody.Value, "deposit"); operationError != "" {
		http.Error(w, operationError, http.StatusUnprocessableEntity)
		return
	}

	response := RequestResponse{
		Status:         "Transfer operation executed with success",
		CurrentBalance: caller.CurrentBalance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func listAllGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

func main() {
	fmt.Println("Welcome to the Goliath National Bank")
	accounts = make([]Account, 0)

	accounts = append(accounts, leeAccount)

	http.HandleFunc("/api/create", createAccount)
	http.HandleFunc("/api/deposit", depositPost)
	http.HandleFunc("/api/account", accountGet)
	http.HandleFunc("/api/withdraw", withdrawPost)
	http.HandleFunc("/api/transfer", transferPost)

	http.HandleFunc("/api/listall", listAllGet)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
