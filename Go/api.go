package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)

}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type apiError struct {
	Error string `json:"error"`
}

func makeHTTPhandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			/* http.Error(w, err.Error(), http.StatusInternalServerError) */
			WriteJson(w, http.StatusInternalServerError, apiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAdress string
	store        *SQLiteStorage
}

func newAPIServer(listenAdress string, store *SQLiteStorage) *APIServer {
	return &APIServer{
		listenAdress: listenAdress,
		store:        store,
	}
}

func (s *APIServer) Run() {

	router := mux.NewRouter()
	router.HandleFunc("/account", makeHTTPhandler(s.handleAccount))
	router.HandleFunc("/accounts", makeHTTPhandler(s.handleGetAccounts))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPhandler(s.handleGetAccount), s.store))
	router.HandleFunc("/transfer/", makeHTTPhandler(s.handleTransfer))

	router.HandleFunc("/files", makeHTTPhandler(s.handleFiles))
	router.HandleFunc("/parse", makeHTTPhandler(s.handleFiles))

	log.Println("Starting server on", s.listenAdress)

	http.ListenAndServe(s.listenAdress, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
	} else if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	} else if r.Method == http.MethodDelete {
		return s.handleDelete(w, r)
	} else if r.Method == http.MethodPut {
		return s.handleTransfer(w, r)
	}

	return fmt.Errorf("Method not allowed")
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		id, err := getId(r)
		if err != nil {
			return err
		}
		account, err := s.store.GetAccountByID(int64(id))
		if err != nil {
			return err
		}

		return WriteJson(w, http.StatusOK, account)
	}
	if r.Method == http.MethodDelete {
		return s.handleDelete(w, r)
	}

	return fmt.Errorf("Method not allowed")
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, "", 0)
	fmt.Printf("%+v\n", account)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, _ := CreateJWT(account)
	fmt.Println("TokenString:", tokenString)

	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJson(w, http.StatusOK, accounts)
}

func (s *APIServer) handleDelete(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJson(w, http.StatusOK, map[string]int64{"deleted": id})
}
func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	TransferRequest := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(TransferRequest); err != nil {
		return err
	}

	defer r.Body.Close()

	return WriteJson(w, http.StatusOK, TransferRequest)
}

func getId(r *http.Request) (int64, error) {
	id_str := mux.Vars(r)["id"]
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %s, %w ", id_str, err)
	}
	return int64(id), nil
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVzQXQiOjE3Mjc4OTc2MjgsImFjY291bnROdW1iZXIiOjQwODU4fQ.V1HkJrkib5-wiEzOVyORBkzVFPZ4EtbnyfPuPyGbSNs
func withJWTAuth(HandlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT Auth middleware")
		tokenString := r.Header.Get("Authorization")
		token, err := ValidateJWT(tokenString)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, apiError{Error: "Invalid token"})
			return
		}

		if !token.Valid {
			WriteJson(w, http.StatusUnauthorized, apiError{Error: "Invalid token"})
			return
		}

		userID, err := getId(r)
		if err != nil {
			PermissionDenied(w)
			return
		}

		account, err := s.GetAccountByID(int64(userID))
		if err != nil {
			PermissionDenied(w)
			return
		}
		fmt.Println("Account:", account)

		HandlerFunc(w, r)
	}
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	fmt.Println("Secret in ValidateJWT:", secret)
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func CreateJWT(account *Account) (string, error) {
	claims := jwt.MapClaims{
		"ExpiresAt":     time.Now().Add(time.Hour * 24).Unix(),
		"accountNumber": account.Number,
	}
	secret := os.Getenv("JWT_SECRET")
	fmt.Println("Secret in CreateJWT:", secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}

func PermissionDenied(w http.ResponseWriter) {
	WriteJson(w, http.StatusUnauthorized, apiError{Error: "Permission Denied"})
}

// Files ---------------------------------------------------------
func (s *APIServer) handleFiles(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleFileGet(w, r)
	} else if r.Method == http.MethodPost {
		return s.handleUploadFile(w, r)
	}

	return fmt.Errorf("Method not allowed")
}

func (s *APIServer) handleUploadFile(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return fmt.Errorf("error parsing form: %v", err)
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return fmt.Errorf("error retrieving the file: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	file.Seek(0, 0)

	mimeType := http.DetectContentType(buf)
	if mimeType != "application/pdf" {
		http.Error(w, "The uploaded file is not a PDF", http.StatusBadRequest)
		return nil
	}

	UserID := r.Header.Get("Authorization")
	if UserID == "" {
		WriteJson(w, http.StatusUnauthorized, apiError{Error: "No user ID found in request"})
		return fmt.Errorf("No user ID found in request")
	}
	userIDInt, err := strconv.Atoi(UserID)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, apiError{Error: "Invalid user ID"})
		return fmt.Errorf("Invalid user ID")
	}

	// Retrieve the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		return fmt.Errorf("error retrieving the file: %v", err)
	}
	defer file.Close()

	// Save the file to the database
	err = s.store.SaveFileToDB(userIDInt, file, header.Filename)
	if err != nil {
		return fmt.Errorf("error saving file to the database: %v", err)
	}

	// Respond with success
	return WriteJson(w, http.StatusOK, map[string]string{
		"message": "All good in the hood",
	})
}

func (s *APIServer) handleFileGet(w http.ResponseWriter, r *http.Request) error {
	UserID := r.Header.Get("Authorization")
	if UserID == "" {
		WriteJson(w, http.StatusUnauthorized, apiError{Error: "No user ID found in request"})
		return fmt.Errorf("No user ID found in request")
	}
	userIDInt, err := strconv.Atoi(UserID)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, apiError{Error: "Invalid user ID"})
		return fmt.Errorf("Invalid user ID")
	}
	files, err := s.store.getFiles(userIDInt)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, apiError{Error: err.Error()})
	}
	WriteJson(w, http.StatusOK, files)
	return nil
}
