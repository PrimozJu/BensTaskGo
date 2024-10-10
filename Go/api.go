package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

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
	router.HandleFunc("/files", makeHTTPhandler(s.handleFiles))
	router.HandleFunc("/parse", makeHTTPhandler(s.getNextFileForParsingHandler))
	router.HandleFunc("/parse-result", makeHTTPhandler(s.receiveParseResultHandler))
	log.Println("Starting server on", s.listenAdress)

	go startParser()
	http.ListenAndServe(s.listenAdress, router)

}

/* func getId(r *http.Request) (int64, error) {
	id_str := mux.Vars(r)["id"]
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return 0, fmt.Errorf("invalid ID %s, %w ", id_str, err)
	}
	return int64(id), nil
} */

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

	return fmt.Errorf("method not allowed")
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
		return fmt.Errorf("no user ID found in request")
	}
	userIDInt, err := strconv.Atoi(UserID)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, apiError{Error: "Invalid user ID"})
		return fmt.Errorf("invalid user ID")
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return fmt.Errorf("error retrieving the file: %v", err)
	}
	defer file.Close()

	err = s.store.SaveFileToDB(userIDInt, file, header.Filename)
	if err != nil {
		return fmt.Errorf("error saving file to the database: %v", err)
	}

	return WriteJson(w, http.StatusOK, map[string]string{
		"message": "File has been uploaded successfully",
	})
}

func (s *APIServer) handleFileGet(w http.ResponseWriter, r *http.Request) error {
	UserID := r.Header.Get("Authorization")
	if UserID == "" {
		WriteJson(w, http.StatusUnauthorized, apiError{Error: "No user ID found in request"})
		return fmt.Errorf("no user ID found in request")
	}
	userIDInt, err := strconv.Atoi(UserID)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, apiError{Error: "Invalid user ID"})
		return fmt.Errorf("invalid user ID")
	}
	files, err := s.store.getFiles(userIDInt)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, apiError{Error: err.Error()})
	}
	WriteJson(w, http.StatusOK, files)
	return nil
}

func (s *APIServer) getNextFileForParsingHandler(w http.ResponseWriter, r *http.Request) error {
	var file FileData
	err := s.store.GetNextQueuedFile(&file)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No files in the queue", http.StatusNotFound)
			return nil
		}
		return fmt.Errorf("error retrieving file: %v", err)
	}

	return json.NewEncoder(w).Encode(file)
}

func (s *APIServer) receiveParseResultHandler(w http.ResponseWriter, r *http.Request) error {
	println("6: Receive parse result CALLED")
	var parseResult ParseResult

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %v", err)
	}
	log.Printf("Raw request body: %s", string(body))

	err = json.Unmarshal(body, &parseResult)
	if err != nil {
		return fmt.Errorf("error decoding parse result: %v", err)
	}
	fmt.Printf("Received parse result: %+v\n", parseResult)

	err = s.store.UpdateFileParseStatus(parseResult)
	if err != nil {
		return fmt.Errorf("error updating file status: %v", err)
	}

	fmt.Fprintf(w, "File status updated successfully")
	return nil
}
