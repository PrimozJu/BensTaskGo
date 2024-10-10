package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage() (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", "Bens_DB_new.db")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Init() error {

	err := s.createTables()
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) createTables() error { /* could be seperate funcs */
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

	CREATE TABLE IF NOT EXISTS file_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			queue_status TEXT NOT NULL,
			queue_updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(file_id) REFERENCES files(id)
		);

		CREATE TABLE IF NOT EXISTS file_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			queue_status TEXT NOT NULL,
			queue_updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(file_id) REFERENCES files(id)
		);

	`
	_, err := s.db.Exec(query)
	return err
}
