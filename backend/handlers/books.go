package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"fableflow/backend/database"
	"fableflow/backend/models"
)

// BooksHandler handles book-related HTTP requests
type BooksHandler struct {
	db *database.Manager
}

// NewBooksHandler creates a new books handler
func NewBooksHandler(db *database.Manager) *BooksHandler {
	return &BooksHandler{db: db}
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
	// Extract ID from URL path (assuming /api/books/{id})
	// This is a simplified version - in a real app you'd use a router
	idStr := r.URL.Path[len("/api/books/"):]
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

	// Extract ID from URL path
	idStr := r.URL.Path[len("/api/download/"):]
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

	// Set headers for file download
	filename := filepath.Base(book.FilePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

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
