package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fableflow/backend/conversion"
	"fableflow/backend/database"
)

// ConversionHandler handles ebook conversion requests
type ConversionHandler struct {
	db     *database.Manager
	tmpDir string
}

// TempFileInfo tracks temporary conversion files
type TempFileInfo struct {
	Path       string
	CreatedAt  time.Time
	Downloaded bool
	BookID     int
	Format     string
}

// Global map to track temporary files
var tempFiles = make(map[string]*TempFileInfo)

// NewConversionHandler creates a new conversion handler
func NewConversionHandler(db *database.Manager, tmpDir string) *ConversionHandler {
	return &ConversionHandler{
		db:     db,
		tmpDir: tmpDir,
	}
}

// ConvertBook converts a book to a different format
func (h *ConversionHandler) ConvertBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		BookID       int    `json:"book_id"`
		OutputFormat string `json:"output_format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate output format
	if req.OutputFormat != "azw3" {
		http.Error(w, "Only AZW3 conversion is currently supported", http.StatusBadRequest)
		return
	}

	// Get book details
	book, err := h.db.GetBookByID(req.BookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if file exists
	if _, err := os.Stat(book.FilePath); os.IsNotExist(err) {
		http.Error(w, "Source file not found", http.StatusNotFound)
		return
	}

	// Check if it's an EPUB file
	if !strings.HasSuffix(strings.ToLower(book.FilePath), ".epub") {
		http.Error(w, "Only EPUB files can be converted to AZW3", http.StatusBadRequest)
		return
	}

	// Generate temporary output path using config tmp_dir
	tempDir := filepath.Join(h.tmpDir, "conversions")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}

	// Create filename based on original book filename
	originalFilename := filepath.Base(book.FilePath)
	nameWithoutExt := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
	tempFilename := fmt.Sprintf("%s.%s", nameWithoutExt, req.OutputFormat)
	outputPath := filepath.Join(tempDir, tempFilename)

	// Perform conversion
	fmt.Printf("Starting conversion: %s -> %s\n", book.FilePath, outputPath)
	err = conversion.ConvertEPUBToAZW3(book.FilePath, outputPath)
	if err != nil {
		fmt.Printf("Conversion failed: %v\n", err)
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Conversion completed successfully\n")

	// Track the temporary file
	tempFileKey := fmt.Sprintf("%d_%s", req.BookID, req.OutputFormat)
	tempFiles[tempFileKey] = &TempFileInfo{
		Path:       outputPath,
		CreatedAt:  time.Now(),
		Downloaded: false,
		BookID:     req.BookID,
		Format:     req.OutputFormat,
	}

	// Start cleanup timer (remove file after 1 hour if not downloaded)
	go func() {
		time.Sleep(1 * time.Hour)
		if tempFile, exists := tempFiles[tempFileKey]; exists && !tempFile.Downloaded {
			os.Remove(tempFile.Path)
			delete(tempFiles, tempFileKey)
			fmt.Printf("Cleaned up temporary file: %s\n", tempFile.Path)
		}
	}()

	// Return success response
	response := map[string]interface{}{
		"success":       true,
		"output_format": req.OutputFormat,
		"message":       "Conversion completed successfully. File will be available for download for 1 hour.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConversionStatus returns the status of the conversion service
func (h *ConversionHandler) GetConversionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"available":         true,
		"supported_formats": []string{"epub"},
		"output_formats":    []string{"azw3"},
		"description":       "EPUB to AZW3 conversion using leotaku/mobi library",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DownloadConvertedBook downloads a converted book
func (h *ConversionHandler) DownloadConvertedBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract book ID and format from URL path
	// Expected format: /api/convert/{book_id}/{format}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	bookIDStr := pathParts[3]
	format := pathParts[4]

	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Get book details (for validation)
	_, err = h.db.GetBookByID(bookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if converted file exists in temp storage
	tempFileKey := fmt.Sprintf("%d_%s", bookID, format)
	tempFile, exists := tempFiles[tempFileKey]
	if !exists {
		http.Error(w, "Converted file not found. Please convert the book first.", http.StatusNotFound)
		return
	}

	outputPath := tempFile.Path
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		http.Error(w, "Converted file not found. Please convert the book first.", http.StatusNotFound)
		return
	}

	// Set headers for file download
	filename := filepath.Base(outputPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Open and serve the file
	file, err := os.Open(outputPath)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy file to response
	io.Copy(w, file)

	// Mark file as downloaded and schedule cleanup
	tempFile.Downloaded = true
	go func() {
		// Wait a bit to ensure download completes, then clean up
		time.Sleep(30 * time.Second)
		os.Remove(outputPath)
		delete(tempFiles, tempFileKey)
		fmt.Printf("Cleaned up downloaded file: %s\n", outputPath)
	}()
}
