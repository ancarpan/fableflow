# FableFlow Backend - Agents Guide

This document provides detailed guidance for AI agents working on the Go backend of FableFlow.

## Backend Architecture Overview

The FableFlow backend is a Go HTTP server that provides RESTful API endpoints for the frontend application. It handles:
- Book library management and persistence
- EPUB file processing and metadata extraction
- Format conversion (EPUB to Kindle)
- Import pipeline for new books
- Directory scanning and file monitoring

### Technology Stack

- **Language**: Go 1.21+
- **Database**: SQLite (file-based, no external DB required)
- **HTTP**: Standard `net/http` package (no external frameworks like Gin)
- **EPUB Processing**: Custom parser + epub.js for reader
- **Format Conversion**: kindlegen (binary tool, platform-specific)

## Directory Structure

```
backend/
├── config/              # Configuration loading
│   └── config.go        # Parse YAML config files
├── database/            # Database layer
│   └── database.go      # SQLite operations, queries, schema
├── handlers/            # HTTP request handlers
│   ├── books.go         # Main book operations (largest file)
│   ├── scan.go          # Directory scanning
│   ├── import.go        # Import service integration
│   ├── conversion.go    # Format conversion endpoints
│   ├── covers.go        # Book cover serving
│   └── health.go        # Health check endpoint
├── models/              # Data structures
│   └── book.go          # Book, QuarantineBook, request/response types
├── metadata/            # Metadata extraction
│   └── extractor.go     # Extract metadata from EPUB
├── epub/                # EPUB manipulation
│   └── editor.go        # Read/modify EPUB structure
├── conversion/          # Format conversion logic
│   ├── converter.go     # Main conversion coordinator
│   ├── kindlegen_converter.go   # Kindle format conversion
│   └── epub_parser.go   # EPUB parsing for conversion
├── importservice/       # Book import pipeline
│   └── service.go       # Import coordination and logging
├── pathmanager/         # Path utilities
│   ├── pathmanager.go   # Safe path operations
│   └── pathmanager_test.go   # Path validation tests
├── kindlegen/           # kindlegen binaries (platform-specific)
│   ├── darwin/          # macOS binary
│   ├── linux/           # Linux binary
│   └── windows/         # Windows binary
├── main.go              # Server entry point and routing
├── Makefile             # Build commands
├── config.dev.yaml      # Development configuration
└── go.mod / go.sum      # Dependency management
```

## Core Packages

### 1. config (Configuration)

**Purpose**: Load and manage application configuration from YAML files.

**Key Types**:
- `Config` - Main configuration struct
- `ServerConfig` - HTTP server settings
- `DatabaseConfig` - SQLite database path
- `LibraryConfig` - Directories for ebooks, imports, quarantine

**Common Usage**:
```go
cfg, err := config.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}
path := cfg.Library.ScanDirectory  // Access config values
```

**When to Modify**:
- Add new configuration options that need to be user-configurable
- Update template: `../config.yaml.template`

### 2. database (Data Persistence)

**Purpose**: All SQLite database operations and queries.

**Key Type**: `Manager`
- Methods: `SaveBook()`, `GetBook()`, `GetAllBooks()`, `DeleteBook()`, `ScanDirectory()`, etc.
- Manages both the main library and quarantine database

**Key Methods**:
- `SaveBook(book *models.Book) error` - Insert/update book
- `GetBook(id int) (*models.Book, error)` - Fetch single book
- `GetAllBooks() ([]models.Book, error)` - Get all books
- `ScanDirectory(path string) error` - Scan for EPUB files and update DB
- `GetQuarantineBooks() ([]models.QuarantineBook, error)` - Quarantine items
- `SearchBooks(query string) ([]models.Book, error)` - Full-text search

**Important Patterns**:
```go
// Use prepared statements to prevent SQL injection
const selectBookQuery = "SELECT id, title, author FROM books WHERE id = ?"
stmt, err := db.db.Prepare(selectBookQuery)
if err != nil {
    return nil, err
}
defer stmt.Close()

rows, err := stmt.Query(id)
// ... handle rows
```

**When to Modify**:
- Add new database queries for new features
- Modify the SQLite schema (be careful with migrations!)
- Add new indices for performance
- Change how data is persisted

### 3. models (Data Structures)

**Purpose**: Define request/response types and domain models.

**Key Types**:
```go
type Book struct {
    ID, Title, Author, FilePath, Format, ISBN, Publisher, AddedAt, UpdatedAt
}

type QuarantineBook struct {
    Book
    QuarantineReason, QuarantineDetail, QuarantineDate, FileInfo
}

type MetadataSearchRequest struct {
    Title, Author
}

type MetadataSearchResponse struct {
    Suggestions []MetadataSuggestion
    Confidence float64
}
```

**When to Add Types**:
- Creating new API endpoints
- Changing request/response format
- Adding new domain concepts

### 4. handlers (HTTP Endpoints)

**Purpose**: Handle HTTP requests and return responses.

**File Organization**:
- `books.go` - Main book endpoints (large file, ~1000+ lines)
- `scan.go` - Directory scanning endpoints
- `import.go` - Import service integration
- `conversion.go` - Format conversion endpoints
- `covers.go` - Cover image serving
- `health.go` - Health check

**Handler Pattern**:
```go
type BooksHandler struct {
    db *database.Manager
    cfg *config.Config
}

func NewBooksHandler(db *database.Manager, cfg *config.Config) *BooksHandler {
    return &BooksHandler{db, cfg}
}

func (h *BooksHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
    // Fetch from database
    books, err := h.db.GetAllBooks()
    if err != nil {
        http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
        return
    }
    
    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(books)
}
```

**Main Endpoints** (see `main.go` for full list):
- `GET /api/books` - All books
- `GET /api/books/{id}` - Single book
- `GET /api/books/recent` - Recently added
- `GET /api/search` - Full-text search
- `GET /api/authors` - All authors
- `GET /api/authors/letter` - Authors by letter
- `POST /api/books/search-metadata` - Metadata lookup
- `POST /api/books/lookup-isbn` - ISBN lookup
- `GET /api/quarantine` - Quarantined books
- `POST /api/quarantine/edit` - Edit quarantine book
- `POST /api/scan` - Trigger directory scan
- `POST /api/convert` - Start format conversion
- `GET /api/import/start` - Start import
- And more...

**When to Modify**:
- Fix bugs in endpoint logic
- Add new endpoints
- Improve response format
- Add input validation

### 5. metadata (Book Information Extraction)

**Purpose**: Extract metadata (title, author, ISBN) from EPUB files.

**Key Type**: 
- `Extractor` - Extract metadata from EPUB files
- External API calls to book metadata services

**Main Function**: `ExtractMetadata(epubPath string) (*models.MetadataSuggestion, error)`

**When to Modify**:
- Add support for additional metadata sources
- Improve extraction accuracy
- Fix bugs in parsing

### 6. epub (EPUB File Manipulation)

**Purpose**: Read and modify EPUB file structure and metadata.

**Key Types**:
- `Editor` - Manipulate EPUB files

**Common Operations**:
- Extract cover images
- Modify metadata
- Parse table of contents
- Edit content files

**When to Modify**:
- Add new EPUB manipulation features
- Fix EPUB parsing issues
- Add support for EPUB variant handling

### 7. conversion (Format Conversion)

**Purpose**: Convert EPUB files to other formats (primarily Kindle).

**Key Files**:
- `converter.go` - Main conversion coordinator
- `kindlegen_converter.go` - Kindle conversion via kindlegen binary
- `epub_parser.go` - Parse EPUB for conversion

**Flow**:
1. User requests conversion in frontend
2. Backend starts async conversion process
3. Frontend polls `/api/convert/status` for progress
4. When complete, frontend downloads converted file

**Important Notes**:
- Uses kindlegen binary (must be in correct platform directory)
- Conversion is async; returns task ID immediately
- Results stored in temp directory with cleanup

**When to Modify**:
- Add support for new output formats
- Improve conversion error handling
- Optimize conversion performance

### 8. importservice (Import Pipeline)

**Purpose**: Manage the import workflow for new ebooks.

**Key Type**: `ImportService`
- Monitors import directory
- Validates EPUB files
- Extracts metadata
- Quarantines problematic files
- Logs all import operations

**Flow**:
1. Books placed in import directory
2. Import service processes them
3. Valid books → main library
4. Invalid books → quarantine with error details
5. Operation logged for user review

**Callback**: After import completes, database scan is triggered

**When to Modify**:
- Change validation rules
- Add preprocessing steps
- Improve error messages for quarantine

### 9. pathmanager (Path Safety)

**Purpose**: Validate and manage file paths safely.

**Key Functions**:
- `ValidatePath(path string) error` - Check path safety
- Path joining and normalization
- Prevention of directory traversal attacks

**Important**: Always use pathmanager for user-provided paths!

**Example**:
```go
// Bad - vulnerable to directory traversal
filePath := filepath.Join(cfg.ScanDir, userInput)

// Good - validated
if err := pathmanager.ValidatePath(filePath); err != nil {
    http.Error(w, "Invalid path", http.StatusBadRequest)
    return
}
```

## Common Development Tasks

### Adding a New API Endpoint

1. **Create handler method** in appropriate handler file:
```go
func (h *BooksHandler) NewEndpoint(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

2. **Register route** in `main.go`:
```go
http.HandleFunc("/api/new-endpoint", corsMiddleware(h.NewEndpoint))
```

3. **Add request/response types** to `models/book.go` if needed

4. **Add database method** to `database/database.go` if accessing DB

5. **Test** the endpoint:
```bash
curl http://localhost:8080/api/new-endpoint
```

### Adding Database Query

1. **Add SQL query** in `database/database.go`:
```go
func (m *Manager) QueryBooks(condition string) ([]models.Book, error) {
    query := fmt.Sprintf("SELECT * FROM books WHERE %s", condition)
    stmt, err := m.db.Prepare(query)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()
    
    rows, err := stmt.Query()
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var books []models.Book
    for rows.Next() {
        var book models.Book
        err := rows.Scan(&book.ID, &book.Title, /* ... */)
        if err != nil {
            return nil, err
        }
        books = append(books, book)
    }
    return books, nil
}
```

2. **Use in handler**:
```go
books, err := h.db.QueryBooks("author LIKE ?")
```

### Modifying EPUB Processing

1. **Update parser** in `epub/editor.go` or `conversion/epub_parser.go`
2. **Test with sample files** in `../data/ebooks/`
3. **Handle edge cases** (corrupted files, rare EPUB variants)
4. **Update quarantine logic** if adding validation

## Important Patterns & Best Practices

### Error Handling

```go
// Good: provide context
if err != nil {
    log.Printf("Failed to extract metadata from %s: %v", filePath, err)
    http.Error(w, fmt.Sprintf("Extraction failed: %v", err), http.StatusInternalServerError)
    return
}

// Not great: bare error return
if err != nil {
    return err  // Lost context
}

// Good: wrap for context
if err != nil {
    return fmt.Errorf("failed to process EPUB %s: %w", filePath, err)
}
```

### Logging

```go
// Informational
log.Printf("Scanning directory: %s", path)

// Warnings
log.Printf("Warning: Found invalid EPUB at %s: %v", filePath, err)

// Errors (before returning)
log.Printf("Error processing book: %v", err)

// Startup failures
log.Fatal("Critical error: cannot start server:", err)
```

### JSON Encoding

```go
// Always set Content-Type
w.Header().Set("Content-Type", "application/json")

// Encode response
json.NewEncoder(w).Encode(response)

// Or marshal and write
data, err := json.Marshal(book)
if err != nil {
    return
}
w.Write(data)
```

### File Operations

```go
// Always defer cleanup
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()

// Use pathmanager for user inputs
if err := pathmanager.ValidatePath(userPath); err != nil {
    return fmt.Errorf("invalid path: %w", err)
}

// Create directories with proper permissions
os.MkdirAll(path, 0755)
```

## Database Schema

The SQLite database includes:
- `books` table - Main library
- `quarantine_books` table - Books with import issues
- Indices for common queries (author, title, etc.)

**Adding a Column**:
1. Add field to `Book` struct in `models/book.go`
2. Create migration SQL
3. Add to schema initialization in `database.go`
4. Update `SaveBook()` and query methods

**Note**: No migration framework is used; schema is initialized from SQL in code.

## Testing

```bash
# Test individual package
cd backend && go test ./database

# Test with verbose output
go test -v ./...

# Run specific test
go test -run TestScanDirectory

# Coverage
go test -cover ./...
```

**Test Files**:
- `pathmanager/pathmanager_test.go` - Path validation tests

**When to Add Tests**:
- New utility functions
- Critical business logic
- Bug fixes (add regression test)

## Configuration Files

### `config.yaml.template`

Template for production/deployment configuration. Users copy this to `config.yaml`.

Key sections:
- `server` - HTTP server host/port
- `database` - SQLite path
- `library` - Directory paths for books, imports, quarantine
- `logging` - Log directory and settings

### `backend/config.dev.yaml`

Development-specific configuration. Used when running `make dev`.

Differences from template:
- Local file paths
- May use different port for testing

## Debugging Tips

### Enable Verbose Logging

```go
// In code, before key operations
log.Printf("DEBUG: Processing EPUB: %s", epubPath)
log.Printf("DEBUG: Extracted title: %s, author: %s", title, author)
```

### Database Inspection

```bash
# Connect to database
sqlite3 data/ebooks.db

# Useful queries
SELECT COUNT(*) FROM books;
SELECT * FROM books LIMIT 5;
SELECT * FROM quarantine_books;
.schema books  # View table structure
```

### Common Issues

**Issue**: API returns 500 error
- **Check**: Backend logs for error message
- **Debug**: Add logging to handler to trace execution

**Issue**: Database locked error
- **Cause**: Another process has database open
- **Fix**: Close other connections, restart server

**Issue**: EPUB parsing fails
- **Debug**: Check if file is valid EPUB (it's a ZIP)
- **Solution**: Quarantine invalid files and log error

**Issue**: Metadata extraction returns empty
- **Check**: Does EPUB have metadata? Some have minimal data
- **Consider**: Fallback to filename parsing

## Performance Considerations

1. **Database Queries**: 
   - `GetAllBooks()` loads entire library into memory
   - For large libraries (>10k books), consider pagination
   - Use indices on frequently queried columns (author, title)

2. **EPUB Processing**:
   - Parsing EPUB requires unzipping and XML parsing
   - Cache extracted metadata
   - Process imports asynchronously

3. **File I/O**:
   - Directory scanning can be slow with many files
   - Run scans in goroutines to avoid blocking API
   - Implement progress reporting for long operations

4. **Image Serving**:
   - Cache cover images after extraction
   - Compress images before serving
   - Consider CDN for production

## Security Considerations

1. **Path Validation**: Always validate file paths to prevent directory traversal
2. **SQL Injection**: Use parameterized queries (already done)
3. **CORS**: Currently permissive (`*`). Restrict in production.
4. **File Uploads**: Validate EPUB structure before processing
5. **Temporary Files**: Clean up temporary conversion files
6. **Logging**: Don't log sensitive information (passwords, tokens)

## Deployment Considerations

### Docker Build

```dockerfile
# See ../Dockerfile
# Go binary compiled in container
# Configuration via volume mounts
```

### Production Checklist

- [ ] Set proper file permissions on data directories
- [ ] Use reverse proxy (Caddy) for TLS
- [ ] Configure CORS for production domain
- [ ] Set up log rotation
- [ ] Backup SQLite database regularly
- [ ] Monitor disk space (ebooks + database)
- [ ] Set resource limits on container

## References

- **Frontend Integration**: See `../AGENTS.md` for frontend guidance
- **Configuration**: `config.yaml.template`
- **Development Setup**: `../DEVELOPMENT.md`
- **Go Best Practices**: https://golang.org/doc/effective_go
- **SQLite Documentation**: https://www.sqlite.org/
