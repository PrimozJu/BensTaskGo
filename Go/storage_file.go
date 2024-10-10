package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"

	_ "modernc.org/sqlite"
)

func (s *SQLiteStorage) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_hash TEXT UNIQUE NOT NULL,
		file_content BLOB NOT NULL,
		original_name TEXT NOT NULL,
		upload_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

	CREATE TABLE IF NOT EXISTS file_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		original_name TEXT NOT NULL,
		file_id INTEGER NOT NULL,
		upload_date Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		parsing_result INT default 0,
		FOREIGN KEY (file_id) REFERENCES files(id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (parsing_result) REFERENCES parse_result(id)
);

	CREATE TABLE IF NOT EXISTS parse_result (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		parse_status TEXT NOT NULL
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStorage) getFiles(userID int) ([]*FileMetadata, error) {
	query := `
    SELECT 
        file_metadata.original_name, 
        file_metadata.upload_date, 
        files.file_hash, 
        file_queue.queue_status, 
        file_queue.queue_updated_at
    FROM 
        file_metadata
    INNER JOIN 
        files ON file_metadata.file_id = files.id
    LEFT JOIN 
        file_queue ON files.id = file_queue.file_id
    WHERE 
        file_metadata.user_id = ?`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying files: %v", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		f := new(FileMetadata)
		err := rows.Scan(&f.OriginalName, &f.UploadDate, &f.FileContentHash, &f.QueueStatus, &f.QueueUpdateAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning file: %v", err)
		}
		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return files, nil
}

func (s *SQLiteStorage) SaveFileToDB(userID int, file multipart.File, fileName string) error {
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	fileHash := generateFileHash(fileContent)

	existingFileID := 0
	err = s.db.QueryRow("SELECT id FROM files WHERE file_hash = ?", fileHash).Scan(&existingFileID)
	println(existingFileID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error querying for existing file: %v", err)
	}

	if existingFileID > 0 {
		/* dodaj check da nemore 1 user 2x dodat istega fajla */
		err = s.insertIntoMetadata(existingFileID, userID, fileName)
		if err != nil {
			return fmt.Errorf("error saving file metadata: %v", err)
		}
		log.Printf("File already exists in DB. Saved metadata for new upload with existing file.")
		return nil
	} else {
		newFileID, err := s.insertIntoFiles(fileHash, fileContent, fileName)
		if err != nil {
			return fmt.Errorf("error inserting file into database: %v", err)
		}
		log.Printf("File does not exist in DB. Inserted new file.")

		err = s.insertIntoMetadata(newFileID, userID, fileName)
		if err != nil {
			return fmt.Errorf("error saving file metadata: %v", err)
		}
		log.Printf("Saved metadata for new upload with new file.")

		err = s.insertIntoFileQueue(newFileID)
		if err != nil {
			return fmt.Errorf("error saving file metadata: %v", err)
		}
		log.Printf("New file has been added to a queue.")

		return nil

	}
}

func generateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

func (s *SQLiteStorage) insertIntoMetadata(existingFileID int, userID int, fileName string) error {
	metadataQuery := `
			INSERT INTO file_metadata (file_id, user_id, original_name, upload_date)
			VALUES (?, ?, ?, datetime('now'));`
	_, err := s.db.Exec(metadataQuery, existingFileID, userID, fileName)
	if err != nil {
		return fmt.Errorf("error saving file metadata: %v", err)
	}
	log.Println("file already exists in DB. Saved metadata for new upload with existing file.")
	return nil
}

func (s *SQLiteStorage) insertIntoFiles(fileHash string, fileContent []byte, fileName string) (int, error) {
	insertFileQuery := `
    INSERT INTO files (file_hash, file_content, original_name, upload_date)
    VALUES (?, ?,  ?, datetime('now'));`
	result, err := s.db.Exec(insertFileQuery, fileHash, fileContent, fileName)

	if err != nil {
		return 0, fmt.Errorf("error inserting file into database: %v", err)
	}

	fileID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %v", err)
	}
	log.Printf("file inserted with ID: %d", fileID)
	return int(fileID), nil
}

func (s *SQLiteStorage) insertIntoFileQueue(fileID int) error {
	insertFileQuery := `
	INSERT INTO file_queue (file_id, queue_status, queue_updated_at)
	VALUES (?, 'queued', datetime('now'));`
	_, err := s.db.Exec(insertFileQuery, fileID)

	if err != nil {
		return fmt.Errorf("error inserting file into database: %v", err)
	}
	log.Printf("file inserted into queue with ID: %d", fileID)
	return nil
}
