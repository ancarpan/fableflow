package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"fableflow/backend/config"
	"fableflow/backend/database"
	"fableflow/backend/handlers"
	"fableflow/backend/importservice"
)

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "c", "config.yaml", "Configuration file path")
	flag.Parse()

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Configuration file '%s' not found\n", configFile)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration from '%s': %v", configFile, err)
	}

	// Create database manager
	db, err := database.NewManager(cfg.Database.Path)
	if err != nil {
		log.Fatal("Failed to create database manager:", err)
	}
	defer db.Close()

	// Ensure tmp directory exists and is clean
	if err := os.MkdirAll(cfg.TmpDir, 0755); err != nil {
		log.Fatal("Failed to create tmp directory:", err)
	}

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
	booksHandler := handlers.NewBooksHandler(db, &handlers.Config{
		Library: struct {
			ScanDirectory       string `yaml:"scan_directory"`
			QuarantineDirectory string `yaml:"quarantine_directory"`
		}{
			ScanDirectory:       cfg.Library.ScanDirectory,
			QuarantineDirectory: cfg.Library.QuarantineDirectory,
		},
	})
	scanHandler := handlers.NewScanHandler(db)
	healthHandler := handlers.NewHealthHandler()
	conversionHandler := handlers.NewConversionHandler(db, cfg.TmpDir)
	coversHandler := handlers.NewCoversHandler(db)

	// Create import service with scan callback
	importConfig := &importservice.Config{
		ImportDirectory:     cfg.Library.ImportDirectory,
		ScanDirectory:       cfg.Library.ScanDirectory,
		QuarantineDirectory: cfg.Library.QuarantineDirectory,
		LogDir:              cfg.LogDir,
		MaxLogs:             cfg.MaxImportLogs,
	}
	importService := importservice.NewImportService(importConfig, func() {
		// Trigger database scan after import completes
		log.Println("Import completed, triggering database scan...")
		if err := db.ScanDirectory(cfg.Library.ScanDirectory); err != nil {
			log.Printf("Error scanning directory after import: %v", err)
		} else {
			log.Println("Database scan completed successfully")
		}
	})
	importHandler := handlers.NewImportHandler(importService)

	// Setup routes
	http.HandleFunc("/api/health", healthHandler.HealthCheck)
	http.HandleFunc("/api/books", booksHandler.GetAllBooks)
	http.HandleFunc("/api/books/", booksHandler.GetBookByID)
	http.HandleFunc("/api/books/recent", corsMiddleware(booksHandler.GetRecentBooks))
	http.HandleFunc("/api/books/random", corsMiddleware(booksHandler.GetRandomBooks))
	http.HandleFunc("/api/books/lookup-isbn", corsMiddleware(booksHandler.LookupISBN))
	http.HandleFunc("/api/quarantine", corsMiddleware(booksHandler.GetQuarantineBooks))
	http.HandleFunc("/api/quarantine/edit", corsMiddleware(booksHandler.EditQuarantineBook))
	http.HandleFunc("/api/search", booksHandler.SearchBooks)
	http.HandleFunc("/api/authors", booksHandler.GetAuthors)
	http.HandleFunc("/api/authors/letter", booksHandler.GetAuthorsByLetter)
	http.HandleFunc("/api/authors/books", booksHandler.GetBooksByAuthor)
	http.HandleFunc("/api/titles", booksHandler.GetTitles)
	http.HandleFunc("/api/titles/letter", booksHandler.GetTitlesByLetter)
	http.HandleFunc("/api/titles/books", booksHandler.GetBooksByTitle)
	http.HandleFunc("/api/scan", scanHandler.ScanDirectory)
	http.HandleFunc("/read/", booksHandler.ServeReader)
	http.HandleFunc("/api/rescan", scanHandler.RescanDirectory)
	http.HandleFunc("/api/download/", booksHandler.DownloadBook)
	http.HandleFunc("/api/epub/", corsMiddleware(booksHandler.ServeEPUBFile))
	http.HandleFunc("/api/convert/status", corsMiddleware(conversionHandler.GetConversionStatus))
	http.HandleFunc("/api/convert/", corsMiddleware(conversionHandler.DownloadConvertedBook))
	http.HandleFunc("/api/convert", corsMiddleware(conversionHandler.ConvertBook))
	http.HandleFunc("/api/covers/", corsMiddleware(coversHandler.ServeCover))
	http.HandleFunc("/api/import/start", corsMiddleware(importHandler.StartImport))
	http.HandleFunc("/api/import/status", corsMiddleware(importHandler.GetImportStatus))
	http.HandleFunc("/api/import/logs/list", corsMiddleware(importHandler.ListImportLogs))
	http.HandleFunc("/api/import/logs/", corsMiddleware(importHandler.GetImportLog))
	http.HandleFunc("/api/import/logs", corsMiddleware(importHandler.GetImportLogs))

	// Conditionally serve static assets
	if cfg.Server.ServeStaticAssets {
		// Serve static files (CSS, JS, images)
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

		// Serve main HTML template for SPA
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Add CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Serve the main HTML template for SPA
			http.ServeFile(w, r, "./templates/index.html")
		})
	} else {
		// API-only mode - return JSON response for root
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Add CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Return API information
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message": "FableFlow API", "version": "1.0.0", "mode": "api-only"}`)
		})
	}

	// Start server
	address := cfg.Server.Host + ":" + cfg.Server.Port
	if cfg.Server.ServeStaticAssets {
		fmt.Printf("🚀 FableFlow starting on http://%s (serving static assets)\n", address)
	} else {
		fmt.Printf("🚀 FableFlow API starting on http://%s (API-only mode)\n", address)
	}
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
