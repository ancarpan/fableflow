function fableFlowApp() {
    return {
        // State management
        currentView: 'home',
        loading: false,
        loadingMessage: 'Caricamento...',
        searchQuery: '',
        searchResults: [],
        recentBooks: [],
        darkMode: false,
        showAboutModal: false,
        
        // Metadata search
        metadataSearch: {
            loading: false,
            suggestions: [],
            confidence: 0,
            searched: false,
            query: '' // Editable search query
        },
        
        // App version
        appVersion: 'v0.1-alpha',
        
        // Edit book data
        editingBook: {
            id: null,
            title: '',
            author: '',
            isbn: '',
            publisher: ''
        },
        
        // ISBN lookup data
        isbnLookup: {
            isbn: '',
            loading: false,
            fetchedData: null
        },
        
        // Navigation data
        breadcrumb: [],
        authorLetters: [],
        titleLetters: [],
        authorsByLetter: [],
        titlesByLetter: [],
        booksByAuthor: [],
        booksByTitle: [],
        randomBooks: [],
        currentLetter: '',
        currentAuthor: '',
        currentTitle: '',
        
        // Author search functionality
        authorSearchQuery: '',
        filteredAuthorsByLetter: [],
        
        
        // Toast notifications
        toast: {
            show: false,
            message: ''
        },
        
        // Import functionality
        importStatus: null,
        importLogs: [],
        selectedLog: null,
        
        // Library statistics
        libraryStats: null,
        
        // Quarantine data
        quarantineBooks: [],
        showDeleteConfirm: false,
        deleteConfirmBook: null,
        
        // Book deletion
        showDeleteBookConfirm: false,
        
        // Book info page
        selectedBook: null,

        // Initialize the application
        init() {
            this.initializeDarkMode();
            this.loadRecentBooks();
        },

        // Initialize dark mode from localStorage
        initializeDarkMode() {
            const savedDarkMode = localStorage.getItem('darkMode');
            console.log('Initializing dark mode, saved value:', savedDarkMode);
            if (savedDarkMode === 'true') {
                this.darkMode = true;
                document.documentElement.classList.add('dark');
                console.log('Dark mode initialized as enabled');
            } else {
                this.darkMode = false;
                document.documentElement.classList.remove('dark');
                console.log('Dark mode initialized as disabled');
            }
        },

        // Toggle dark mode
        toggleDarkMode() {
            console.log('Toggle dark mode called, current state:', this.darkMode);
            this.darkMode = !this.darkMode;
            if (this.darkMode) {
                document.documentElement.classList.add('dark');
                localStorage.setItem('darkMode', 'true');
                console.log('Dark mode enabled - class added:', document.documentElement.classList.contains('dark'));
            } else {
                document.documentElement.classList.remove('dark');
                localStorage.setItem('darkMode', 'false');
                console.log('Dark mode disabled - class removed:', !document.documentElement.classList.contains('dark'));
            }
        },

        // Generate letter arrays for navigation (fallback only)
        generateLetterArrays() {
            // Generate A-Z arrays as fallback when no data is available
            this.authorLetters = Array.from({length: 26}, (_, i) => String.fromCharCode(65 + i));
            this.titleLetters = Array.from({length: 26}, (_, i) => String.fromCharCode(65 + i));
        },

        // Navigation methods
        goHome() {
            this.currentView = 'home';
            this.breadcrumb = [];
            this.searchQuery = '';
            this.loadRecentBooks();
        },

        goBack() {
            if (this.breadcrumb.length > 1) {
                // Remove last breadcrumb item and navigate to the previous level
                this.breadcrumb = this.breadcrumb.slice(0, -1);
                this.navigateToPage(this.breadcrumb[this.breadcrumb.length - 1]);
            } else {
                // If only one breadcrumb item left, go home
                this.goHome();
            }
        },

        navigateToBreadcrumb(index) {
            // Navigate back to a specific breadcrumb level
            // index 0 = Home, index 1 = first breadcrumb item, etc.
            if (index === 0) {
                this.goHome();
                return;
            }
            
            this.breadcrumb = this.breadcrumb.slice(0, index);
            this.navigateToPage(this.breadcrumb[this.breadcrumb.length - 1]);
        },

        // Centralized navigation method
        navigateToPage(pageName) {
            switch (pageName) {
                case 'Home':
                    this.goHome();
                    break;
                case 'Random':
                    this.browseRandom();
                    break;
                case 'Admin':
                    this.showAdminPanel();
                    break;
                case 'Quarantine':
                    this.showQuarantine();
                    break;
                case 'Edit Book':
                    // Stay on current edit page
                    break;
                case 'Authors':
                    this.browseAuthors();
                    break;
                case 'Authors (A)':
                    this.browseAuthorsByLetter(this.currentLetter);
                    break;
                case 'Author Name':
                    this.browseBooksByAuthor(this.currentAuthor);
                    break;
                case 'Titles':
                    this.browseTitles();
                    break;
                case 'Titles (A)':
                    this.browseTitlesByLetter(this.currentLetter);
                    break;
                case 'Title Name':
                    this.browseBooksByTitle(this.currentTitle);
                    break;
                default:
                    // Handle dynamic author/title names and letter patterns
                    if (pageName.startsWith('Authors (')) {
                        const letter = pageName.match(/Authors \((.)\)/)[1];
                        this.browseAuthorsByLetter(letter);
                    } else if (pageName.startsWith('Titles (')) {
                        const letter = pageName.match(/Titles \((.)\)/)[1];
                        this.browseTitlesByLetter(letter);
                    } else {
                        // Handle dynamic author/title names (not in switch cases)
                        this.goHome();
                    }
            }
        },


        // Search functionality
        async performSearch() {
            if (!this.searchQuery.trim()) {
                // If no search query, show all books
                console.log('No search query, showing all books');
                this.currentView = 'search';
                this.breadcrumb = ['Home', 'All Books'];
                this.loading = true;
                this.loadingMessage = 'Caricamento di tutti i libri...';
                
                try {
                    const response = await fetch('/api/search');
                    if (!response.ok) throw new Error('Failed to load books');
                    
                    const results = await response.json();
                    console.log('All books results:', results);
                    this.searchResults = Array.isArray(results) ? results : [];
                } catch (error) {
                    console.error('Error loading all books:', error);
                    this.showToast('Failed to load books');
                    this.searchResults = [];
                } finally {
                    this.loading = false;
                }
                return;
            }
            
            console.log('Performing search for:', this.searchQuery);
            this.loading = true;
            this.loadingMessage = `Ricerca di "${this.searchQuery}"...`;
            this.currentView = 'search';
            this.breadcrumb = ['Home', `Search: "${this.searchQuery}"`];
            
            try {
                const url = `/api/search?q=${encodeURIComponent(this.searchQuery)}`;
                console.log('Search URL:', url);
                const response = await fetch(url);
                console.log('Search response status:', response.status);
                
                if (!response.ok) throw new Error('Search failed');
                
                const results = await response.json();
                console.log('Search results:', results);
                // Ensure we have a valid array
                this.searchResults = Array.isArray(results) ? results : [];
                console.log('Processed search results:', this.searchResults);
            } catch (error) {
                console.error('Search error:', error);
                this.showToast('Search failed. Please try again.');
                this.searchResults = [];
            } finally {
                this.loading = false;
            }
        },

        // Filter authors by search query
        filterAuthorsByLetter() {
            if (!this.authorSearchQuery.trim()) {
                this.filteredAuthorsByLetter = this.authorsByLetter;
                return;
            }
            
            const query = this.authorSearchQuery.toLowerCase();
            this.filteredAuthorsByLetter = this.authorsByLetter.filter(author => 
                author.toLowerCase().includes(query)
            );
        },

        // Author browsing
        async browseAuthors() {
            this.loading = true;
            this.loadingMessage = 'Caricamento autori...';
            this.currentView = 'authors';
            this.breadcrumb = ['Home', 'Authors'];
            
            try {
                const response = await fetch('/api/authors');
                if (!response.ok) throw new Error('Failed to load authors');
                
                const authors = await response.json();
                
                // Handle null or empty responses gracefully
                if (!authors || !Array.isArray(authors) || authors.length === 0) {
                    this.authorLetters = [];
                    return;
                }
                
                // Generate letters that have authors
                const lettersWithAuthors = new Set();
                authors.forEach(author => {
                    if (author && author.length > 0) {
                        lettersWithAuthors.add(author.charAt(0).toUpperCase());
                    }
                });
                this.authorLetters = Array.from(lettersWithAuthors).sort();
            } catch (error) {
                console.error('Authors error:', error);
                this.showToast('Failed to load authors.');
                this.authorLetters = [];
            } finally {
                this.loading = false;
            }
        },

        async browseAuthorsByLetter(letter) {
            this.loading = true;
            this.loadingMessage = `Caricamento autori con lettera ${letter}...`;
            this.currentView = 'authors-letter';
            this.currentLetter = letter;
            this.authorSearchQuery = ''; // Reset search when changing letters
            this.breadcrumb = ['Home', 'Authors', `Authors (${letter})`];
            
            try {
                const response = await fetch(`/api/authors/letter?letter=${encodeURIComponent(letter)}`);
                if (!response.ok) throw new Error('Failed to load authors by letter');
                
                this.authorsByLetter = await response.json();
                this.filteredAuthorsByLetter = this.authorsByLetter; // Initialize filtered list
            } catch (error) {
                console.error('Authors by letter error:', error);
                this.showToast('Failed to load authors.');
            } finally {
                this.loading = false;
            }
        },

        async browseBooksByAuthor(author) {
            this.loading = true;
            this.loadingMessage = `Caricamento libri di ${author}...`;
            this.currentView = 'books-by-author';
            this.currentAuthor = author;
            
            // Set currentLetter if not already set (e.g., when clicking from book card)
            if (!this.currentLetter && author && author.length > 0) {
                this.currentLetter = author.charAt(0).toUpperCase();
            }
            
            // Build breadcrumb - use currentLetter if available, otherwise just show author
            if (this.currentLetter) {
                this.breadcrumb = ['Home', 'Authors', `Authors (${this.currentLetter})`, author];
            } else {
                this.breadcrumb = ['Home', 'Authors', author];
            }
            
            try {
                const response = await fetch(`/api/authors/books?author=${encodeURIComponent(author)}`);
                if (!response.ok) throw new Error('Failed to load books by author');
                
                this.booksByAuthor = await response.json();
            } catch (error) {
                console.error('Books by author error:', error);
                this.showToast('Failed to load books.');
            } finally {
                this.loading = false;
            }
        },

        // Title browsing
        async browseTitles() {
            this.loading = true;
            this.loadingMessage = 'Caricamento titoli...';
            this.currentView = 'titles';
            this.breadcrumb = ['Home', 'Titles'];
            
            try {
                const response = await fetch('/api/titles');
                if (!response.ok) throw new Error('Failed to load titles');
                
                const titles = await response.json();
                
                // Handle null or empty responses gracefully
                if (!titles || !Array.isArray(titles) || titles.length === 0) {
                    this.titleLetters = [];
                    return;
                }
                
                // Generate letters that have titles
                const lettersWithTitles = new Set();
                titles.forEach(title => {
                    if (title && title.length > 0) {
                        lettersWithTitles.add(title.charAt(0).toUpperCase());
                    }
                });
                this.titleLetters = Array.from(lettersWithTitles).sort();
            } catch (error) {
                console.error('Titles error:', error);
                this.showToast('Failed to load titles.');
                this.titleLetters = [];
            } finally {
                this.loading = false;
            }
        },

        // Random books browsing
        async browseRandom() {
            this.loading = true;
            this.loadingMessage = 'Caricamento libri random...';
            this.currentView = 'random';
            this.breadcrumb = ['Home', 'Random'];
            
            try {
                const response = await fetch('/api/books/random?limit=24');
                if (!response.ok) throw new Error('Failed to load random books');
                
                const books = await response.json();
                
                // Handle null or empty responses gracefully
                if (!books || !Array.isArray(books)) {
                    this.randomBooks = [];
                } else {
                    this.randomBooks = books;
                }
            } catch (error) {
                console.error('Random books error:', error);
                this.showToast('Failed to load random books.');
                this.randomBooks = [];
            } finally {
                this.loading = false;
            }
        },

        async browseTitlesByLetter(letter) {
            this.loading = true;
            this.loadingMessage = `Caricamento titoli con lettera ${letter}...`;
            this.currentView = 'titles-letter';
            this.currentLetter = letter;
            this.breadcrumb = ['Home', 'Titles', `Titles (${letter})`];
            
            try {
                const response = await fetch(`/api/titles/letter?letter=${encodeURIComponent(letter)}`);
                if (!response.ok) throw new Error('Failed to load titles by letter');
                
                this.titlesByLetter = await response.json();
            } catch (error) {
                console.error('Titles by letter error:', error);
                this.showToast('Failed to load titles.');
            } finally {
                this.loading = false;
            }
        },

        async browseBooksByTitle(title) {
            this.loading = true;
            this.loadingMessage = `Caricamento libri con titolo "${title}"...`;
            this.currentView = 'books-by-title';
            this.currentTitle = title;
            this.breadcrumb = ['Home', 'Titles', `Titles (${this.currentLetter})`, title];
            
            try {
                const response = await fetch(`/api/titles/books?title=${encodeURIComponent(title)}`);
                if (!response.ok) throw new Error('Failed to load books by title');
                
                this.booksByTitle = await response.json();
            } catch (error) {
                console.error('Books by title error:', error);
                this.showToast('Failed to load books.');
            } finally {
                this.loading = false;
            }
        },

        // Library management
        async scanLibrary() {
            this.loading = true;
            this.loadingMessage = 'Scansione libreria in corso...';
            this.showToast('Scanning library...');
            
            try {
                const response = await fetch('/api/scan', { method: 'POST' });
                if (!response.ok) throw new Error('Scan failed');
                
                const result = await response.json();
                this.showToast(`Scan completed! Added: ${result.added}, Removed: ${result.removed}`);
            } catch (error) {
                console.error('Scan error:', error);
                this.showToast('Scan failed. Please try again.');
            } finally {
                this.loading = false;
            }
        },

        // Conversion functionality
        async convertBook(bookId, format) {
            this.loading = true;
            this.loadingMessage = `Conversione in ${format.toUpperCase()}...`;
            this.showToast(`Converting to ${format.toUpperCase()}...`);
            
            try {
                const response = await fetch('/api/convert', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        book_id: bookId,
                        output_format: format
                    })
                });
                
                if (!response.ok) {
                    const errorText = await response.text();
                    let error;
                    try {
                        error = JSON.parse(errorText);
                    } catch (e) {
                        error = { message: errorText };
                    }
                    throw new Error(error.message || 'Conversion failed');
                }
                
                const result = await response.json();
                this.showToast(`Conversion completed! File will be available for download for 1 hour.`);
                
                // Automatically download the converted file
                window.open(`/api/convert/${bookId}/${format}`, '_blank');
                
            } catch (error) {
                console.error('Conversion error:', error);
                this.showToast(`Conversion failed: ${error.message}`);
            } finally {
                this.loading = false;
            }
        },

        async checkConversionStatus() {
            try {
                const response = await fetch('/api/convert/status');
                if (!response.ok) throw new Error('Failed to check conversion status');
                
                const status = await response.json();
                return status;
            } catch (error) {
                console.error('Status check error:', error);
                return { available: false, supported_formats: [], output_formats: [] };
            }
        },


        // Utility functions
        formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },

        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleDateString();
        },

        showToast(message) {
            console.log('showToast called with message:', message);
            this.toast.message = message;
            this.toast.show = true;
            setTimeout(() => {
                this.toast.show = false;
            }, 5000); // Increased timeout to 5 seconds for better visibility
        },

        // Load recent books for homepage
        async loadRecentBooks() {
            try {
                const response = await fetch('/api/books/recent?limit=24');
                if (response.ok) {
                    const books = await response.json();
                    // Check if books is not null/undefined and is an array
                    if (books && Array.isArray(books)) {
                        this.recentBooks = books;
                    } else {
                        // No books available
                        this.recentBooks = [];
                    }
                }
            } catch (error) {
                console.error('Error loading recent books:', error);
            }
        },

        // Show admin panel
        showAdminPanel() {
            this.currentView = 'admin';
            this.breadcrumb = ['Home', 'Admin'];
            // Load import logs when showing admin panel
            this.loadImportLogs();
        },

        // Show library info page
        showLibraryInfo() {
            this.currentView = 'library-info';
            this.breadcrumb = ['Home', 'Library Info'];
            console.log('Library Info breadcrumb set to:', this.breadcrumb);
            console.log('Breadcrumb type:', typeof this.breadcrumb, 'Is array:', Array.isArray(this.breadcrumb));
            // Load library statistics
            this.loadLibraryStats();
        },

        // Show about popup
        showAbout() {
            this.showAboutModal = true;
        },

        // Show book info page
        async showBookInfo(bookId) {
            try {
                this.loading = true;
                this.loadingMessage = 'Caricamento informazioni libro...';
                const response = await fetch(`/api/books/${bookId}`);
                if (response.ok) {
                    const book = await response.json();
                    this.selectedBook = book;
                    this.currentView = 'book-info';
                    this.breadcrumb = ['Home', book.title || 'Book Info'];
                } else {
                    this.showToast('Error loading book data');
                }
            } catch (error) {
                console.error('Error loading book:', error);
                this.showToast('Error loading book data');
            } finally {
                this.loading = false;
            }
        },

        // Edit book metadata
        async editBook(bookId) {
            try {
                this.loading = true;
                this.loadingMessage = 'Caricamento dati libro...';
                const response = await fetch(`/api/books/${bookId}`);
                if (response.ok) {
                    const book = await response.json();
                    this.editingBook = {
                        id: book.id,
                        title: book.title || '',
                        author: book.author || '',
                        isbn: book.isbn || '',
                        publisher: book.publisher || ''
                    };
                    
                    // Reset ISBN lookup data when editing a new book
                    this.isbnLookup = {
                        isbn: '',
                        loading: false,
                        fetchedData: null
                    };
                    
                    // Auto-populate search query for metadata search
                    if (book.title || book.author) {
                        this.metadataSearch.query = [book.title || '', book.author || ''].filter(Boolean).join(', ');
                    } else {
                        this.metadataSearch.query = '';
                    }
                    
                    // Reset metadata search results
                    this.metadataSearch.suggestions = [];
                    this.metadataSearch.searched = false;
                    this.metadataSearch.confidence = 0;
                    
                    this.currentView = 'edit';
                    this.breadcrumb = ['Home', 'Edit Book'];
                } else {
                    this.showToast('Error loading book data');
                }
            } catch (error) {
                console.error('Error loading book:', error);
                this.showToast('Error loading book data');
            } finally {
                this.loading = false;
            }
        },

        // Save book metadata
        async saveBookMetadata() {
            try {
                this.loading = true;
                this.loadingMessage = 'Salvataggio modifiche...';
                
                // Check if this is a quarantine book (ID is 0 or null)
                const isQuarantineBook = !this.editingBook.id || this.editingBook.id === 0;
                
                let response;
                if (isQuarantineBook) {
                    // Use quarantine edit endpoint
                    response = await fetch('/api/quarantine/edit', {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            file_path: this.editingBook.file_path,
                            title: this.editingBook.title,
                            author: this.editingBook.author,
                            isbn: this.editingBook.isbn,
                            publisher: this.editingBook.publisher
                        })
                    });
                } else {
                    // Use regular book edit endpoint
                    response = await fetch(`/api/books/${this.editingBook.id}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            title: this.editingBook.title,
                            author: this.editingBook.author,
                            isbn: this.editingBook.isbn,
                            publisher: this.editingBook.publisher
                        })
                    });
                }

                if (response.ok) {
                    if (isQuarantineBook) {
                        this.showToast('Quarantine book processed successfully!');
                        // Go back to quarantine page
                        this.showQuarantine();
                    } else {
                        this.showToast('Book metadata updated successfully!');
                    }
                } else {
                    this.showToast('Error saving book metadata');
                }
            } catch (error) {
                console.error('Error saving book:', error);
                this.showToast('Error saving book metadata');
            } finally {
                this.loading = false;
            }
        },

        // Cancel edit
        cancelEdit() {
            this.editingBook = {
                id: null,
                title: '',
                author: '',
                isbn: '',
                publisher: ''
            };
            
            // Reset ISBN lookup data when canceling edit
            this.isbnLookup = {
                isbn: '',
                loading: false,
                fetchedData: null
            };
            
            this.goBack();
        },
        
        // Import functionality
        async startImport(dryRun) {
            try {
                const response = await fetch('/api/import/start', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ dry_run: dryRun })
                });
                
                if (!response.ok) {
                    const error = await response.text();
                    throw new Error(error);
                }
                
                const result = await response.json();
                console.log('Import started:', result);
                
                // Start auto-refresh for real-time updates
                this.startAutoRefresh();
                
            } catch (error) {
                console.error('Failed to start import:', error);
                this.showToast('Failed to start import: ' + error.message);
            }
        },
        
        async refreshImportStatus() {
            try {
                const response = await fetch('/api/import/status');
                
                if (response.status === 404) {
                    this.importStatus = null;
                    localStorage.removeItem('importRunning');
                    return;
                }
                
                if (!response.ok) {
                    throw new Error('Failed to get import status');
                }
                
                this.importStatus = await response.json();
                
                // Clear the refresh flag if import is no longer running
                if (this.importStatus && this.importStatus.status !== 'running') {
                    localStorage.removeItem('importRunning');
                }
                
            } catch (error) {
                console.error('Failed to get import status:', error);
                this.importStatus = null;
                localStorage.removeItem('importRunning');
            }
        },
        
        pollImportStatus() {
            const poll = () => {
                this.refreshImportStatus();
                
                // Continue polling if import is still running
                if (this.importStatus && this.importStatus.status === 'running') {
                    setTimeout(poll, 1000); // Poll every 1 second
                }
            };
            
            poll();
        },
        
        // Auto-refresh import status when import is running
        startAutoRefresh() {
            // Set a flag to indicate import is running
            localStorage.setItem('importRunning', 'true');
            
            // Immediate first refresh
            this.refreshImportStatus();
            
            // Then refresh the import status every second
            const refreshInterval = setInterval(() => {
                // Check if import is still running by checking localStorage
                const isRunning = localStorage.getItem('importRunning') === 'true';
                if (isRunning) {
                    // Just refresh the import status, not the whole page
                    this.refreshImportStatus();
                } else {
                    clearInterval(refreshInterval);
                }
            }, 1000);
        },
        
        // Load import logs
        async loadImportLogs() {
            try {
                const response = await fetch('/api/import/logs/list');
                if (!response.ok) {
                    throw new Error('Failed to load import logs');
                }
                
                this.importLogs = await response.json();
            } catch (error) {
                console.error('Failed to load import logs:', error);
                this.showToast('Failed to load import logs');
            }
        },

        // Load library statistics
        async loadLibraryStats() {
            try {
                const response = await fetch('/api/library/stats');
                if (!response.ok) {
                    throw new Error('Failed to load library statistics');
                }
                
                this.libraryStats = await response.json();
            } catch (error) {
                console.error('Failed to load library statistics:', error);
                this.showToast('Failed to load library statistics');
            }
        },
        
        // View specific import log
        async viewImportLog(sessionId) {
            try {
                const response = await fetch(`/api/import/logs/${sessionId}`);
                if (!response.ok) {
                    throw new Error('Failed to load import log');
                }
                
                this.selectedLog = await response.json();
                console.log('Import log details:', this.selectedLog);
            } catch (error) {
                console.error('Failed to load import log:', error);
                this.showToast('Failed to load import log');
            }
        },
        
        // Format date and time for display
        formatDateTime(dateString) {
            if (!dateString) return 'N/A';
            const date = new Date(dateString);
            return date.toLocaleString();
        },
        
        // Format duration between two dates
        formatDuration(startTime, endTime) {
            if (!startTime || !endTime) return 'N/A';
            const start = new Date(startTime);
            const end = new Date(endTime);
            const diffMs = end - start;
            
            if (diffMs < 1000) {
                return `${diffMs}ms`;
            } else if (diffMs < 60000) {
                return `${Math.round(diffMs / 1000)}s`;
            } else {
                const minutes = Math.floor(diffMs / 60000);
                const seconds = Math.floor((diffMs % 60000) / 1000);
                return `${minutes}m ${seconds}s`;
            }
        },
        
        // ISBN lookup functionality
        async lookupISBN() {
            if (!this.isbnLookup.isbn.trim()) {
                this.showToast('Please enter an ISBN');
                return;
            }
            
            this.isbnLookup.loading = true;
            this.isbnLookup.fetchedData = null;
            
            try {
                const response = await fetch('/api/books/lookup-isbn', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ isbn: this.isbnLookup.isbn })
                });
                
                if (!response.ok) {
                    // Read response body once (as text, then parse if JSON)
                    const responseText = await response.text();
                    let errorMessage = null;
                    
                    // Try to parse as JSON (backend returns JSON errors)
                    try {
                        const errorData = JSON.parse(responseText);
                        if (errorData.error) {
                            errorMessage = errorData.error;
                        }
                    } catch (e) {
                        // Not JSON, will use status-based fallback
                        console.log('Failed to parse error response as JSON:', responseText);
                    }
                    
                    // Determine the message to show
                    let messageToShow = null;
                    if (errorMessage) {
                        messageToShow = errorMessage;
                    } else if (response.status === 404) {
                        messageToShow = 'No book found for this ISBN. The book may not be available in Google Books.';
                    } else if (response.status === 503) {
                        messageToShow = 'Google Books API is temporarily unavailable. Please try again later.';
                    } else if (response.status === 502) {
                        messageToShow = 'Unable to connect to Google Books API. Please try again later.';
                    } else {
                        messageToShow = 'ISBN lookup failed. Please try again.';
                    }
                    
                    // Always show toast message
                    console.log('ISBN lookup error - Status:', response.status, 'Message:', messageToShow);
                    this.showToast(messageToShow);
                    this.isbnLookup.fetchedData = null;
                    this.isbnLookup.loading = false;
                    return;
                }
                
                this.isbnLookup.fetchedData = await response.json();
                this.showToast('Metadata found successfully!');
                
            } catch (error) {
                console.error('ISBN lookup failed:', error);
                this.showToast('ISBN lookup failed: ' + (error.message || 'Network error occurred'));
                this.isbnLookup.fetchedData = null;
            } finally {
                this.isbnLookup.loading = false;
            }
        },
        
        // Import field from fetched data to current form
        importField(fieldName) {
            if (!this.isbnLookup.fetchedData) {
                this.showToast('No fetched data available');
                return;
            }
            
            const value = this.isbnLookup.fetchedData[fieldName];
            if (!value || value === 'Not found') {
                this.showToast(`No ${fieldName} found in fetched data`);
                return;
            }
            
            // Update the current form field
            this.editingBook[fieldName] = value;
            this.showToast(`${fieldName} imported successfully!`);
        },
        
        // Quarantine functionality
        async showQuarantine() {
            this.currentView = 'quarantine';
            this.breadcrumb = ['Home', 'Admin', 'Quarantine'];
            await this.loadQuarantineBooks();
        },
        
        async loadQuarantineBooks() {
            try {
                const response = await fetch('/api/quarantine');
                if (!response.ok) {
                    throw new Error('Failed to load quarantine books');
                }
                this.quarantineBooks = await response.json();
            } catch (error) {
                console.error('Error loading quarantine books:', error);
                this.showToast('Failed to load quarantine books');
                this.quarantineBooks = [];
            }
        },
        
        editQuarantineBook(book) {
            // Set up the editing book with quarantine book data
            this.editingBook = {
                id: book.id || 0, // Quarantine books have no database ID
                file_path: book.file_path, // Store file path for quarantine books
                title: book.title,
                author: book.author,
                isbn: book.isbn || '',
                publisher: book.publisher || '',
                // Quarantine information
                quarantine_reason: book.quarantine_reason || '',
                quarantine_detail: book.quarantine_detail || '',
                quarantine_date: book.quarantine_date || ''
            };
            
            // Reset ISBN lookup data
            this.isbnLookup = {
                isbn: '',
                loading: false,
                fetchedData: null
            };
            
            // Clear metadata search state when switching books
            this.metadataSearch = {
                loading: false,
                suggestions: [],
                confidence: 0,
                searched: false
            };
            
            this.currentView = 'edit';
            this.breadcrumb = ['Home', 'Admin', 'Quarantine', 'Edit Book'];
        },
        
        deleteQuarantineBook(book) {
            this.deleteConfirmBook = book;
            this.showDeleteConfirm = true;
        },
        
        async confirmDeleteQuarantineBook() {
            if (!this.deleteConfirmBook) return;
            
            const book = this.deleteConfirmBook;
            this.showDeleteConfirm = false;
            
            try {
                const response = await fetch('/api/quarantine/delete', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        file_path: book.file_path
                    }),
                });

                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.error || 'Failed to delete quarantine book');
                }

                this.showToast('Quarantine book deleted successfully');
                // Reload quarantine books
                await this.loadQuarantineBooks();
            } catch (error) {
                console.error('Error deleting quarantine book:', error);
                this.showToast('Failed to delete quarantine book: ' + error.message);
            } finally {
                this.deleteConfirmBook = null;
            }
        },
        
        cancelDeleteQuarantineBook() {
            this.showDeleteConfirm = false;
            this.deleteConfirmBook = null;
        },
        
        // Book deletion functions
        deleteBook() {
            if (!this.editingBook || !this.editingBook.id) {
                this.showToast('No book selected for deletion');
                return;
            }
            this.showDeleteBookConfirm = true;
        },
        
        async confirmDeleteBook() {
            if (!this.editingBook || !this.editingBook.id) {
                return;
            }
            
            const bookId = this.editingBook.id;
            this.showDeleteBookConfirm = false;
            
            try {
                const response = await fetch(`/api/books/${bookId}`, {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });

                if (!response.ok) {
                    const responseText = await response.text();
                    let errorMessage = 'Failed to delete book';
                    
                    // Try to parse as JSON
                    try {
                        const errorData = JSON.parse(responseText);
                        if (errorData.error) {
                            errorMessage = errorData.error;
                        }
                    } catch (e) {
                        // Not JSON, use text directly
                        if (responseText) {
                            errorMessage = responseText;
                        }
                    }
                    
                    throw new Error(errorMessage);
                }

                const result = await response.json();
                this.showToast(result.message || 'Book deleted successfully');
                
                // Redirect to home
                this.goHome();
            } catch (error) {
                console.error('Error deleting book:', error);
                this.showToast('Failed to delete book: ' + error.message);
            } finally {
                this.showDeleteBookConfirm = false;
            }
        },
        
        cancelDeleteBook() {
            this.showDeleteBookConfirm = false;
        },
        
        // Metadata search functions
        async searchMetadata(source = 'openlibrary') {
            if (!this.metadataSearch.query || !this.metadataSearch.query.trim()) {
                this.showToast('Please enter a search query');
                return;
            }
            
            this.metadataSearch.loading = true;
            this.metadataSearch.searched = true;
            
            try {
                // Parse query - try to split into title and author
                // Simple heuristic: if query contains comma, split by comma
                // Otherwise, use entire query as title
                const queryParts = this.metadataSearch.query.split(',').map(s => s.trim());
                let title = '';
                let author = '';
                
                if (queryParts.length >= 2) {
                    // Assume format: "title, author" or "author, title"
                    // Try to detect which is which based on common patterns
                    title = queryParts[0];
                    author = queryParts.slice(1).join(', ');
                } else {
                    // Use entire query as title
                    title = this.metadataSearch.query.trim();
                }
                
                // Build API URL with source parameter
                const apiUrl = `/api/books/search-metadata?source=${encodeURIComponent(source)}`;
                
                const response = await fetch(apiUrl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        title: title,
                        author: author
                    })
                });
                
                if (!response.ok) {
                    throw new Error('Search failed');
                }
                
                const data = await response.json();
                // Ensure suggestions is always an array (handle null/undefined)
                const suggestions = Array.isArray(data.suggestions) ? data.suggestions : [];
                this.metadataSearch.suggestions = suggestions;
                this.metadataSearch.confidence = data.confidence || 0;
                
                // Auto-suggest if high confidence (>80%) and we have suggestions
                if (data.confidence > 0.8 && suggestions.length > 0) {
                    this.useSuggestion(suggestions[0]);
                    this.showToast('High confidence match found and applied automatically!');
                } else if (suggestions.length > 0) {
                    const sourceName = source === 'googlebooks' ? 'Google Books' : 'Open Library';
                    this.showToast(`Found ${suggestions.length} suggestions from ${sourceName}. Please review and choose the best match.`);
                } else {
                    const sourceName = source === 'googlebooks' ? 'Google Books' : 'Open Library';
                    this.showToast(`No matching books found in ${sourceName}.`);
                }
                
            } catch (error) {
                console.error('Metadata search error:', error);
                this.showToast('Failed to search for metadata. Please try again.');
                this.metadataSearch.suggestions = [];
                this.metadataSearch.confidence = 0;
            } finally {
                this.metadataSearch.loading = false;
            }
        },
        
        useSuggestion(suggestion) {
            // Update the form with the suggested metadata
            this.editingBook.title = suggestion.title;
            this.editingBook.author = suggestion.author;
            this.editingBook.isbn = suggestion.isbn || '';
            this.editingBook.publisher = suggestion.publisher || '';
            
            // Clear the search results
            this.metadataSearch.suggestions = [];
            this.metadataSearch.searched = false;
            
            this.showToast('Metadata updated from Open Library');
        },
        
        // Get cover image URL for the current editing book
        getCoverImage() {
            // For quarantine books, generate cover filename from file path
            if (this.editingBook.file_path) {
                const baseName = this.editingBook.file_path.split('/').pop();
                const coverName = baseName.replace(/\.epub$/i, '_cover.jpg');
                return '/api/quarantine/covers/' + coverName + '?t=' + Date.now();
            }
            
            // For regular books with database ID
            if (this.editingBook.id) {
                return '/api/covers/' + this.editingBook.id + '?t=' + Date.now();
            }
            
            // Default cover
            return '/static/default-book.svg';
        }
    }
}