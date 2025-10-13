package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"fableflow/backend/importservice"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importService *importservice.ImportService
}

// NewImportHandler creates a new ImportHandler
func NewImportHandler(importService *importservice.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// StartImportRequest represents the request to start an import
type StartImportRequest struct {
	DryRun bool `json:"dry_run"`
}

// StartImportResponse represents the response from starting an import
type StartImportResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// ImportStatusResponse represents the import status response
type ImportStatusResponse struct {
	SessionID        string   `json:"session_id"`
	Status           string   `json:"status"`
	TotalFiles       int      `json:"total_files"`
	ProcessedFiles   int      `json:"processed_files"`
	ImportedFiles    int      `json:"imported_files"`
	QuarantinedFiles int      `json:"quarantined_files"`
	SkippedFiles     int      `json:"skipped_files"`
	Progress         int      `json:"progress"` // Percentage
	Errors           []string `json:"errors"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time,omitempty"`
}

// StartImport handles starting a new import session
func (h *ImportHandler) StartImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Start import session
	session, err := h.importService.StartImport(req.DryRun)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	response := StartImportResponse{
		SessionID: session.ID,
		Message:   "Import session started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImportStatus handles getting the current import status
func (h *ImportHandler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := h.importService.GetStatus()
	if session == nil {
		http.Error(w, "No active import session", http.StatusNotFound)
		return
	}

	// Calculate progress percentage
	progress := 0
	if session.TotalFiles > 0 {
		progress = (session.ProcessedFiles * 100) / session.TotalFiles
	}

	response := ImportStatusResponse{
		SessionID:        session.ID,
		Status:           session.Status,
		TotalFiles:       session.TotalFiles,
		ProcessedFiles:   session.ProcessedFiles,
		ImportedFiles:    session.ImportedFiles,
		QuarantinedFiles: session.QuarantinedFiles,
		SkippedFiles:     session.SkippedFiles,
		Progress:         progress,
		Errors:           session.Errors,
		StartTime:        session.StartTime.Format("2006-01-02 15:04:05"),
	}

	if session.EndTime != nil {
		response.EndTime = session.EndTime.Format("2006-01-02 15:04:05")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImportLogs handles getting import session logs
func (h *ImportHandler) GetImportLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id parameter required", http.StatusBadRequest)
		return
	}

	// For now, return a simple response
	// TODO: Implement log retrieval from disk
	response := map[string]string{
		"message":    "Log retrieval not yet implemented",
		"session_id": sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListImportLogs handles listing available import session logs
func (h *ImportHandler) ListImportLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get list of available logs from the import service
	logs, err := h.importService.GetAvailableLogs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetImportLog handles getting a specific import session log
func (h *ImportHandler) GetImportLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	sessionID := pathParts[4]

	// Get the specific log
	log, err := h.importService.GetLog(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}
