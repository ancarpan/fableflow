function fableFlowApp() {
    return {
        // State management
        currentView: 'home',
        loading: false,
        searchQuery: '',
        searchResults: [],
        recentBooks: [],
        darkMode: false,
        showAboutModal: false,
        
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
        },

        navigateToBreadcrumb(index) {
            // Navigate back to a specific breadcrumb level
            // index 0 = Home, index 1 = first breadcrumb item, etc.
            if (index === 0) {
                this.goHome();
                return;
            }
            
            const targetBreadcrumb = this.breadcrumb[index - 1];
            this.breadcrumb = this.breadcrumb.slice(0, index);
            
            // Set the current view without making API calls
            this.currentView = targetBreadcrumb.view;
        },

        addToBreadcrumb(name, view) {
            this.breadcrumb.push({ name, view });
        },

        // Search functionality
        async performSearch() {
            if (!this.searchQuery.trim()) return;
            
            this.loading = true;
            this.currentView = 'search';
            this.breadcrumb = [];
            this.addToBreadcrumb(`Search: "${this.searchQuery}"`, 'search');
            
            try {
                const response = await fetch(`/api/search?q=${encodeURIComponent(this.searchQuery)}`);
                if (!response.ok) throw new Error('Search failed');
                
                this.searchResults = await response.json();
            } catch (error) {
                console.error('Search error:', error);
                this.showToast('Search failed. Please try again.');
            } finally {
                this.loading = false;
            }
        },

        // Author browsing
        async browseAuthors() {
            this.loading = true;
            this.currentView = 'authors';
            this.breadcrumb = [];
            this.addToBreadcrumb('Authors', 'authors');
            
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
            this.currentView = 'authors-letter';
            this.currentLetter = letter;
            this.addToBreadcrumb(`Authors (${letter})`, 'authors-letter');
            
            try {
                const response = await fetch(`/api/authors/letter?letter=${encodeURIComponent(letter)}`);
                if (!response.ok) throw new Error('Failed to load authors by letter');
                
                this.authorsByLetter = await response.json();
            } catch (error) {
                console.error('Authors by letter error:', error);
                this.showToast('Failed to load authors.');
            } finally {
                this.loading = false;
            }
        },

        async browseBooksByAuthor(author) {
            this.loading = true;
            this.currentView = 'books-by-author';
            this.currentAuthor = author;
            this.addToBreadcrumb(author, 'books-by-author');
            
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
            this.currentView = 'titles';
            this.breadcrumb = [];
            this.addToBreadcrumb('Titles', 'titles');
            
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
            this.currentView = 'random';
            this.breadcrumb = [];
            this.addToBreadcrumb('Random', 'random');
            
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
            this.currentView = 'titles-letter';
            this.currentLetter = letter;
            this.addToBreadcrumb(`Titles (${letter})`, 'titles-letter');
            
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
            this.currentView = 'books-by-title';
            this.currentTitle = title;
            this.addToBreadcrumb(title, 'books-by-title');
            
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
            this.toast.message = message;
            this.toast.show = true;
            setTimeout(() => {
                this.toast.show = false;
            }, 3000);
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
            this.breadcrumb = [
                { name: 'Admin', action: null }
            ];
            // Load import logs when showing admin panel
            this.loadImportLogs();
        },

        // Show library info page
        showLibraryInfo() {
            this.currentView = 'library-info';
            this.breadcrumb = [
                { name: 'Library Info', action: null }
            ];
            // Load library statistics
            this.loadLibraryStats();
        },

        // Show about popup
        showAbout() {
            this.showAboutModal = true;
        },

        // Edit book metadata
        async editBook(bookId) {
            try {
                this.loading = true;
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
                    
                    this.currentView = 'edit';
                    this.breadcrumb = [
                        { name: 'Home', action: () => this.goHome() },
                        { name: 'Edit Book', action: null }
                    ];
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
                    response = await fetch(`/api/books/${this.editingBook.id}/edit`, {
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
                    const error = await response.text();
                    throw new Error(error);
                }
                
                this.isbnLookup.fetchedData = await response.json();
                this.showToast('Metadata found successfully!');
                
            } catch (error) {
                console.error('ISBN lookup failed:', error);
                this.showToast('ISBN lookup failed: ' + error.message);
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
            this.breadcrumb = [
                { name: 'Quarantine', action: null }
            ];
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
                publisher: book.publisher || ''
            };
            
            // Reset ISBN lookup data
            this.isbnLookup = {
                isbn: '',
                loading: false,
                fetchedData: null
            };
            
            this.currentView = 'edit';
            this.breadcrumb = [
                { name: 'Quarantine', action: () => this.showQuarantine() },
                { name: 'Edit Book', action: null }
            ];
        }
    }
}