package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"fableflow/backend/database"
	"fableflow/backend/models"
)

// ScanHandler handles scan-related HTTP requests
type ScanHandler struct {
	db *database.Manager
}

// NewScanHandler creates a new scan handler
func NewScanHandler(db *database.Manager) *ScanHandler {
	return &ScanHandler{db: db}
}

// ScanDirectory starts a scan of the specified directory
func (h *ScanHandler) ScanDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	// Start scan in background
	go func() {
		log.Printf("Starting scan of: %s", req.Path)
		err := h.db.ScanDirectory(req.Path)
		if err != nil {
			log.Printf("Error scanning directory: %v", err)
		} else {
			log.Printf("Scan completed for: %s", req.Path)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ScanResponse{Status: "scan started"})
}

// RescanDirectory performs a rescan that adds new books and removes unavailable ones
func (h *ScanHandler) RescanDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	log.Printf("Starting rescan of: %s", req.Path)
	added, removed, err := h.db.RescanDirectory(req.Path)
	if err != nil {
		log.Printf("Error rescanning directory: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Rescan completed for: %s - Added: %d, Removed: %d", req.Path, added, removed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ScanResponse{
		Status:  "rescan completed",
		Added:   added,
		Removed: removed,
	})
}
