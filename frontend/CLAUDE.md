# FableFlow Frontend - Agents Guide

This document provides detailed guidance for AI agents working on the JavaScript frontend of FableFlow.

## Frontend Architecture Overview

The FableFlow frontend is a vanilla JavaScript single-page application (SPA) that provides a responsive web interface for managing an ebook library. It communicates with the Go backend via REST API calls.

### Technology Stack

- **Language**: Vanilla JavaScript (ES6+)
- **Templating**: HTML with Alpine.js-like reactive patterns
- **Styling**: CSS (custom, no framework)
- **Build**: No build step needed (served as static files)
- **Dependencies**: 
  - `epub.js` - EPUB reader library
  - `jszip` - ZIP file handling for EPUB
- **Server**: Python development server (for frontend assets)

### Key Points

- **No Framework**: Not using React, Vue, Angular - plain JavaScript
- **Alpine.js-like Pattern**: Reactive state management with automatic DOM updates
- **Single Page App**: Navigation without full page reloads
- **Responsive Design**: Works on desktop, tablet, and mobile

## Directory Structure

```
frontend/
├── static/
│   ├── js/
│   │   └── app.js              # Main application (entire SPA logic)
│   ├── css/
│   │   └── style.css           # All styling (single CSS file)
│   ├── default-book.svg        # Placeholder book cover image
│   ├── epub.min.js             # EPUB reader library
│   └── jszip.min.js            # ZIP library (for EPUB processing)
├── templates/
│   ├── index.html              # Main interface
│   └── reader.html             # EPUB reader page
└── [AGENTS.md]                 # This file
```

## Core File: `static/js/app.js`

This single JavaScript file contains the entire frontend application logic. It's a large file (~2000+ lines) with the following structure:

### Application State Structure

```javascript
function fableFlowApp() {
    return {
        // View/Navigation State
        currentView: 'home',  // Current page: 'home', 'search', 'authors', etc.
        
        // Search & Discovery
        searchQuery: '',
        searchResults: [],
        recentBooks: [],
        randomBooks: [],
        
        // Metadata & Editing
        metadataSearch: {
            loading: false,
            suggestions: [],
            confidence: 0,
            query: ''
        },
        
        // Navigation (Browsing)
        authorLetters: [],
        authorsByLetter: [],
        currentAuthor: '',
        titleLetters: [],
        titlesByLetter: [],
        
        // Book Management
        editingBook: {
            id: null,
            title: '',
            author: '',
            isbn: ''
        },
        selectedBook: null,
        quarantineBooks: [],
        
        // Import & Conversion
        importStatus: null,
        importLogs: [],
        
        // Library
        libraryStats: null,
        
        // UI
        darkMode: false,
        toast: { show: false, message: '' },
        loading: false,
        
        // Methods...
    }
}
```

### Key Methods/Functions

**Lifecycle:**
- `init()` - Initialize app on page load
- `initializeDarkMode()` - Load dark mode preference

**Navigation:**
- `setView(view)` - Switch between views (home, search, authors, etc.)
- `navigateTo(view, param)` - Navigate with parameters

**Data Loading:**
- `loadRecentBooks()` - Fetch recent additions
- `loadRandomBooks()` - Fetch random selection
- `loadQuarantineBooks()` - Load books with import issues
- `getAuthors()` - Fetch author list
- `getTitles()` - Fetch title list
- `getBooksByAuthor(author)` - Filter by author
- `getBooksByTitle(title)` - Filter by title
- `getLibraryStats()` - Get library statistics

**Search:**
- `searchBooks()` - Full-text search implementation
- `searchMetadata()` - Search external metadata sources
- `lookupISBN(isbn)` - Look up ISBN information

**Book Management:**
- `getBookById(id)` - Get specific book details
- `editBook(book)` - Edit book metadata
- `deleteBook(id)` - Delete from library
- `showBookInfo(book)` - Display book information page

**Quarantine (Import Issues):**
- `handleQuarantineEdit()` - Edit quarantined book
- `handleQuarantineDelete()` - Remove quarantined book

**Format Conversion:**
- `startConversion(bookId, format)` - Initiate conversion
- `checkConversionStatus()` - Poll conversion progress

**Import:**
- `startImport()` - Trigger import process
- `checkImportStatus()` - Poll import progress
- `loadImportLogs()` - Fetch import operation logs

**UI Helpers:**
- `toggleDarkMode()` - Toggle dark/light theme
- `showToast(message)` - Display temporary notification
- `closeModal()` - Close UI modals

## Main Templates

### `templates/index.html`

The main interface with:

**Key Sections**:
- Header with search bar
- Navigation menu
- Main content area (changes with `currentView`)
- Modals for dialogs
- Toast notifications

**Views** (controlled by `currentView`):
1. **home** - Recent books grid display
2. **search** - Search results
3. **authors** - Browse by author
4. **titles** - Browse by title
5. **book-info** - Detailed book information
6. **quarantine** - Import issues
7. **import** - Import status and logs
8. **conversion** - Format conversion

**Key Elements**:
```html
<div x-data="fableFlowApp()" x-init="init()">
    <!-- Navigation -->
    <header>
        <input type="text" x-model="searchQuery" @keyup="searchBooks()">
    </header>
    
    <!-- Content -->
    <div x-show="currentView === 'home'">
        <!-- Home view content -->
    </div>
    
    <!-- Repeat for each view -->
</div>
```

### `templates/reader.html`

Dedicated EPUB reader page:

**Features**:
- Embedded `epub.js` reader
- Book navigation controls
- Metadata display
- Download/conversion options

**Key Logic**:
- Reads book ID from URL parameter
- Initializes EPUB reader with book file
- Handles page navigation
- Displays book title and progress

## API Communication Pattern

### Standard Fetch Pattern

```javascript
// Simple GET
async loadBooks() {
    this.loading = true;
    try {
        const response = await fetch('/api/books');
        if (!response.ok) {
            throw new Error(`API Error: ${response.status}`);
        }
        const data = await response.json();
        this.books = data.books || [];
    } catch (error) {
        console.error('Error loading books:', error);
        this.showToast('Failed to load books');
    } finally {
        this.loading = false;
    }
}

// POST with data
async searchMetadata() {
    const response = await fetch('/api/books/search-metadata', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            title: this.metadataSearch.query,
            author: ''
        })
    });
    const data = await response.json();
    this.metadataSearch.suggestions = data.suggestions || [];
}
```

### API Endpoints Called

**Books**:
- `GET /api/books` - All books
- `GET /api/books/{id}` - Single book
- `GET /api/books/recent` - Recent additions
- `GET /api/books/random` - Random selection
- `GET /api/search?q=...` - Full-text search
- `POST /api/books/search-metadata` - Metadata lookup
- `POST /api/books/lookup-isbn` - ISBN lookup

**Organization**:
- `GET /api/authors` - All authors
- `GET /api/authors/letter` - Authors starting with letter
- `GET /api/authors/books?author=...` - Books by author
- `GET /api/titles` - All titles
- `GET /api/titles/letter` - Titles starting with letter
- `GET /api/titles/books?title=...` - Books by title

**Reading**:
- `GET /read/{id}` - Redirect to reader page
- `GET /api/epub/{id}` - Get EPUB file for reading

**Management**:
- `GET /api/quarantine` - Quarantined books
- `POST /api/quarantine/edit` - Edit quarantine book
- `POST /api/quarantine/delete` - Delete quarantine book

**Import & Conversion**:
- `POST /api/import/start` - Start import
- `GET /api/import/status` - Check import status
- `GET /api/import/logs` - Get import logs
- `POST /api/convert` - Start conversion
- `GET /api/convert/status` - Check conversion progress
- `GET /api/convert/{id}` - Download converted file

**Utilities**:
- `GET /api/covers/{id}` - Book cover image
- `GET /api/library/stats` - Library statistics

## Styling: `static/css/style.css`

Single CSS file with all styling for the entire application.

### CSS Organization

```css
/* Reset and base styles */
* { /* global resets */ }
body { /* base styling */ }

/* Layout */
.container { /* main container */ }
header { /* top navigation */ }
main { /* content area */ }

/* Components */
.book-card { /* book grid items */ }
.modal { /* dialogs */ }
.toast { /* notifications */ }
button { /* button styling */ }
input { /* form inputs */ }

/* Views */
.home-view { /* home page */ }
.search-view { /* search results */ }
.authors-view { /* author browsing */ }

/* Dark Mode */
@media (prefers-color-scheme: dark) { /* dark theme */ }
```

### Key Classes

**Layout**:
- `.container` - Main width constraint
- `.grid` - Book grid layout
- `.flex` - Flexbox utility
- `.full-width` - Full viewport width

**Components**:
- `.book-card` - Individual book in grid
- `.button` - Button styling
- `.input` - Text input styling
- `.modal` - Dialog box
- `.toast` - Notification popup
- `.spinner` - Loading indicator

**Visibility**:
- Use `x-show` for conditional display (no DOM removal)
- Use `x-if` for conditional rendering (DOM manipulation)
- Animate transitions with CSS

**Dark Mode**:
- Toggle class `dark` on document element
- CSS media query or class-based rules for theme

## Common Development Tasks

### Adding a New View/Page

1. **Add view case** in navigation:
```javascript
navigateTo(view, param) {
    this.currentView = view;
    this.breadcrumb = [];
    
    switch(view) {
        case 'my-new-view':
            this.loadMyData();
            break;
    }
}
```

2. **Add template** in `index.html`:
```html
<div x-show="currentView === 'my-new-view'">
    <!-- HTML structure for new view -->
</div>
```

3. **Add styling** to `style.css`:
```css
.my-new-view {
    /* styling */
}
```

4. **Add data** to app state:
```javascript
myViewData: {
    items: [],
    selectedItem: null,
    loading: false
}
```

5. **Add methods** to load/display data:
```javascript
async loadMyData() {
    const response = await fetch('/api/my-endpoint');
    this.myViewData.items = await response.json();
}
```

### Adding a New API Call

1. **Create method** in `app.js`:
```javascript
async myAPICall() {
    this.loading = true;
    try {
        const response = await fetch('/api/endpoint', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(this.myData)
        });
        if (!response.ok) throw new Error('API Error');
        return await response.json();
    } catch (error) {
        console.error('Error:', error);
        this.showToast('Error: ' + error.message);
    } finally {
        this.loading = false;
    }
}
```

2. **Call from event handler**:
```html
<button @click="myAPICall()">Click Me</button>
```

3. **Update state** with response:
```javascript
const result = await this.myAPICall();
if (result) {
    this.myData = result;
}
```

### Modifying Styling

1. **Edit** `static/css/style.css`
2. **Test** in browser (refresh, no build needed)
3. **Check** responsive design (mobile view)
4. **Verify** dark mode looks good

### Adding a New Modal/Dialog

1. **Add HTML** to `index.html`:
```html
<div x-show="showMyModal" class="modal">
    <div class="modal-content">
        <h2>My Dialog</h2>
        <input x-model="myInput">
        <button @click="handleMyAction()">Action</button>
        <button @click="showMyModal = false">Close</button>
    </div>
</div>
```

2. **Add state**:
```javascript
showMyModal: false,
myInput: ''
```

3. **Add action handler**:
```javascript
async handleMyAction() {
    // Do something with this.myInput
    this.showMyModal = false;
}
```

### Dark Mode Implementation

The frontend supports dark mode with:

**Storage**: `localStorage` key `darkMode`

**Toggle**:
```javascript
toggleDarkMode() {
    this.darkMode = !this.darkMode;
    if (this.darkMode) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
    localStorage.setItem('darkMode', this.darkMode);
}
```

**Styling**:
```css
/* Light mode (default) */
body { background: white; color: black; }

/* Dark mode */
body.dark { background: #1a1a1a; color: #fff; }
```

## Important Patterns & Best Practices

### State Management

```javascript
// Good: Clear structure for related data
editingBook: {
    id: null,
    title: '',
    author: '',
    isbn: ''
}

// Bad: Scattered related state
editBookTitle: '',
editBookAuthor: '',
editBookISBN: ''

// Update state
this.editingBook.title = 'New Title';

// Trigger reactivity with entire object reassignment if needed
this.editingBook = { ...this.editingBook, title: 'New Title' };
```

### Conditional Rendering

```html
<!-- Show/hide without DOM manipulation -->
<div x-show="isLoading">Loading...</div>

<!-- Add/remove from DOM -->
<div x-if="showDetails">
    <!-- Only rendered when true -->
</div>

<!-- Loop over items -->
<div x-for="book in books" :key="book.id">
    <h3 x-text="book.title"></h3>
</div>
```

### Error Handling

```javascript
async loadData() {
    try {
        const response = await fetch('/api/endpoint');
        if (!response.ok) {
            throw new Error(`HTTP Error: ${response.status}`);
        }
        this.data = await response.json();
    } catch (error) {
        console.error('Failed to load data:', error);
        this.showToast('Error: Unable to load data');
        // Don't leave user hanging - show error state
        this.data = [];
    }
}
```

### Event Binding

```html
<!-- Click handler -->
<button @click="doSomething()">Click</button>

<!-- Input with model binding -->
<input x-model="searchQuery" @keyup="onSearch()">

<!-- Form submission -->
<form @submit.prevent="handleSubmit()">
    <input x-model="email">
    <button type="submit">Submit</button>
</form>

<!-- Keyboard shortcuts -->
<input @keydown.enter="submit()">
```

### DOM Updates

```javascript
// Array updates - assign whole array for reactivity
this.books = [...this.books, newBook];  // Add
this.books = this.books.filter(b => b.id !== id);  // Remove

// Or object method (works the same)
this.books.push(newBook);
```

## Performance Considerations

1. **Lazy Loading**: Don't load all data at startup
   - Load recent books initially
   - Load author list only when visiting author view
   - Load search results only on search

2. **Debouncing**: For rapid events like typing
   ```javascript
   let searchTimeout;
   onSearchInput() {
       clearTimeout(searchTimeout);
       searchTimeout = setTimeout(() => this.searchBooks(), 300);
   }
   ```

3. **Caching**: Avoid redundant API calls
   ```javascript
   authors: null,
   async getAuthors() {
       if (this.authors) return this.authors;  // Cached
       // Load from API
   }
   ```

4. **Image Optimization**: Book covers can be large
   - Load cover only when needed
   - Cache in localStorage if possible
   - Compress images on backend

## Accessibility (a11y)

1. **Semantic HTML**: Use `<button>`, `<a>`, `<input>` appropriately
2. **ARIA Labels**: Add labels for screen readers
   ```html
   <input aria-label="Search books" type="text">
   ```
3. **Keyboard Navigation**: All controls should be keyboard-accessible
4. **Color Contrast**: Ensure text is readable
5. **Focus States**: Visible focus indicators

## Browser Compatibility

- Modern browsers (Chrome, Firefox, Safari, Edge)
- CSS Grid and Flexbox support required
- ES6 JavaScript features (const, let, arrow functions)
- Fetch API support

**Testing**: Test in:
- Chrome/Chromium
- Firefox
- Safari (macOS and iOS)
- Mobile browsers

## Debugging

### Console Logging

```javascript
// Useful for debugging
console.log('Current state:', this);
console.log('Books loaded:', this.books);
console.log('Search results:', this.searchResults);
```

### Common Issues

**Issue**: View doesn't update after API call
- **Cause**: State not updated, or updated incorrectly
- **Debug**: Log state after API call
- **Fix**: Ensure state assignment happens

**Issue**: Modal doesn't close
- **Cause**: Modal flag not toggled
- **Fix**: Check `@click="showModal = false"` is present

**Issue**: API returns 404
- **Cause**: Wrong API endpoint URL
- **Debug**: Check Network tab in DevTools
- **Fix**: Verify endpoint path matches backend

**Issue**: Styling doesn't apply
- **Cause**: CSS not loaded or selector specificity
- **Debug**: Check DevTools Elements tab
- **Fix**: Verify CSS file link, check selector

## File Permissions & Assets

**Static Files** (served as-is):
- `js/app.js` - Application logic
- `css/style.css` - Styling
- `*.svg` - Images
- `*.min.js` - Libraries

**All files are public** - No sensitive data in frontend

## Security Considerations

1. **XSS Prevention**: Avoid `innerHTML`, use `x-text` or `textContent`
   ```javascript
   // Bad
   element.innerHTML = userInput;
   
   // Good
   element.textContent = userInput;
   // Or in template
   <div x-text="userInput"></div>
   ```

2. **CSRF Protection**: Use POST for state-changing operations
3. **Input Validation**: Validate on frontend for UX, but also on backend
4. **No Secrets**: Never store API keys, passwords, or tokens in frontend code
5. **CORS**: Backend should validate origin in production

## Development Workflow

### Local Development

```bash
make dev
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
```

### Making Changes

1. Edit files in `frontend/`
2. Refresh browser (no build needed)
3. Check DevTools for errors
4. Verify backend is working

### Testing Features

```javascript
// Open console (F12) and test manually
// Example:
fetch('/api/books').then(r => r.json()).then(d => console.log(d))
```

### Code Review Checklist

- [ ] No console errors in DevTools
- [ ] Works in multiple browsers
- [ ] Responsive design (test mobile view)
- [ ] Dark mode works
- [ ] Error states handled
- [ ] Loading states shown
- [ ] No hardcoded values
- [ ] API endpoints correct
- [ ] Accessibility considered

## References

- **Main Application Guide**: `../AGENTS.md`
- **Backend API Guide**: `../backend/AGENTS.md`
- **Development Setup**: `../DEVELOPMENT.md`
- **EPUB.js Documentation**: https://github.com/futurepress/epub.js
- **MDN JavaScript Guide**: https://developer.mozilla.org/en-US/docs/Web/JavaScript
