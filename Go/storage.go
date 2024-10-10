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
	err := s.createUsersTable()
	if err != nil {
		return err
	}
	err = s.createTables()
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) createFileQueueTable() error {
	query := `
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
