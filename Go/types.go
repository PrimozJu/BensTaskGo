package main

import (
	"database/sql"
	"time"
)

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"toAmount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type File struct {
	ID              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID          int       `json:"user_id" gorm:"not null"`
	OriginalName    string    `json:"original_name" gorm:"size:255;not null"`
	FileContentHash string    `json:"file_content_hash" gorm:"size:64;not null"`
	UploadDate      time.Time `json:"upload_date" gorm:"autoCreateTime"`
	ParseStatus     string    `json:"parse_status" gorm:"type:ENUM('queued', 'in_progress', 'parsed', 'failed');not null"`
	ParseResult     string    `json:"parse_result" gorm:"type:json"`
	ImportStatus    int       `json:"import_status" gorm:"default:0"`
}

type FileQueue struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	FileID         int       `json:"file_id" gorm:"not null"`
	QueueStatus    string    `json:"queue_status" gorm:"type:ENUM('queued', 'processing', 'completed');not null"`
	QueueUpdatedAt time.Time `json:"queue_updated_at" gorm:"autoUpdateTime"`
}

type FileMetadata struct {
	OriginalName    string
	UploadDate      time.Time
	FileContentHash string
	QueueStatus     sql.NullString
	QueueUpdateAt   sql.NullTime
}
