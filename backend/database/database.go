package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fableflow/backend/metadata"
	"fableflow/backend/models"

	_ "github.com/mattn/go-sqlite3"
)

// Manager handles all database operations
type Manager struct {
	db        *sql.DB
	extractor *metadata.Extractor
}

// NewManager creates a new database manager
func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	dm := &Manager{
		db:        db,
		extractor: metadata.NewExtractor(),
	}
	err = dm.initDatabase()
	if err != nil {
		return nil, err
	}

	return dm, nil
}

// Close closes the database connection
func (dm *Manager) Close() error {
	return dm.db.Close()
}

// initDatabase creates the books table if it doesn't exist
func (dm *Manager) initDatabase() error {
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT,
		file_path TEXT UNIQUE NOT NULL,
		file_size INTEGER,
		format TEXT,
		isbn TEXT,
		publisher TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := dm.db.Exec(query)
	if err != nil {
		return err
	}

	// Add publisher column if it doesn't exist (migration)
	_, err = dm.db.Exec(`ALTER TABLE books ADD COLUMN publisher TEXT;`)
	if err != nil {
		// Column might already exist, ignore the error
		// In a production app, you'd check if the column exists first
	}

	// Add updated_at column if it doesn't exist (migration)
	_, err = dm.db.Exec(`ALTER TABLE books ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;`)
	if err != nil {
		// Column might already exist, ignore the error
		// In a production app, you'd check if the column exists first
	}

	return nil
}

// GetAllBooks returns all books from the database
func (dm *Manager) GetAllBooks() ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books ORDER BY title"
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// SearchBooks searches for books by title or author
func (dm *Manager) SearchBooks(query string) ([]models.Book, error) {
	searchQuery := `SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at 
					FROM books 
					WHERE title LIKE ? OR author LIKE ? 
					ORDER BY title`
	searchTerm := "%" + query + "%"

	rows, err := dm.db.Query(searchQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// AddBook adds a new book to the database
func (dm *Manager) AddBook(book models.BookRequest) error {
	query := `INSERT INTO books (title, author, file_path, file_size, format, isbn, publisher, added_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := dm.db.Exec(query, book.Title, book.Author, book.FilePath, book.FileSize, book.Format, book.ISBN, book.Publisher, time.Now())
	return err
}

// RemoveBook removes a book from the database by ID
func (dm *Manager) RemoveBook(bookID int) error {
	query := `DELETE FROM books WHERE id = ?`
	_, err := dm.db.Exec(query, bookID)
	return err
}

// BookExists checks if a book with the given file path already exists
func (dm *Manager) BookExists(filePath string) (bool, error) {
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM books WHERE file_path = ?", filePath).Scan(&count)
	return count > 0, err
}

// ScanDirectory recursively scans a directory for ebook files
func (dm *Manager) ScanDirectory(rootPath string) error {
	supportedFormats := map[string]bool{
		".epub": true,
		// Only scan for EPUB files to avoid importing converted files
	}

	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedFormats[ext] {
			return nil // Skip unsupported files
		}

		// Check if book already exists in database
		exists, err := dm.BookExists(path)
		if err != nil || exists {
			return nil
		}

		// Extract metadata from the ebook file
		bookMetadata, err := dm.extractor.ExtractMetadata(path)
		if err != nil {
			log.Printf("Failed to extract metadata from %s: %v", path, err)
			// Fallback to filename parsing
			bookMetadata = dm.extractor.ExtractFromFilename(path)
		}

		title := bookMetadata.Title
		author := bookMetadata.Author
		isbn := bookMetadata.ISBN

		book := models.BookRequest{
			Title:     title,
			Author:    author,
			FilePath:  path,
			FileSize:  info.Size(),
			Format:    strings.TrimPrefix(ext, "."),
			ISBN:      isbn,
			Publisher: bookMetadata.Publisher,
		}

		err = dm.AddBook(book)
		if err != nil {
			log.Printf("Error adding book %s: %v", path, err)
		} else {
			log.Printf("Added book: %s by %s", title, author)
		}

		return nil
	})
}

// RescanDirectory performs a rescan that adds new books and removes unavailable ones
func (dm *Manager) RescanDirectory(rootPath string) (int, int, error) {
	supportedFormats := map[string]bool{
		".epub": true,
		// Only scan for EPUB files to avoid importing converted files
	}

	// Get all current books from database
	currentBooks, err := dm.GetAllBooks()
	if err != nil {
		return 0, 0, err
	}

	// Create a map of current file paths for quick lookup
	currentPaths := make(map[string]bool)
	for _, book := range currentBooks {
		currentPaths[book.FilePath] = true
	}

	// Track files found during scan
	foundPaths := make(map[string]bool)
	added := 0
	removed := 0

	// Scan directory for new books
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedFormats[ext] {
			return nil // Skip unsupported files
		}

		foundPaths[path] = true

		// Check if book already exists in database
		exists, err := dm.BookExists(path)
		if err != nil || exists {
			return nil
		}

		// Extract metadata from the ebook file
		bookMetadata, err := dm.extractor.ExtractMetadata(path)
		if err != nil {
			log.Printf("Failed to extract metadata from %s: %v", path, err)
			// Fallback to filename parsing
			bookMetadata = dm.extractor.ExtractFromFilename(path)
		}

		title := bookMetadata.Title
		author := bookMetadata.Author
		isbn := bookMetadata.ISBN

		book := models.BookRequest{
			Title:     title,
			Author:    author,
			FilePath:  path,
			FileSize:  info.Size(),
			Format:    strings.TrimPrefix(ext, "."),
			ISBN:      isbn,
			Publisher: bookMetadata.Publisher,
		}

		err = dm.AddBook(book)
		if err != nil {
			log.Printf("Error adding book %s: %v", path, err)
		} else {
			log.Printf("Added book: %s by %s", title, author)
			added++
		}

		return nil
	})

	if err != nil {
		return added, removed, err
	}

	// Remove books that are no longer available
	for _, book := range currentBooks {
		if !foundPaths[book.FilePath] {
			err := dm.RemoveBook(book.ID)
			if err != nil {
				log.Printf("Error removing book %s: %v", book.FilePath, err)
			} else {
				log.Printf("Removed book: %s by %s", book.Title, book.Author)
				removed++
			}
		}
	}

	return added, removed, nil
}

// GetAllAuthors returns all unique authors
func (dm *Manager) GetAllAuthors() ([]string, error) {
	query := "SELECT DISTINCT author FROM books ORDER BY author"
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var author string
		err := rows.Scan(&author)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// GetAuthorsByLetter returns authors starting with a specific letter
func (dm *Manager) GetAuthorsByLetter(letter string) ([]string, error) {
	query := "SELECT DISTINCT author FROM books WHERE author LIKE ? ORDER BY author"
	searchTerm := letter + "%"

	rows, err := dm.db.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var author string
		err := rows.Scan(&author)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// GetBooksByAuthor returns all books by a specific author
func (dm *Manager) GetBooksByAuthor(author string) ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books WHERE author = ? ORDER BY title"
	rows, err := dm.db.Query(query, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetAllTitles returns all unique titles
func (dm *Manager) GetAllTitles() ([]string, error) {
	query := "SELECT DISTINCT title FROM books ORDER BY title"
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetTitlesByLetter returns titles starting with a specific letter
func (dm *Manager) GetTitlesByLetter(letter string) ([]string, error) {
	query := "SELECT DISTINCT title FROM books WHERE title LIKE ? ORDER BY title"
	searchTerm := letter + "%"

	rows, err := dm.db.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}

	return titles, nil
}

// GetBooksByTitle returns all books with a specific title
func (dm *Manager) GetBooksByTitle(title string) ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books WHERE title = ? ORDER BY author"
	rows, err := dm.db.Query(query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetRecentBooks returns the most recently added books
func (dm *Manager) GetRecentBooks(limit int) ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books ORDER BY added_at DESC LIMIT ?"
	rows, err := dm.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetRandomBooks returns a random selection of books
func (dm *Manager) GetRandomBooks(limit int) ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books ORDER BY RANDOM() LIMIT ?"
	rows, err := dm.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetBookByID returns a book by its ID
func (dm *Manager) GetBookByID(id int) (models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, isbn, publisher, added_at, updated_at FROM books WHERE id = ?"
	row := dm.db.QueryRow(query, id)

	var book models.Book
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.ISBN, &book.Publisher, &book.AddedAt, &book.UpdatedAt)
	if err != nil {
		return models.Book{}, err
	}

	return book, nil
}

// UpdateBook updates book metadata in the database
func (m *Manager) UpdateBook(id int, title, author, isbn, publisher string) error {
	query := `
		UPDATE books 
		SET title = ?, author = ?, isbn = ?, publisher = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`

	_, err := m.db.Exec(query, title, author, isbn, publisher, id)
	if err != nil {
		return fmt.Errorf("failed to update book: %v", err)
	}

	return nil
}

// UpdateBookWithPath updates book metadata and file path in the database
func (m *Manager) UpdateBookWithPath(id int, title, author, isbn, publisher, filePath string) error {
	query := `
		UPDATE books 
		SET title = ?, author = ?, isbn = ?, publisher = ?, file_path = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`

	_, err := m.db.Exec(query, title, author, isbn, publisher, filePath, id)
	if err != nil {
		return fmt.Errorf("failed to update book: %v", err)
	}

	return nil
}

// GetTotalBooksCount returns the total number of books in the library
func (m *Manager) GetTotalBooksCount() (int, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	return count, err
}

// GetQuarantineBooksCount returns the number of books in quarantine
func (m *Manager) GetQuarantineBooksCount() (int, error) {
	// This method should be called from the BooksHandler since it needs access to config
	// For now, return 0 as a placeholder - the actual implementation is in BooksHandler.GetQuarantineBooks
	return 0, nil
}

// GetTotalAuthorsCount returns the number of unique authors
func (m *Manager) GetTotalAuthorsCount() (int, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(DISTINCT author) FROM books").Scan(&count)
	return count, err
}

// GetTotalPublishersCount returns the number of unique publishers
func (m *Manager) GetTotalPublishersCount() (int, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(DISTINCT publisher) FROM books WHERE publisher != ''").Scan(&count)
	return count, err
}

// GetLibrarySizeInfo returns total size and average book size
func (m *Manager) GetLibrarySizeInfo() (int64, int64, error) {
	var totalSize sql.NullInt64
	var avgSize sql.NullFloat64
	err := m.db.QueryRow("SELECT COALESCE(SUM(file_size), 0), COALESCE(AVG(file_size), 0) FROM books WHERE file_size IS NOT NULL").Scan(&totalSize, &avgSize)
	if err != nil {
		return 0, 0, err
	}

	// Convert NullInt64 to int64, defaulting to 0 if NULL
	total := int64(0)
	avg := int64(0)
	if totalSize.Valid {
		total = totalSize.Int64
	}
	if avgSize.Valid {
		avg = int64(avgSize.Float64) // Convert float64 to int64
	}

	return total, avg, nil
}

// GetLastActivityDates returns the last import and scan dates
func (m *Manager) GetLastActivityDates() (string, string, error) {
	var lastImport, lastScan sql.NullString

	// Get the most recent added_at date as last scan
	err := m.db.QueryRow("SELECT MAX(added_at) FROM books").Scan(&lastScan)
	if err != nil {
		return "Never", "Never", nil // Return default values instead of error
	}

	// For now, use the same date for both (you can implement separate tracking later)
	if lastScan.Valid {
		lastImport = lastScan
	} else {
		lastImport = sql.NullString{String: "Never", Valid: true}
	}

	// Convert to strings, defaulting to "Never" if NULL
	lastImportStr := "Never"
	lastScanStr := "Never"
	if lastImport.Valid {
		lastImportStr = lastImport.String
	}
	if lastScan.Valid {
		lastScanStr = lastScan.String
	}

	return lastImportStr, lastScanStr, nil
}
