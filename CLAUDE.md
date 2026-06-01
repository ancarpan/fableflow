# FableFlow - Agents Guide

This document provides guidance for AI agents working on the FableFlow codebase. FableFlow is a modern ebook management system with a Go backend API and a JavaScript frontend.

## Project Overview

**FableFlow** is a full-stack ebook library management application with:
- 📚 EPUB ebook library management with metadata extraction
- 🔍 Search and browsing by author, title, ISBN
- 📖 Built-in EPUB reader
- 🔄 Auto-import and directory scanning
- 📱 Responsive web interface
- 🐳 Docker containerized deployment

## Architecture

### Technology Stack

| Layer | Technology | Port |
|-------|-----------|------|
| **Frontend** | Vanilla JavaScript (Alpine.js-like) | 3000 |
| **Backend** | Go 1.21+ | 8080 |
| **Database** | SQLite | - |
| **Storage** | Local filesystem | - |
| **Deployment** | Docker (Caddy reverse proxy) | - |

### Directory Structure

```
fableflow/
├── backend/              # Go API server and core logic
│   ├── config/          # Configuration loading
│   ├── database/        # SQLite database operations
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data structures and types
│   ├── metadata/        # Book metadata extraction
│   ├── epub/            # EPUB editing capabilities
│   ├── conversion/      # Format conversion (EPUB to Kindle, etc)
│   ├── importservice/   # Book import pipeline
│   ├── pathmanager/     # File path utilities
│   ├── kindlegen/       # Kindlegen binaries for conversion
│   └── main.go          # API server entry point
├── frontend/            # Web interface
│   ├── static/
│   │   ├── js/         # JavaScript application logic
│   │   ├── css/        # Styling
│   │   └── ...         # Assets (SVGs, libs)
│   └── templates/       # HTML files
├── dev/                 # Development utilities
├── data/                # Runtime data (ebooks, logs, db)
├── Makefile            # Build commands
└── docker-compose.yml  # Development setup
```

## Key Principles

1. **Clear Separation of Concerns**: Backend handles API, data, and business logic. Frontend handles UI and user interaction.
2. **API-First Design**: All frontend features communicate via the `/api/*` endpoints.
3. **Library Isolation**: Books exist in three states: main library, quarantine (import problems), and imports.
4. **File-Based Operations**: Heavy use of filesystem operations for EPUB processing, imports, and storage.

## Core Features & Architecture

### 1. Book Management
- **Database**: SQLite with models for `Book`, `QuarantineBook`
- **Main API Endpoints**:
  - `GET /api/books` - Fetch all books
  - `GET /api/books/{id}` - Get single book details
  - `GET /api/books/recent` - Recently added books
  - `GET /api/books/random` - Random book selection
  - `GET /api/search` - Full-text search

### 2. Metadata & Discovery
- **External Metadata**: Integration with book metadata services
- **API Endpoints**:
  - `POST /api/books/search-metadata` - Find metadata for book
  - `POST /api/books/lookup-isbn` - ISBN lookup
- **Metadata Extraction**: Automatic extraction from EPUB files on import

### 3. Library Organization
- **Authors**: Browse books by author with alphabetical indexing
- **Titles**: Browse books by title with alphabetical indexing
- **Quarantine System**: Books with import issues are quarantined for manual review
- **Import Service**: Automatic or manual import with logging and error handling

### 4. Format Conversion
- **EPUB to Kindle**: Uses kindlegen binary (platform-specific)
- **Status Tracking**: Async conversion with status polling
- **API Endpoints**:
  - `POST /api/convert` - Start conversion
  - `GET /api/convert/status` - Check conversion progress
  - `GET /api/convert/{id}` - Download converted file

### 5. EPUB Processing
- **Reading**: Built-in EPUB reader using epub.js library
- **Editing**: Ability to modify EPUB metadata and content
- **Parsing**: Extract information, cover images, TOC

## Development Workflow

### Getting Started

```bash
# Development mode
make dev

# Access services
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# Backend Health: http://localhost:8080/api/health
```

### Making Changes

**Backend Changes** (Go):
1. Edit file in `backend/`
2. Stop the server (Ctrl+C)
3. Run `make dev` again to restart with changes

**Frontend Changes** (JavaScript):
1. Edit files in `frontend/`
2. Changes are served immediately (no restart needed)

### Testing & Verification

```bash
# Check backend compiles
cd backend && go build

# Run tests
cd backend && go test ./...
```

## Working with the API

### Request/Response Patterns

All API responses follow a consistent pattern:

```javascript
// Success (with data)
{ "data": {...}, "success": true }

// Success (list)
{ "books": [...], "total": 100 }

// Error
{ "error": "description", "message": "detailed message" }
```

### CORS & Cross-Origin Requests

- All API endpoints have CORS enabled (`Access-Control-Allow-Origin: *`)
- Frontend makes requests from port 3000 to backend on port 8080
- In production (Docker), Caddy handles proxying

### Authentication

**Current Status**: No authentication implemented. All endpoints are open.

**Considerations for Future Implementation**:
- Add auth middleware in backend
- Store tokens in frontend's localStorage
- Include `Authorization` header in API requests (already included in CORS headers)

## Common Tasks for Agents

### Adding a New API Endpoint

1. **Define the handler** in `backend/handlers/` (or create new file if different domain)
2. **Register the route** in `backend/main.go` with CORS middleware if needed
3. **Add request/response types** to `backend/models/book.go` if needed
4. **Connect to database** via the `db` parameter in handler initialization
5. **Call from frontend** in `frontend/static/js/app.js` using `fetch()` to `/api/...`

### Adding a UI Feature

1. **Plan the UI** in the appropriate template file (`frontend/templates/index.html` or `reader.html`)
2. **Add HTML structure** with proper Alpine.js bindings if using reactive data
3. **Add styling** to `frontend/static/css/style.css`
4. **Implement logic** in `frontend/static/js/app.js` or separate module
5. **Connect to backend** via API calls

### Processing Files/EPUB Content

1. **File Operations**: Use `backend/pathmanager/` for safe path handling
2. **EPUB Parsing**: Use `backend/epub/` package for reading EPUB structure
3. **Metadata Extraction**: Use `backend/metadata/` for extracting book info
4. **Error Handling**: Use quarantine system for problematic files

### Database Operations

- All database interactions go through `backend/database/Manager`
- Use prepared statements to prevent SQL injection
- Always defer `Close()` on database connections
- Handle errors explicitly (don't ignore database failures)

## Important Patterns & Conventions

### Error Handling (Backend)

```go
// Good: wrap errors with context
if err != nil {
    log.Printf("Failed to scan directory %s: %v", path, err)
    http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
    return
}

// Bad: silent failures
if err != nil {
    _ = err  // Don't do this
}
```

### Frontend State Management

The frontend uses a single app object with reactive data:

```javascript
fableFlowApp() {
    return {
        // State
        currentView: 'home',
        searchQuery: '',
        books: [],
        
        // Methods
        init() { ... },
        async loadBooks() { ... }
    }
}
```

When modifying state:
- Use direct property assignment (`this.property = value`)
- Frontend re-renders automatically
- For lists, reassign the entire array to trigger reactivity

### Configuration

- **Config File**: `backend/config.yaml` (or `config.dev.yaml` for dev)
- **Environment Variables**: Not currently used (but could be added)
- **Runtime Paths**: Managed by `config` and `pathmanager` packages
- **No Hardcoded Paths**: All paths come from config

### Logging

**Backend**:
- Use `log.Printf()` for informational messages
- Use `log.Fatalf()` for critical startup failures
- Include context in log messages (what, where, why)

**Frontend**:
- Use `console.log()` for debugging
- Remove verbose logging before final commits
- Avoid logging sensitive information

## Important Considerations

### Performance

1. **Database Queries**: The `GetAllBooks` endpoint loads all books into memory. May need pagination for large libraries.
2. **EPUB Reading**: Uses epub.js library which is loaded for every reader page
3. **Directory Scanning**: Can be slow with many files; runs asynchronously

### Security

1. **File Path Validation**: Always use `pathmanager` to validate file paths
2. **SQL Injection**: Use parameterized queries (already done in codebase)
3. **CORS**: Currently permissive (`*`). Consider restricting in production.
4. **No Auth**: Add authentication before exposing to untrusted networks

### Compatibility

1. **Browser Support**: No specific constraints mentioned; test in modern browsers
2. **EPUB Format**: Supports EPUB 2 and 3, but some rare EPUB structures may cause issues
3. **Platform-Specific**: kindlegen binaries included for darwin, linux, windows

## Frontend-Backend Communication

### Common API Calls Pattern

```javascript
// Frontend
const response = await fetch('/api/books');
const data = await response.json();

// Handle errors
if (!response.ok) {
    console.error('API Error:', data.error);
    return;
}

// Update state
this.books = data.books || [];
```

### Async Workflows

Some operations are async (imports, conversions):
1. Frontend POSTs to start operation
2. Backend returns a task ID or status
3. Frontend polls `/api/{operation}/status` periodically
4. When complete, frontend fetches results

## Debugging

### Backend Debugging

```bash
# Rebuild and check for compile errors
cd backend && go build

# Run with verbose logging
GOLOG_DEBUG=true go run . -c config.dev.yaml

# Check database
sqlite3 data/ebooks.db ".tables"
```

### Frontend Debugging

1. Open browser DevTools (F12)
2. Check Console tab for JavaScript errors
3. Check Network tab for API failures
4. Use `console.log()` to trace execution

### Common Issues

1. **Port Conflicts**: Kill existing processes on 3000 and 8080
2. **Database Locked**: Close other connections to `ebooks.db`
3. **Missing Config**: Copy `config.yaml.template` to `config.yaml`
4. **Import Failures**: Check quarantine section for error details

## Code Quality Guidelines

1. **Comments**: Add comments for non-obvious logic
2. **Error Messages**: Be specific and helpful
3. **Naming**: Use clear, descriptive names for variables and functions
4. **Modularity**: Keep functions focused and testable
5. **Testing**: Add tests for new functionality
6. **No Console Spam**: Remove excessive logging after development

## File Permissions & Ownership

- Ebook files: Read-only for the app
- Database file: Read-write
- Import directory: Read-only monitoring
- Quarantine directory: Read-write for storing problematic files
- Log files: Write-only for logging

## References

- **Backend**: `backend/AGENTS.md` - Detailed Go backend guidance
- **Frontend**: `frontend/AGENTS.md` - Detailed JavaScript frontend guidance
- **Development**: `DEVELOPMENT.md` - Development setup instructions
- **Configuration**: `config.yaml.template` - Configuration options
