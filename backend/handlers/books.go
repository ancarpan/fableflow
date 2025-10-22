package handlers

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"fableflow/backend/config"
	"fableflow/backend/database"
	"fableflow/backend/epub"
	"fableflow/backend/models"
)

// BooksHandler handles book-related HTTP requests
type BooksHandler struct {
	db     *database.Manager
	config *config.Config
}


// NewBooksHandler creates a new books handler
func NewBooksHandler(db *database.Manager, config *config.Config) *BooksHandler {
	return &BooksHandler{db: db, config: config}
}

// GetAllBooks returns all books
func (h *BooksHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.db.GetAllBooks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// SearchBooks searches for books by title or author
func (h *BooksHandler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// If no query, return all books
		h.GetAllBooks(w, r)
		return
	}

	books, err := h.db.SearchBooks(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetBookByID returns a specific book by ID
func (h *BooksHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	// Handle different HTTP methods
	if r.Method == "PUT" {
		// This is an edit request, delegate to EditBookMetadata
		h.EditBookMetadata(w, r)
		return
	}

	// Extract ID from URL path (assuming /api/books/{id})
	// This is a simplified version - in a real app you'd use a router
	idStr := r.URL.Path[len("/api/books/"):]

	// Handle cases where the path might be /api/books/{id}/edit
	if strings.Contains(idStr, "/") {
		idStr = strings.Split(idStr, "/")[0]
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// For now, we'll get all books and find the one with matching ID
	// In a real app, you'd have a GetBookByID method in the database layer
	books, err := h.db.GetAllBooks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, book := range books {
		if book.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(book)
			return
		}
	}

	http.Error(w, "Book not found", http.StatusNotFound)
}

// AddBook adds a new book
func (h *BooksHandler) AddBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var book models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if book.Title == "" || book.FilePath == "" {
		http.Error(w, "Title and file path are required", http.StatusBadRequest)
		return
	}

	err := h.db.AddBook(book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "book added"})
}

// RemoveBook removes a book by ID
func (h *BooksHandler) RemoveBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/api/books/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	err = h.db.RemoveBook(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "book removed"})
}

// GetAuthors returns all unique authors
func (h *BooksHandler) GetAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := h.db.GetAllAuthors()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null
	if authors == nil {
		authors = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

// GetAuthorsByLetter returns authors starting with a specific letter
func (h *BooksHandler) GetAuthorsByLetter(w http.ResponseWriter, r *http.Request) {
	letter := r.URL.Query().Get("letter")
	if letter == "" {
		http.Error(w, "Letter parameter is required", http.StatusBadRequest)
		return
	}

	authors, err := h.db.GetAuthorsByLetter(letter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

// GetBooksByAuthor returns all books by a specific author
func (h *BooksHandler) GetBooksByAuthor(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	if author == "" {
		http.Error(w, "Author parameter is required", http.StatusBadRequest)
		return
	}

	books, err := h.db.GetBooksByAuthor(author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetTitles returns all unique titles
func (h *BooksHandler) GetTitles(w http.ResponseWriter, r *http.Request) {
	titles, err := h.db.GetAllTitles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null
	if titles == nil {
		titles = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(titles)
}

// GetTitlesByLetter returns titles starting with a specific letter
func (h *BooksHandler) GetTitlesByLetter(w http.ResponseWriter, r *http.Request) {
	letter := r.URL.Query().Get("letter")
	if letter == "" {
		http.Error(w, "Letter parameter is required", http.StatusBadRequest)
		return
	}

	titles, err := h.db.GetTitlesByLetter(letter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(titles)
}

// GetRecentBooks returns the most recently added books
func (h *BooksHandler) GetRecentBooks(w http.ResponseWriter, r *http.Request) {
	// Get limit from query parameter, default to 12
	limitStr := r.URL.Query().Get("limit")
	limit := 12
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	books, err := h.db.GetRecentBooks(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null
	if books == nil {
		books = []models.Book{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetRandomBooks returns a random selection of books
func (h *BooksHandler) GetRandomBooks(w http.ResponseWriter, r *http.Request) {
	// Get limit from query parameter, default to 12
	limitStr := r.URL.Query().Get("limit")
	limit := 12
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	books, err := h.db.GetRandomBooks(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we return an empty array instead of null
	if books == nil {
		books = []models.Book{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetBooksByTitle returns all books with a specific title
func (h *BooksHandler) GetBooksByTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Title parameter is required", http.StatusBadRequest)
		return
	}

	books, err := h.db.GetBooksByTitle(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// DownloadBook downloads a book file by ID
func (h *BooksHandler) DownloadBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path (remove .epub extension if present)
	idStr := r.URL.Path[len("/api/download/"):]

	// Remove .epub extension if present
	if strings.HasSuffix(idStr, ".epub") {
		idStr = idStr[:len(idStr)-5] // Remove ".epub" (5 characters)
	}

	// Log for debugging
	fmt.Printf("URL: %s, ID string: %s\n", r.URL.Path, idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Get book details
	book, err := h.db.GetBookByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if file exists
	if _, err := os.Stat(book.FilePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for EPUB file serving
	filename := filepath.Base(book.FilePath)
	w.Header().Set("Content-Type", "application/epub+zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Open and serve the file
	file, err := os.Open(book.FilePath)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy file to response
	io.Copy(w, file)
}

// ServeReader serves the EPUB reader page
func (h *BooksHandler) ServeReader(w http.ResponseWriter, r *http.Request) {
	// Extract book ID from URL path
	bookIDStr := r.URL.Path[len("/read/"):]
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Get book from database
	book, err := h.db.GetBookByID(bookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if it's an EPUB file
	if book.Format != "epub" {
		http.Error(w, "Only EPUB files can be read", http.StatusBadRequest)
		return
	}

	// Serve the reader HTML page
	readerPath := filepath.Join("..", "frontend", "templates", "reader.html")
	http.ServeFile(w, r, readerPath)
}

// ServeEPUBFile serves internal EPUB files (like META-INF/container.xml)
func (h *BooksHandler) ServeEPUBFile(w http.ResponseWriter, r *http.Request) {
	// Extract book ID and file path from URL
	// URL format: /api/epub/{bookID}/{filepath}
	path := r.URL.Path[len("/api/epub/"):]
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid EPUB file path", http.StatusBadRequest)
		return
	}

	bookIDStr := parts[0]
	filePath := parts[1]

	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Get book from database
	book, err := h.db.GetBookByID(bookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if it's an EPUB file
	if book.Format != "epub" {
		http.Error(w, "Only EPUB files can be read", http.StatusBadRequest)
		return
	}

	// Open the EPUB file as a ZIP archive
	reader, err := zip.OpenReader(book.FilePath)
	if err != nil {
		http.Error(w, "Failed to open EPUB file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Find the requested file in the EPUB
	for _, file := range reader.File {
		if file.Name == filePath {
			// Open the file
			rc, err := file.Open()
			if err != nil {
				http.Error(w, "Failed to open file in EPUB", http.StatusInternalServerError)
				return
			}
			defer rc.Close()

			// Set appropriate content type
			ext := strings.ToLower(filepath.Ext(filePath))
			switch ext {
			case ".xml":
				w.Header().Set("Content-Type", "application/xml")
			case ".xhtml", ".html":
				w.Header().Set("Content-Type", "application/xhtml+xml")
			case ".css":
				w.Header().Set("Content-Type", "text/css")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".gif":
				w.Header().Set("Content-Type", "image/gif")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			default:
				w.Header().Set("Content-Type", "application/octet-stream")
			}

			// Copy file content to response
			_, err = io.Copy(w, rc)
			if err != nil {
				http.Error(w, "Failed to serve file content", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	// File not found in EPUB
	http.Error(w, "File not found in EPUB", http.StatusNotFound)
}

// EditBookMetadata handles editing book metadata
func (h *BooksHandler) EditBookMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract book ID from URL path
	// URL format: /api/books/{id}/edit
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[4] != "edit" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	bookID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var editRequest struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		ISBN      string `json:"isbn"`
		Publisher string `json:"publisher"`
	}

	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get book from database
	book, err := h.db.GetBookByID(bookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if it's an EPUB file
	if book.Format != "epub" {
		http.Error(w, "Only EPUB files can be edited", http.StatusBadRequest)
		return
	}

	// Create EPUB editor and load the file
	editor := epub.NewEPUBEditor(book.FilePath)
	if err := editor.Load(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to load EPUB file: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metadata in the EPUB file
	if err := editor.UpdateMetadata(editRequest.Title, editRequest.Author, editRequest.ISBN, editRequest.Publisher); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update EPUB metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the modified EPUB file
	if err := editor.Save(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save EPUB file: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if author or title changed to determine if file needs to be moved
	needsFileMove := (book.Author != editRequest.Author) || (book.Title != editRequest.Title)
	var newFilePath string

	if needsFileMove {
		// Generate new file path based on new author/title
		newFilePath = h.generateNewFilePath(editRequest.Author, editRequest.Title, book.Format)

		// Move the file to new location
		if err := h.moveBookFile(book.FilePath, newFilePath); err != nil {
			http.Error(w, fmt.Sprintf("Failed to move file: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Keep the same file path
		newFilePath = book.FilePath
	}

	// Update database with new metadata and file path
	if needsFileMove {
		if err := h.db.UpdateBookWithPath(bookID, editRequest.Title, editRequest.Author, editRequest.ISBN, editRequest.Publisher, newFilePath); err != nil {
			http.Error(w, "Failed to update database", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.db.UpdateBook(bookID, editRequest.Title, editRequest.Author, editRequest.ISBN, editRequest.Publisher); err != nil {
			http.Error(w, "Failed to update database", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Book metadata updated successfully",
	})
}

// LookupISBN handles ISBN lookup requests
func (h *BooksHandler) LookupISBN(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ISBN string `json:"isbn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.ISBN == "" {
		http.Error(w, "ISBN is required", http.StatusBadRequest)
		return
	}

	// Lookup metadata from Google Books API
	metadata, err := h.lookupGoogleBooks(request.ISBN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// lookupGoogleBooks queries Google Books API for book metadata
func (h *BooksHandler) lookupGoogleBooks(isbn string) (map[string]interface{}, error) {
	// Clean ISBN (remove hyphens, spaces, etc.)
	cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	// Google Books API URL
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=isbn:%s", cleanISBN)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query Google Books API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google Books API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Items []struct {
			VolumeInfo struct {
				Title               string   `json:"title"`
				Authors             []string `json:"authors"`
				Publisher           string   `json:"publisher"`
				PublishedDate       string   `json:"publishedDate"`
				Description         string   `json:"description"`
				IndustryIdentifiers []struct {
					Type       string `json:"type"`
					Identifier string `json:"identifier"`
				} `json:"industryIdentifiers"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse Google Books response: %v", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no book found for ISBN: %s", isbn)
	}

	// Extract metadata from first result
	book := result.Items[0].VolumeInfo

	// Find ISBN in industry identifiers - prefer the same format as search term
	var foundISBN string
	searchISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	// Determine if search term is ISBN-13 (starts with 978 or 979) or ISBN-10
	var preferISBN13 bool
	if len(searchISBN) == 13 && (strings.HasPrefix(searchISBN, "978") || strings.HasPrefix(searchISBN, "979")) {
		preferISBN13 = true
	}

	// Look for preferred format first
	for _, id := range book.IndustryIdentifiers {
		if preferISBN13 && id.Type == "ISBN_13" {
			foundISBN = id.Identifier
			break
		} else if !preferISBN13 && id.Type == "ISBN_10" {
			foundISBN = id.Identifier
			break
		}
	}

	// Fallback to any ISBN if preferred format not found
	if foundISBN == "" {
		for _, id := range book.IndustryIdentifiers {
			if id.Type == "ISBN_13" || id.Type == "ISBN_10" {
				foundISBN = id.Identifier
				break
			}
		}
	}

	// Final fallback
	if foundISBN == "" && len(book.IndustryIdentifiers) > 0 {
		foundISBN = book.IndustryIdentifiers[0].Identifier
	}

	// Format authors
	var author string
	if len(book.Authors) > 0 {
		author = strings.Join(book.Authors, ", ")
	}

	return map[string]interface{}{
		"title":          book.Title,
		"author":         author,
		"publisher":      book.Publisher,
		"isbn":           foundISBN,
		"published_date": book.PublishedDate,
		"description":    book.Description,
	}, nil
}

// generateNewFilePath creates a new file path based on author and title
func (h *BooksHandler) generateNewFilePath(author, title, format string) string {
	// Clean author and title for filesystem
	cleanAuthor := h.cleanForFilesystem(author)
	cleanTitle := h.cleanForFilesystem(title)

	// Create directory structure: Author/Title/
	// Use scan directory from config
	dirPath := filepath.Join(h.config.Library.ScanDirectory, cleanAuthor, cleanTitle)

	// Generate filename: Title - Author.epub
	filename := fmt.Sprintf("%s - %s.%s", cleanTitle, cleanAuthor, format)

	return filepath.Join(dirPath, filename)
}

// cleanForFilesystem removes invalid characters for filesystem paths
func (h *BooksHandler) cleanForFilesystem(s string) string {
	// Remove or replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := s
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "")
	}

	// Trim whitespace
	result = strings.TrimSpace(result)

	// Ensure it's not empty
	if result == "" {
		result = "Unknown"
	}

	return result
}

// moveBookFile moves a book file to a new location
func (h *BooksHandler) moveBookFile(oldPath, newPath string) error {
	// Create the new directory if it doesn't exist
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", newDir, err)
	}

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %v", oldPath, newPath, err)
	}

	// Clean up empty directories from the old location
	if err := h.cleanupEmptyDirectories(filepath.Dir(oldPath)); err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Warning: failed to cleanup empty directories: %v\n", err)
	}

	return nil
}

// GetQuarantineBooks returns all books in the quarantine directory
func (h *BooksHandler) GetQuarantineBooks(w http.ResponseWriter, r *http.Request) {
	// Get quarantine directory from config
	quarantineDir := h.config.Library.QuarantineDirectory
	if quarantineDir == "" {
		http.Error(w, "Quarantine directory not configured", http.StatusInternalServerError)
		return
	}

	// Load quarantine reasons from import logs
	quarantineReasons, err := h.loadQuarantineReasons()
	if err != nil {
		// Log error but don't fail - we can still show books without reasons
		fmt.Printf("Warning: failed to load quarantine reasons: %v\n", err)
	}

	// Scan quarantine directory for EPUB files
	var quarantineBooks []models.QuarantineBook
	err = filepath.Walk(quarantineDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Only process EPUB files
		if strings.ToLower(filepath.Ext(path)) != ".epub" {
			return nil
		}

		// Extract metadata from the EPUB file
		bookMetadata, err := h.extractMetadata(path)
		if err != nil {
			// Use filename as fallback
			bookMetadata = h.extractFromFilename(path)
		}

		// Create book entry
		book := models.QuarantineBook{
			Book: models.Book{
				ID:        0, // No database ID for quarantine books
				Title:     bookMetadata.Title,
				Author:    bookMetadata.Author,
				FilePath:  path,
				FileSize:  info.Size(),
				Format:    "epub",
				ISBN:      bookMetadata.ISBN,
				Publisher: bookMetadata.Publisher,
			},
		}

		// Look up quarantine reason for this file
		if reason, exists := quarantineReasons[path]; exists {
			book.QuarantineReason = reason.Reason
			book.QuarantineDetail = reason.ErrorDetail
			book.QuarantineDate = reason.Timestamp.Format("2006-01-02 15:04:05")
		}

		quarantineBooks = append(quarantineBooks, book)
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scan quarantine directory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quarantineBooks)
}

// ServeQuarantineCover serves cover images for quarantine books using the same logic as main library
func (h *BooksHandler) ServeQuarantineCover(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract file path from URL
	// URL format: /api/quarantine/covers/{filename}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid cover path", http.StatusBadRequest)
		return
	}

	filename := pathParts[len(pathParts)-1]
	
	// Find the quarantine book by filename
	var quarantineBook *models.QuarantineBook
	quarantineDir := h.config.Library.QuarantineDirectory
	
	err := filepath.Walk(quarantineDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".epub") {
			// Extract metadata from EPUB
			bookMetadata, err := h.extractMetadata(path)
			if err != nil {
				return nil // Skip files with metadata extraction errors
			}
			
			// Check if this is the book we're looking for
			baseName := filepath.Base(path)
			expectedName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + "_cover.jpg"
			if filename == expectedName {
				quarantineBook = &models.QuarantineBook{
					Book: models.Book{
						FilePath: path,
						Title:    bookMetadata.Title,
						Author:   bookMetadata.Author,
					},
				}
				return filepath.SkipDir // Stop walking
			}
		}
		return nil
	})
	
	if err != nil || quarantineBook == nil {
		http.Error(w, "Quarantine book not found", http.StatusNotFound)
		return
	}

	// Use the same cover extraction logic as the main library
	reader, err := zip.OpenReader(quarantineBook.FilePath)
	if err != nil {
		http.Error(w, "Failed to open EPUB file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Find cover image using the same logic as CoversHandler
	coverPath, err := h.findCoverInOPF(reader)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cover not found: %v", err), http.StatusNotFound)
		return
	}

	// Read cover image
	coverFile, err := reader.Open(coverPath)
	if err != nil {
		http.Error(w, "Failed to open cover image", http.StatusInternalServerError)
		return
	}
	defer coverFile.Close()

	// Read image data
	imageData, err := io.ReadAll(coverFile)
	if err != nil {
		http.Error(w, "Failed to read cover image", http.StatusInternalServerError)
		return
	}

	// Serve full image (no thumbnail generation for quarantine)
	contentType := http.DetectContentType(imageData)
	w.Header().Set("Content-Type", contentType)
	w.Write(imageData)
}

// findCoverInOPF finds the cover image path in the OPF file using XML parsing
// This is a copy of the logic from CoversHandler to avoid circular dependencies
func (h *BooksHandler) findCoverInOPF(reader *zip.ReadCloser) (string, error) {
	// Find the OPF file
	var opfPath string
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".opf") {
			opfPath = file.Name
			break
		}
	}

	if opfPath == "" {
		return "", fmt.Errorf("no OPF file found")
	}

	// Read and parse the OPF file
	opfFile, err := reader.Open(opfPath)
	if err != nil {
		return "", err
	}
	defer opfFile.Close()

	// Parse XML using Go's standard library
	opfData, err := io.ReadAll(opfFile)
	if err != nil {
		return "", err
	}

	// Define OPF structures locally to avoid conflicts
	type OPFDocument struct {
		XMLName  xml.Name `xml:"package"`
		Metadata struct {
			Meta []MetaTag `xml:"meta"`
		} `xml:"metadata"`
		Manifest struct {
			Items []ManifestItem `xml:"item"`
		} `xml:"manifest"`
	}

	type MetaTag struct {
		Name    string `xml:"name,attr"`
		Content string `xml:"content,attr"`
	}

	type ManifestItem struct {
		ID   string `xml:"id,attr"`
		Href string `xml:"href,attr"`
	}

	var opf OPFDocument
	if err := xml.Unmarshal(opfData, &opf); err != nil {
		return "", fmt.Errorf("failed to parse OPF XML: %v", err)
	}

	// Step 1: Find cover metadata
	var coverID string
	for _, meta := range opf.Metadata.Meta {
		if meta.Name == "cover" {
			coverID = meta.Content
			fmt.Printf("Found cover metadata: <meta name=\"cover\" content=\"%s\"/>\n", coverID)
			break
		}
	}

	if coverID == "" {
		// Fallback: look for direct cover references in manifest
		for _, item := range opf.Manifest.Items {
			if item.ID == "cover" || strings.Contains(item.ID, "cover") {
				fmt.Printf("Found direct cover reference: %s\n", item.Href)
				// Make path relative to OPF file location
				opfDir := filepath.Dir(opfPath)
				if opfDir != "." {
					return filepath.Join(opfDir, item.Href), nil
				}
				return item.Href, nil
			}
		}
		return "", fmt.Errorf("no cover metadata found in OPF")
	}

	// Step 2: Find manifest item by cover ID
	var coverPath string
	for _, item := range opf.Manifest.Items {
		if item.ID == coverID {
			coverPath = item.Href
			fmt.Printf("Found cover image in manifest: %s\n", coverPath)
			break
		}
	}

	if coverPath == "" {
		return "", fmt.Errorf("cover ID '%s' not found in manifest", coverID)
	}

	// Step 3: Make path relative to OPF file location
	opfDir := filepath.Dir(opfPath)
	if opfDir != "." {
		coverPath = filepath.Join(opfDir, coverPath)
	}

	fmt.Printf("Resolved cover path: %s\n", coverPath)
	return coverPath, nil
}

// SearchMetadata searches for book metadata using Open Library API
func (h *BooksHandler) SearchMetadata(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🚀 SearchMetadata API called\n")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var searchRequest models.MetadataSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		fmt.Printf("❌ JSON Decode Error: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("📝 Search Request:\n")
	fmt.Printf("   Title: '%s'\n", searchRequest.Title)
	fmt.Printf("   Author: '%s'\n", searchRequest.Author)

	// Validate required fields
	if searchRequest.Title == "" {
		fmt.Printf("❌ Validation Error: Title is required\n")
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Search Open Library
	fmt.Printf("🔍 Starting Open Library search...\n")
	suggestions, confidence, err := h.searchOpenLibrary(searchRequest.Title, searchRequest.Author)
	if err != nil {
		fmt.Printf("❌ Search Error: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to search metadata: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.MetadataSearchResponse{
		Suggestions: suggestions,
		Confidence:  confidence,
	}

	if len(suggestions) == 0 {
		response.Message = "No matching books found in Open Library"
		fmt.Printf("⚠️ No suggestions found\n")
	} else {
		fmt.Printf("✅ Returning %d suggestions with confidence %.2f\n", len(suggestions), confidence)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// normalizeSearchText cleans and normalizes text for search
func (h *BooksHandler) normalizeSearchText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove special characters, keep only letters and spaces
	re := regexp.MustCompile(`[^a-z\s]`)
	text = re.ReplaceAllString(text, " ")

	// Remove extra spaces
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// searchOpenLibrary searches for books using Open Library API
func (h *BooksHandler) searchOpenLibrary(title, author string) ([]models.MetadataSuggestion, float64, error) {
	// Normalize and combine search terms
	searchQuery := h.normalizeSearchText(title)
	if author != "" {
		searchQuery += " " + h.normalizeSearchText(author)
	}

	// Build URL with generic q parameter
	baseURL := "https://openlibrary.org/search.json"
	searchURL := fmt.Sprintf("%s?q=%s", baseURL, url.QueryEscape(searchQuery))

	// Debug logging
	fmt.Printf("🔍 Open Library Search Request:\n")
	fmt.Printf("   Original Title: '%s'\n", title)
	fmt.Printf("   Original Author: '%s'\n", author)
	fmt.Printf("   Normalized Query: '%s'\n", searchQuery)
	fmt.Printf("   URL: %s\n", searchURL)

	// Show normalization examples for debugging
	fmt.Printf("   📝 Normalization Examples:\n")
	fmt.Printf("      Title: '%s' -> '%s'\n", title, h.normalizeSearchText(title))
	if author != "" {
		fmt.Printf("      Author: '%s' -> '%s'\n", author, h.normalizeSearchText(author))
	}

	// Make HTTP request
	resp, err := http.Get(searchURL)
	if err != nil {
		fmt.Printf("❌ HTTP Request Error: %v\n", err)
		return nil, 0, fmt.Errorf("failed to query Open Library: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Open Library Response Status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Open Library Error Response: %s\n", string(body))
		return nil, 0, fmt.Errorf("Open Library API returned status %d", resp.StatusCode)
	}

	// Parse response
	var searchResponse struct {
		Docs []struct {
			Title            string   `json:"title"`
			AuthorName       []string `json:"author_name"`
			FirstPublishYear int      `json:"first_publish_year"`
			Key              string   `json:"key"`
			ISBN             []string `json:"isbn"`
			Publisher        []string `json:"publisher"`
		} `json:"docs"`
	}

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		return nil, 0, fmt.Errorf("failed to read Open Library response: %v", err)
	}

	// Show first 500 characters of response
	bodyPreview := body
	if len(body) > 500 {
		bodyPreview = body[:500]
	}
	fmt.Printf("📄 Open Library Response Body (first 500 chars):\n%s\n", string(bodyPreview))

	if err := json.Unmarshal(body, &searchResponse); err != nil {
		fmt.Printf("❌ JSON Parse Error: %v\n", err)
		fmt.Printf("❌ Raw Response: %s\n", string(body))
		return nil, 0, fmt.Errorf("failed to parse Open Library response: %v", err)
	}

	fmt.Printf("📚 Found %d documents in Open Library response\n", len(searchResponse.Docs))

	// Process results and calculate confidence scores
	var suggestions []models.MetadataSuggestion
	var totalConfidence float64

	for i, doc := range searchResponse.Docs {
		fmt.Printf("📖 Processing document %d:\n", i+1)
		fmt.Printf("   Title: '%s'\n", doc.Title)
		fmt.Printf("   Authors: %v\n", doc.AuthorName)
		fmt.Printf("   Year: %d\n", doc.FirstPublishYear)
		fmt.Printf("   Key: %s\n", doc.Key)

		if doc.Title == "" {
			fmt.Printf("   ⚠️ Skipping - no title\n")
			continue
		}

		// Get detailed information for this work
		fmt.Printf("   🔍 Fetching work details from: %s\n", doc.Key)
		workDetails, err := h.getWorkDetails(doc.Key)
		if err != nil {
			fmt.Printf("   ❌ Failed to get work details: %v\n", err)
			continue // Skip if we can't get details
		}

		fmt.Printf("   📋 Work details - ISBN: %v, Publisher: %v\n", workDetails.ISBN, workDetails.Publisher)

		// Calculate confidence score
		confidence := h.calculateConfidence(title, author, doc.Title, doc.AuthorName, workDetails)
		fmt.Printf("   🎯 Confidence score: %.2f\n", confidence)

		// Only include suggestions with reasonable confidence
		if confidence > 0.3 {
			suggestion := models.MetadataSuggestion{
				Title:      doc.Title,
				Author:     strings.Join(doc.AuthorName, ", "),
				ISBN:       h.extractBestISBN(workDetails.ISBN),
				Publisher:  h.extractBestPublisher(workDetails.Publisher),
				Year:       doc.FirstPublishYear,
				Confidence: confidence,
				Source:     "Open Library",
			}
			suggestions = append(suggestions, suggestion)
			totalConfidence += confidence
			fmt.Printf("   ✅ Added to suggestions\n")
		} else {
			fmt.Printf("   ⚠️ Skipping - confidence too low (%.2f < 0.3)\n", confidence)
		}
	}

	// Sort by confidence (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	// Limit to top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	// Calculate average confidence
	avgConfidence := 0.0
	if len(suggestions) > 0 {
		avgConfidence = totalConfidence / float64(len(suggestions))
	}

	fmt.Printf("🎯 Final Results:\n")
	fmt.Printf("   Total suggestions: %d\n", len(suggestions))
	fmt.Printf("   Average confidence: %.2f\n", avgConfidence)
	for i, suggestion := range suggestions {
		fmt.Printf("   Suggestion %d: '%s' by %s (confidence: %.2f)\n",
			i+1, suggestion.Title, suggestion.Author, suggestion.Confidence)
	}

	return suggestions, avgConfidence, nil
}

// WorkDetails represents detailed work information from Open Library
type WorkDetails struct {
	ISBN      []string `json:"isbn"`
	Publisher []string `json:"publisher"`
}

// getWorkDetails fetches detailed information for a work
func (h *BooksHandler) getWorkDetails(workKey string) (*WorkDetails, error) {
	workURL := "https://openlibrary.org" + workKey + ".json"

	fmt.Printf("      🔗 Fetching work details from: %s\n", workURL)

	resp, err := http.Get(workURL)
	if err != nil {
		fmt.Printf("      ❌ HTTP Error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("      📡 Work API Status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("      ❌ Work API Error Response: %s\n", string(body))
		return nil, fmt.Errorf("work API returned status %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("      ❌ Failed to read work response: %v\n", err)
		return nil, err
	}

	var details WorkDetails
	if err := json.Unmarshal(body, &details); err != nil {
		fmt.Printf("      ❌ Failed to parse work response: %v\n", err)
		// Show first 200 characters
		bodyPreview := body
		if len(body) > 200 {
			bodyPreview = body[:200]
		}
		fmt.Printf("      📄 Work response (first 200 chars): %s\n", string(bodyPreview))
		return nil, err
	}

	fmt.Printf("      ✅ Work details parsed successfully\n")
	return &details, nil
}

// calculateConfidence calculates how confident we are in a match
func (h *BooksHandler) calculateConfidence(searchTitle, searchAuthor, resultTitle string, resultAuthors []string, details *WorkDetails) float64 {
	confidence := 0.0

	// Title matching (50% weight) - more important with generic search
	titleScore := h.calculateStringSimilarity(searchTitle, resultTitle)
	confidence += titleScore * 0.5

	// Author matching (30% weight)
	if searchAuthor != "" && len(resultAuthors) > 0 {
		authorScore := 0.0
		for _, author := range resultAuthors {
			similarity := h.calculateStringSimilarity(searchAuthor, author)
			if similarity > authorScore {
				authorScore = similarity
			}
		}
		confidence += authorScore * 0.3
	}

	// Data completeness (20% weight)
	completenessScore := 0.0
	if len(details.ISBN) > 0 {
		completenessScore += 0.3
	}
	if len(details.Publisher) > 0 {
		completenessScore += 0.2
	}
	confidence += completenessScore * 0.2

	return confidence
}

// calculateStringSimilarity calculates similarity between two strings (0.0 to 1.0)
func (h *BooksHandler) calculateStringSimilarity(s1, s2 string) float64 {
	// Normalize both strings for comparison
	s1 = h.normalizeSearchText(s1)
	s2 = h.normalizeSearchText(s2)

	if s1 == s2 {
		return 1.0
	}

	// Simple similarity based on common words
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	commonWords := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				commonWords++
				break
			}
		}
	}

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Calculate similarity as ratio of common words to total unique words
	totalWords := len(words1) + len(words2) - commonWords
	if totalWords == 0 {
		return 0.0
	}

	return float64(commonWords) / float64(totalWords)
}

// extractBestISBN extracts the best ISBN from a list
func (h *BooksHandler) extractBestISBN(isbns []string) string {
	if len(isbns) == 0 {
		return ""
	}

	// Prefer ISBN-13 over ISBN-10
	for _, isbn := range isbns {
		if len(isbn) == 13 {
			return isbn
		}
	}

	// Fallback to first ISBN
	return isbns[0]
}

// extractBestPublisher extracts the best publisher from a list
func (h *BooksHandler) extractBestPublisher(publishers []string) string {
	if len(publishers) == 0 {
		return ""
	}
	return publishers[0]
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// loadQuarantineReasons loads quarantine reasons from import logs
func (h *BooksHandler) loadQuarantineReasons() (map[string]QuarantineReason, error) {
	quarantineReasons := make(map[string]QuarantineReason)

	// Get log directory from config
	logDir := h.config.LogDir
	if logDir == "" {
		return quarantineReasons, nil // No log directory configured
	}

	// Read all log files
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return quarantineReasons, err
	}

	// Process log files in reverse chronological order (newest first)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			logPath := filepath.Join(logDir, file.Name())
			data, err := ioutil.ReadFile(logPath)
			if err != nil {
				continue // Skip corrupted files
			}

			// Parse the log file to extract quarantined books
			var logData map[string]interface{}
			if err := json.Unmarshal(data, &logData); err != nil {
				continue // Skip invalid JSON
			}

			// Check if this log has quarantined books
			if quarantinedBooks, exists := logData["quarantined_books"]; exists {
				if books, ok := quarantinedBooks.([]interface{}); ok {
					for _, book := range books {
						if bookMap, ok := book.(map[string]interface{}); ok {
							quarantinePath, _ := bookMap["quarantine_path"].(string)
							reason, _ := bookMap["reason"].(string)
							errorDetail, _ := bookMap["error_detail"].(string)
							timestampStr, _ := bookMap["timestamp"].(string)

							if quarantinePath != "" && reason != "" {
								// Parse timestamp
								timestamp, _ := time.Parse(time.RFC3339, timestampStr)

								// Only add if we don't already have a reason for this file
								if _, exists := quarantineReasons[quarantinePath]; !exists {
									quarantineReasons[quarantinePath] = QuarantineReason{
										Reason:      reason,
										ErrorDetail: errorDetail,
										Timestamp:   timestamp,
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return quarantineReasons, nil
}

// QuarantineReason represents quarantine information for a book
type QuarantineReason struct {
	Reason      string    `json:"reason"`
	ErrorDetail string    `json:"error_detail"`
	Timestamp   time.Time `json:"timestamp"`
}

// extractMetadata extracts metadata from an EPUB file
func (h *BooksHandler) extractMetadata(filePath string) (models.BookRequest, error) {
	// This is a simplified version - in a real implementation, you'd use the metadata extractor
	// For now, return basic info from filename
	return h.extractFromFilename(filePath), nil
}

// extractFromFilename extracts basic metadata from filename
func (h *BooksHandler) extractFromFilename(filePath string) models.BookRequest {
	filename := filepath.Base(filePath)
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try to parse "Title - Author" format
	parts := strings.Split(name, " - ")
	if len(parts) >= 2 {
		return models.BookRequest{
			Title:  strings.TrimSpace(parts[0]),
			Author: strings.TrimSpace(parts[1]),
		}
	}

	// Fallback to filename as title
	return models.BookRequest{
		Title:  name,
		Author: "Unknown",
	}
}

// EditQuarantineBook handles editing metadata for quarantine books
func (h *BooksHandler) EditQuarantineBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var editRequest struct {
		FilePath  string `json:"file_path"`
		Title     string `json:"title"`
		Author    string `json:"author"`
		ISBN      string `json:"isbn"`
		Publisher string `json:"publisher"`
	}

	if err := json.NewDecoder(r.Body).Decode(&editRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if editRequest.Title == "" || editRequest.Author == "" {
		http.Error(w, "Title and author are required", http.StatusBadRequest)
		return
	}

	// Check if file exists in quarantine
	if _, err := os.Stat(editRequest.FilePath); os.IsNotExist(err) {
		http.Error(w, "Quarantine file not found", http.StatusNotFound)
		return
	}

	// Generate new file path in scan directory
	newFilePath := h.generateNewFilePath(editRequest.Author, editRequest.Title, "epub")

	// Create the new directory structure
	newDir := filepath.Dir(newFilePath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Move file from quarantine to scan directory
	if err := os.Rename(editRequest.FilePath, newFilePath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to move file: %v", err), http.StatusInternalServerError)
		return
	}

	// Get file info for database
	fileInfo, err := os.Stat(newFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file info: %v", err), http.StatusInternalServerError)
		return
	}

	// Add book to database
	book := models.BookRequest{
		Title:     editRequest.Title,
		Author:    editRequest.Author,
		FilePath:  newFilePath,
		FileSize:  fileInfo.Size(),
		Format:    "epub",
		ISBN:      editRequest.ISBN,
		Publisher: editRequest.Publisher,
	}

	if err := h.db.AddBook(book); err != nil {
		// If database add fails, try to move file back to quarantine
		os.Rename(newFilePath, editRequest.FilePath)
		http.Error(w, fmt.Sprintf("Failed to add book to database: %v", err), http.StatusInternalServerError)
		return
	}

	// Clean up empty quarantine directories
	if err := h.cleanupEmptyDirectories(filepath.Dir(editRequest.FilePath)); err != nil {
		// Log warning but don't fail the operation
		fmt.Printf("Warning: failed to cleanup quarantine directories: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Quarantine book processed successfully",
	})
}

// cleanupEmptyDirectories recursively removes empty directories
func (h *BooksHandler) cleanupEmptyDirectories(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", dirPath, err)
	}

	// If directory is not empty, don't remove it
	if len(entries) > 0 {
		return nil
	}

	// Directory is empty, remove it
	if err := os.Remove(dirPath); err != nil {
		return fmt.Errorf("failed to remove empty directory %s: %v", dirPath, err)
	}

	// Move up one level and check parent directory
	parentDir := filepath.Dir(dirPath)

	// Don't go above the scan directory
	scanDir := h.config.Library.ScanDirectory
	if parentDir == scanDir || parentDir == filepath.Dir(scanDir) {
		return nil // Stop at scan directory level
	}

	// Recursively check parent directory
	return h.cleanupEmptyDirectories(parentDir)
}

// GetLibraryStats returns library statistics
func (h *BooksHandler) GetLibraryStats(w http.ResponseWriter, r *http.Request) {
	// Get total books count
	totalBooks, err := h.db.GetTotalBooksCount()
	if err != nil {
		http.Error(w, "Failed to get total books count", http.StatusInternalServerError)
		return
	}

	// Get quarantine books count
	quarantineBooks, err := h.getQuarantineBooksCount()
	if err != nil {
		log.Printf("Error getting quarantine books count: %v", err)
		http.Error(w, "Failed to get quarantine books count", http.StatusInternalServerError)
		return
	}

	// Get total authors count
	totalAuthors, err := h.db.GetTotalAuthorsCount()
	if err != nil {
		http.Error(w, "Failed to get total authors count", http.StatusInternalServerError)
		return
	}

	// Get total publishers count
	totalPublishers, err := h.db.GetTotalPublishersCount()
	if err != nil {
		http.Error(w, "Failed to get total publishers count", http.StatusInternalServerError)
		return
	}

	// Get library size information
	log.Printf("Calling GetLibrarySizeInfo...")
	totalSize, avgSize, err := h.db.GetLibrarySizeInfo()
	if err != nil {
		log.Printf("Error getting library size info: %v", err)
		http.Error(w, "Failed to get library size info", http.StatusInternalServerError)
		return
	}
	log.Printf("GetLibrarySizeInfo successful: total=%d, avg=%d", totalSize, avgSize)

	// Get last activity dates
	lastImport, lastScan, err := h.db.GetLastActivityDates()
	if err != nil {
		http.Error(w, "Failed to get last activity dates", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"total_books":      totalBooks,
		"quarantine_books": quarantineBooks,
		"total_authors":    totalAuthors,
		"total_publishers": totalPublishers,
		"total_size":       formatFileSize(totalSize),
		"avg_book_size":    formatFileSize(avgSize),
		"last_import":      lastImport,
		"last_scan":        lastScan,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getQuarantineBooksCount returns the number of books in quarantine directory
func (h *BooksHandler) getQuarantineBooksCount() (int, error) {
	// Get quarantine directory from config
	quarantineDir := h.config.Library.QuarantineDirectory
	if quarantineDir == "" {
		return 0, fmt.Errorf("quarantine directory not configured")
	}

	// Check if quarantine directory exists
	if _, err := os.Stat(quarantineDir); os.IsNotExist(err) {
		return 0, nil // Directory doesn't exist, so no quarantine books
	}

	// Count EPUB files in quarantine directory
	count := 0
	err := filepath.Walk(quarantineDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".epub" {
			count++
		}
		return nil
	})

	return count, err
}

// formatFileSize formats file size in bytes to human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
