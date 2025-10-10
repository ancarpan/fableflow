package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"fableflow/backend/config"
	"fableflow/backend/database"
	"fableflow/backend/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create database manager
	db, err := database.NewManager(cfg.Database.Path)
	if err != nil {
		log.Fatal("Failed to create database manager:", err)
	}
	defer db.Close()

	// Auto-scan if enabled
	if cfg.Library.AutoScan {
		log.Printf("Auto-scanning enabled, scanning: %s", cfg.Library.ScanDirectory)
		go func() {
			err := db.ScanDirectory(cfg.Library.ScanDirectory)
			if err != nil {
				log.Printf("Auto-scan error: %v", err)
			} else {
				log.Printf("Auto-scan completed")
			}
		}()
	}

	// Create handlers
	booksHandler := handlers.NewBooksHandler(db)
	scanHandler := handlers.NewScanHandler(db)
	healthHandler := handlers.NewHealthHandler()

	// Setup routes
	http.HandleFunc("/api/health", healthHandler.HealthCheck)
	http.HandleFunc("/api/books", booksHandler.GetAllBooks)
	http.HandleFunc("/api/books/", booksHandler.GetBookByID)
	http.HandleFunc("/api/search", booksHandler.SearchBooks)
	http.HandleFunc("/api/authors", booksHandler.GetAuthors)
	http.HandleFunc("/api/authors/letter", booksHandler.GetAuthorsByLetter)
	http.HandleFunc("/api/authors/books", booksHandler.GetBooksByAuthor)
	http.HandleFunc("/api/titles", booksHandler.GetTitles)
	http.HandleFunc("/api/titles/letter", booksHandler.GetTitlesByLetter)
	http.HandleFunc("/api/titles/books", booksHandler.GetBooksByTitle)
	http.HandleFunc("/api/scan", scanHandler.ScanDirectory)
	http.HandleFunc("/api/rescan", scanHandler.RescanDirectory)
	http.HandleFunc("/api/download/", booksHandler.DownloadBook)

	// Add CORS headers for frontend development
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For now, return a simple message
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "FableFlow API", "version": "1.0.0"}`)
	})

	// Start server
	address := cfg.Server.Host + ":" + cfg.Server.Port
	fmt.Printf("🚀 FableFlow API starting on http://%s\n", address)
	fmt.Printf("📚 Default scan directory: %s\n", cfg.Library.ScanDirectory)
	fmt.Printf("🔧 Configuration: %s\n", func() string {
		if _, err := os.Stat("config.yaml"); err == nil {
			return "config.yaml (loaded)"
		}
		return "defaults (config.yaml not found)"
	}())
	fmt.Println("📖 API is ready to serve requests!")

	log.Fatal(http.ListenAndServe(address, nil))
}
