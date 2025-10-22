package models

import "time"

// Book represents an ebook in our collection
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	Format    string    `json:"format"`
	ISBN      string    `json:"isbn"`
	Publisher string    `json:"publisher"`
	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BookRequest represents a request to add/update a book
type BookRequest struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	Format    string `json:"format"`
	ISBN      string `json:"isbn"`
	Publisher string `json:"publisher"`
}

// QuarantineBook represents a book in quarantine with additional quarantine information
type QuarantineBook struct {
	Book
	QuarantineReason string `json:"quarantine_reason,omitempty"`
	QuarantineDetail string `json:"quarantine_detail,omitempty"`
	QuarantineDate   string `json:"quarantine_date,omitempty"`
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

// MetadataSearchRequest represents a request to search for book metadata
type MetadataSearchRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// MetadataSearchResponse represents the response from metadata search
type MetadataSearchResponse struct {
	Suggestions []MetadataSuggestion `json:"suggestions"`
	Confidence  float64              `json:"confidence"`
	Message     string               `json:"message,omitempty"`
}

// MetadataSuggestion represents a suggested metadata from external source
type MetadataSuggestion struct {
	Title      string  `json:"title"`
	Author     string  `json:"author"`
	ISBN       string  `json:"isbn"`
	Publisher  string  `json:"publisher"`
	Year       int     `json:"year"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}
