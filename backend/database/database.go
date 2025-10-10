package database

import (
	"database/sql"
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
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := dm.db.Exec(query)
	return err
}

// GetAllBooks returns all books from the database
func (dm *Manager) GetAllBooks() ([]models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, added_at FROM books ORDER BY title"
	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.AddedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// SearchBooks searches for books by title or author
func (dm *Manager) SearchBooks(query string) ([]models.Book, error) {
	searchQuery := `SELECT id, title, author, file_path, file_size, format, added_at 
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
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.AddedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// AddBook adds a new book to the database
func (dm *Manager) AddBook(book models.BookRequest) error {
	query := `INSERT INTO books (title, author, file_path, file_size, format, added_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := dm.db.Exec(query, book.Title, book.Author, book.FilePath, book.FileSize, book.Format, time.Now())
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
		".pdf":  true,
		".mobi": true,
		".azw":  true,
		".azw3": true,
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

		book := models.BookRequest{
			Title:    title,
			Author:   author,
			FilePath: path,
			FileSize: info.Size(),
			Format:   strings.TrimPrefix(ext, "."),
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
		".pdf":  true,
		".mobi": true,
		".azw":  true,
		".azw3": true,
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

		book := models.BookRequest{
			Title:    title,
			Author:   author,
			FilePath: path,
			FileSize: info.Size(),
			Format:   strings.TrimPrefix(ext, "."),
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
	query := "SELECT id, title, author, file_path, file_size, format, added_at FROM books WHERE author = ? ORDER BY title"
	rows, err := dm.db.Query(query, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.AddedAt)
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
	query := "SELECT id, title, author, file_path, file_size, format, added_at FROM books WHERE title = ? ORDER BY author"
	rows, err := dm.db.Query(query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.AddedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetBookByID returns a book by its ID
func (dm *Manager) GetBookByID(id int) (models.Book, error) {
	query := "SELECT id, title, author, file_path, file_size, format, added_at FROM books WHERE id = ?"
	row := dm.db.QueryRow(query, id)

	var book models.Book
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath, &book.FileSize, &book.Format, &book.AddedAt)
	if err != nil {
		return models.Book{}, err
	}

	return book, nil
}
