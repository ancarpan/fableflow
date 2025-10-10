package main

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Serve the main HTML template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("templates", "index.html"))
	})

	// Proxy API requests to backend
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proxy to backend API
		backendURL := "http://localhost:8080" + r.URL.Path
		if r.URL.RawQuery != "" {
			backendURL += "?" + r.URL.RawQuery
		}

		// Create request to backend
		req, err := http.NewRequest(r.Method, backendURL, r.Body)
		if err != nil {
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
	})

	log.Println("Frontend server starting on http://localhost:3000")
	log.Println("Make sure the backend API is running on port 8080")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
