package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ParseResult struct {
	FileID int         `json:"file_id"`
	Status string      `json:"status"`
	Result interface{} `json:"result"`
}

type FileData struct {
	FileID  int    `json:"file_id"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
	Name    string `json:"name"`
}

func (s *SQLiteStorage) UpdateFileParseStatus(result ParseResult) error {
	println("7: Update file parse status CALLED")
	log.Printf("Updating file parse status: %+v, %+v, %+v", result.Result, result.Status, result.FileID)
	println("----")
	println(result.Status)
	println(result.FileID)
	query := `
        UPDATE file_queue
        SET queue_status = ?
        WHERE file_id = ?`

	_, err := s.db.Exec(query, result.Status, result.FileID)
	if err != nil {
		log.Fatalf("Error updating file parse status: %v", err)
	} else {
		log.Printf("File parse status updated successfully")
	}
	return err
}

func requestNextFile() {
	println("1: Request next file CALLED")
	resp, err := http.Get("http://127.0.0.1:3000/parse")
	if err != nil {
		log.Fatalf("Error requesting next file: %v", err)
	}
	println("2: ReSPONSE GOTTEN ", resp)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		println("3: Status code je ok")
		var fileData FileData
		err := json.NewDecoder(resp.Body).Decode(&fileData)
		if err != nil {
			log.Fatalf("Error decoding file data: %v", err)
		}

		parseResult := parseFile(fileData)
		log.Println("4: Parse result je ", parseResult)

		sendParseResult(fileData.FileID, parseResult)
	} else {
		println("No files in the queue")
	}
}

func (s *SQLiteStorage) GetNextQueuedFile(file *FileData) error {
	query := `
        SELECT f.id, f.file_content, f.file_hash
        FROM files f
        JOIN file_queue q ON f.id = q.file_id
        WHERE q.queue_status == 'queued'
        LIMIT 1`

	return s.db.QueryRow(query).Scan(&file.FileID, &file.Content, &file.Name)

}

func parseFile(file FileData) ParseResult {
	println("3.5 Parse file CALLED")
	return ParseResult{
		FileID: file.FileID,
		Status: "success",
		Result: `{"parsed_content": "example"}`,
	}
}

// Send the result back to the main server
func sendParseResult(fileID int, result ParseResult) {
	println("5: Send parse result CALLED")
	log.Printf("Sending parse result for file %d: %+v", fileID, result.Status)
	parseResultPayload := ParseResult{
		FileID: fileID,
		Status: result.Status,
		Result: result.Result,
	}
	log.Printf("Sending parse result payload: %+v", parseResultPayload)

	jsonData, err := json.Marshal(parseResultPayload)
	if err != nil {
		log.Fatalf("Error marshaling parse result payload: %v", err)
	}
	log.Printf("JSON data to send: %s", string(jsonData))

	resp, err := http.Post("http://127.0.0.1:3000/parse-result", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending parse result: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("Response status: %s", resp.Status)
}

func startParser() {
	for {
		println("Requesting next file...")
		time.Sleep(20 * time.Second) // 20s za simulacijo
		requestNextFile()
	}
}
