package models

import "time"

// Book represents an ebook in our collection
type Book struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	FilePath string    `json:"file_path"`
	FileSize int64     `json:"file_size"`
	Format   string    `json:"format"`
	AddedAt  time.Time `json:"added_at"`
}

// BookRequest represents a request to add/update a book
type BookRequest struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	Format   string `json:"format"`
}

// ScanRequest represents a request to scan a directory
type ScanRequest struct {
	Path string `json:"path"`
}

// ScanResponse represents the response from a scan operation
type ScanResponse struct {
	Status  string `json:"status"`
	Added   int    `json:"added,omitempty"`
	Removed int    `json:"removed,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
