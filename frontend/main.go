package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Get content directory from environment variable
	contentDir := os.Getenv("CONTENT_DIR")
	if contentDir == "" {
		contentDir = "." // Default to current directory
	}

	// Get backend address from environment variable
	backendAddr := os.Getenv("BACKEND_ADDR")
	if backendAddr == "" {
		backendAddr = "http://localhost:8080" // Default backend address
	}

	log.Printf("📁 [CONFIG] Content directory: %s", contentDir)
	log.Printf("🔗 [CONFIG] Backend address: %s", backendAddr)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(contentDir, "static/")))))

	// Serve the main HTML template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📄 [STATIC] %s %s", r.Method, r.URL.Path)
		http.ServeFile(w, r, filepath.Join(contentDir, "templates", "index.html"))
	})

	// Proxy API requests to backend
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔗 [API] %s %s", r.Method, r.URL.Path)
		proxyToBackend(w, r, backendAddr)
	})

	// Proxy reader requests to backend
	http.HandleFunc("/read/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📖 [READER] %s %s", r.Method, r.URL.Path)
		proxyToBackend(w, r, backendAddr)
	})

	log.Println("Frontend server starting on http://localhost:3000")
	log.Println("Make sure the backend API is running on port 8080")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

// Proxy function to avoid code duplication
func proxyToBackend(w http.ResponseWriter, r *http.Request, backendAddr string) {
	start := time.Now()

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		log.Printf("✅ [CORS] %s %s (OPTIONS)", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Proxy to backend API
	backendURL := backendAddr + r.URL.Path
	if r.URL.RawQuery != "" {
		backendURL += "?" + r.URL.RawQuery
	}

	log.Printf("🔄 [PROXY] %s %s -> %s", r.Method, r.URL.Path, backendURL)

	// Create request to backend
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		log.Printf("❌ [ERROR] Failed to create request: %v", err)
		http.Error(w, "Error creating request to backend", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make request to backend
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [ERROR] Backend request failed: %v", err)
		http.Error(w, "Backend API not available", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	duration := time.Since(start)
	log.Printf("✅ [RESPONSE] %s %s -> %d (%v)", r.Method, r.URL.Path, resp.StatusCode, duration)
}
