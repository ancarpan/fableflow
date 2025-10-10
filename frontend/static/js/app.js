function fableFlowApp() {
    return {
        // State management
        currentView: 'home',
        loading: false,
        searchQuery: '',
        searchResults: [],
        
        // Navigation data
        breadcrumb: [],
        authorLetters: [],
        titleLetters: [],
        authorsByLetter: [],
        titlesByLetter: [],
        booksByAuthor: [],
        booksByTitle: [],
        currentLetter: '',
        currentAuthor: '',
        currentTitle: '',
        
        // Toast notifications
        toast: {
            show: false,
            message: ''
        },

        // Initialize the application
        init() {
            this.generateLetterArrays();
        },

        // Generate letter arrays for navigation
        generateLetterArrays() {
            // Generate A-Z arrays
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
        }
    }
}