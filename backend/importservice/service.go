package importservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fableflow/backend/metadata"
)

// QuarantinedBook represents a book that was quarantined during import
type QuarantinedBook struct {
	FilePath       string    `json:"file_path"`
	QuarantinePath string    `json:"quarantine_path"`
	Reason         string    `json:"reason"`
	ErrorDetail    string    `json:"error_detail"`
	Timestamp      time.Time `json:"timestamp"`
}

// ImportSession represents a single import session
type ImportSession struct {
	ID               string            `json:"id"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          *time.Time        `json:"end_time,omitempty"`
	Status           string            `json:"status"` // "running", "completed", "failed"
	DryRun           bool              `json:"dry_run"`
	TotalFiles       int               `json:"total_files"`
	ProcessedFiles   int               `json:"processed_files"`
	ImportedFiles    int               `json:"imported_files"`
	QuarantinedFiles int               `json:"quarantined_files"`
	SkippedFiles     int               `json:"skipped_files"`
	Errors           []string          `json:"errors"`
	QuarantinedBooks []QuarantinedBook `json:"quarantined_books,omitempty"`
	LogPath          string            `json:"log_path"`
}

// ImportService manages book import operations
type ImportService struct {
	config            *Config
	metadataExtractor *metadata.Extractor
	currentSession    *ImportSession
	sessionMutex      sync.RWMutex
	logDir            string
	maxLogs           int
	onComplete        func() // Callback function called when import completes
}

// Config represents the configuration for the import service
type Config struct {
	ImportDirectory     string
	ScanDirectory       string
	QuarantineDirectory string
	LogDir              string
	MaxLogs             int
}

// NewImportService creates a new import service
func NewImportService(config *Config, onComplete func()) *ImportService {
	return &ImportService{
		config:            config,
		metadataExtractor: metadata.NewExtractor(),
		logDir:            config.LogDir,
		maxLogs:           config.MaxLogs,
		onComplete:        onComplete,
	}
}

// StartImport starts a new import session
func (s *ImportService) StartImport(dryRun bool) (*ImportSession, error) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	// Check if there's already an active session
	if s.currentSession != nil && s.currentSession.Status == "running" {
		return nil, fmt.Errorf("import session already in progress")
	}

	// Create new session
	sessionID := fmt.Sprintf("import_%d", time.Now().Unix())
	session := &ImportSession{
		ID:        sessionID,
		StartTime: time.Now(),
		Status:    "running",
		DryRun:    dryRun,
		Errors:    []string{},
		LogPath:   filepath.Join(s.logDir, fmt.Sprintf("%s.json", sessionID)),
	}

	s.currentSession = session

	// Start import process in goroutine
	go s.runImport(session)

	return session, nil
}

// GetStatus returns the current import session status
func (s *ImportService) GetStatus() *ImportSession {
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()

	if s.currentSession == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	session := *s.currentSession
	return &session
}

// runImport performs the actual import process
func (s *ImportService) runImport(session *ImportSession) {
	defer func() {
		s.sessionMutex.Lock()
		if s.currentSession != nil {
			endTime := time.Now()
			s.currentSession.EndTime = &endTime
			if s.currentSession.Status == "running" {
				s.currentSession.Status = "completed"
			}
		}
		s.sessionMutex.Unlock()

		// Save session log
		s.saveSessionLog(session)

		// Call completion callback if not a dry run
		if !session.DryRun && s.onComplete != nil {
			s.onComplete()
		}
	}()

	// Ensure log directory exists
	if err := os.MkdirAll(s.logDir, 0755); err != nil {
		s.logError(session, fmt.Sprintf("Failed to create log directory: %v", err))
		return
	}

	// Scan import directory for EPUB files
	epubFiles, err := s.scanForEPUBFiles(s.config.ImportDirectory)
	if err != nil {
		s.logError(session, fmt.Sprintf("Failed to scan import directory: %v", err))
		return
	}

	s.sessionMutex.Lock()
	s.currentSession.TotalFiles = len(epubFiles)
	s.sessionMutex.Unlock()

	// Process each EPUB file
	for _, filePath := range epubFiles {
		s.processFile(session, filePath)
	}
}

// scanForEPUBFiles recursively scans a directory for EPUB files
func (s *ImportService) scanForEPUBFiles(rootPath string) ([]string, error) {
	var epubFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if !info.IsDir() && filepath.Ext(path) == ".epub" {
			epubFiles = append(epubFiles, path)
		}

		return nil
	})

	return epubFiles, err
}

// processFile processes a single EPUB file
func (s *ImportService) processFile(session *ImportSession, filePath string) {
	// Always increment processed files at the start - this file is being processed
	s.incrementProcessed(session)

	// Extract metadata
	bookMetadata, err := s.metadataExtractor.ExtractMetadata(filePath)
	if err != nil {
		s.logError(session, fmt.Sprintf("Failed to extract metadata from %s: %v", filePath, err))
		s.quarantineFile(session, filePath, "metadata extraction failed")
		return
	}

	// Check if we have required metadata
	if bookMetadata.Title == "" || bookMetadata.Author == "" {
		s.logError(session, fmt.Sprintf("Missing required metadata (title or author) in %s", filePath))
		s.quarantineFile(session, filePath, "missing title or author")
		return
	}

	// Create target directory structure
	targetDir := filepath.Join(s.config.ScanDirectory, bookMetadata.Author, bookMetadata.Title)
	targetFile := filepath.Join(targetDir, fmt.Sprintf("%s - %s.epub", bookMetadata.Title, bookMetadata.Author))

	// Check if file already exists
	if _, err := os.Stat(targetFile); err == nil {
		s.logError(session, fmt.Sprintf("File already exists, skipping: %s", targetFile))
		s.incrementSkipped(session)
		return
	}

	if session.DryRun {
		// Dry run - just log what would happen
		s.logInfo(session, fmt.Sprintf("Would import: %s -> %s", filePath, targetFile))
		return
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		s.logError(session, fmt.Sprintf("Failed to create target directory %s: %v", targetDir, err))
		return
	}

	// Copy file to target location
	if err := s.copyFile(filePath, targetFile); err != nil {
		s.logError(session, fmt.Sprintf("Failed to copy file %s to %s: %v", filePath, targetFile, err))
		return
	}

	s.logInfo(session, fmt.Sprintf("Imported: %s -> %s", filePath, targetFile))
	s.incrementImported(session)
}

// copyFile copies a file from source to destination
func (s *ImportService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// quarantineFile moves a file to the quarantine directory
func (s *ImportService) quarantineFile(session *ImportSession, filePath, reason string) {
	if session.DryRun {
		s.logInfo(session, fmt.Sprintf("Would quarantine %s (reason: %s)", filePath, reason))
		s.incrementQuarantined(session)
		return
	}

	// Ensure quarantine directory exists
	if err := os.MkdirAll(s.config.QuarantineDirectory, 0755); err != nil {
		s.logError(session, fmt.Sprintf("Failed to create quarantine directory: %v", err))
		return
	}

	// Generate quarantine filename
	baseName := filepath.Base(filePath)
	quarantinePath := filepath.Join(s.config.QuarantineDirectory, baseName)

	// Copy to quarantine
	if err := s.copyFile(filePath, quarantinePath); err != nil {
		s.logError(session, fmt.Sprintf("Failed to quarantine file %s: %v", filePath, err))
		return
	}

	// Add to quarantined books list
	s.addQuarantinedBook(session, filePath, quarantinePath, reason)

	s.logInfo(session, fmt.Sprintf("Quarantined: %s (reason: %s)", filePath, reason))
	s.incrementQuarantined(session)
}

// addQuarantinedBook adds a quarantined book to the session's quarantined books list
func (s *ImportService) addQuarantinedBook(session *ImportSession, filePath, quarantinePath, reason string) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	// Extract error detail from the most recent error message
	errorDetail := ""
	if len(session.Errors) > 0 {
		// Get the last error message for this file
		for i := len(session.Errors) - 1; i >= 0; i-- {
			if strings.Contains(session.Errors[i], filePath) {
				// Extract the error detail part after the colon
				parts := strings.Split(session.Errors[i], ": ")
				if len(parts) > 1 {
					errorDetail = parts[1]
				}
				break
			}
		}
	}

	quarantinedBook := QuarantinedBook{
		FilePath:       filePath,
		QuarantinePath: quarantinePath,
		Reason:         reason,
		ErrorDetail:    errorDetail,
		Timestamp:      time.Now(),
	}

	session.QuarantinedBooks = append(session.QuarantinedBooks, quarantinedBook)
}

// Helper methods for updating session counters
func (s *ImportService) incrementProcessed(session *ImportSession) {
	s.sessionMutex.Lock()
	s.currentSession.ProcessedFiles++
	s.sessionMutex.Unlock()
}

func (s *ImportService) incrementImported(session *ImportSession) {
	s.sessionMutex.Lock()
	s.currentSession.ImportedFiles++
	s.sessionMutex.Unlock()
}

func (s *ImportService) incrementQuarantined(session *ImportSession) {
	s.sessionMutex.Lock()
	s.currentSession.QuarantinedFiles++
	s.sessionMutex.Unlock()
}

func (s *ImportService) incrementSkipped(session *ImportSession) {
	s.sessionMutex.Lock()
	s.currentSession.SkippedFiles++
	s.sessionMutex.Unlock()
}

// Logging methods
func (s *ImportService) logError(session *ImportSession, message string) {
	s.sessionMutex.Lock()
	s.currentSession.Errors = append(s.currentSession.Errors, message)
	s.sessionMutex.Unlock()
	log.Printf("[%s] ERROR: %s", session.ID, message)
}

func (s *ImportService) logInfo(session *ImportSession, message string) {
	log.Printf("[%s] INFO: %s", session.ID, message)
}

// saveSessionLog saves the session log to disk
func (s *ImportService) saveSessionLog(session *ImportSession) {
	// Ensure log directory exists
	if err := os.MkdirAll(s.logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return
	}

	// Write session log
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal session log: %v", err)
		return
	}

	if err := ioutil.WriteFile(session.LogPath, data, 0644); err != nil {
		log.Printf("Failed to write session log: %v", err)
		return
	}

	// Clean up old logs if we exceed max logs
	s.cleanupOldLogs()
}

// GetAvailableLogs returns a list of available import session logs
func (s *ImportService) GetAvailableLogs() ([]map[string]interface{}, error) {
	files, err := ioutil.ReadDir(s.logDir)
	if err != nil {
		return nil, err
	}

	var logs []map[string]interface{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			// Read the log file to get basic info
			logPath := filepath.Join(s.logDir, file.Name())
			data, err := ioutil.ReadFile(logPath)
			if err != nil {
				continue // Skip corrupted files
			}

			var session ImportSession
			if err := json.Unmarshal(data, &session); err != nil {
				continue // Skip invalid JSON
			}

			logs = append(logs, map[string]interface{}{
				"session_id":        session.ID,
				"start_time":        session.StartTime,
				"end_time":          session.EndTime,
				"status":            session.Status,
				"dry_run":           session.DryRun,
				"total_files":       session.TotalFiles,
				"imported_files":    session.ImportedFiles,
				"quarantined_files": session.QuarantinedFiles,
				"skipped_files":     session.SkippedFiles,
				"modified":          file.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	// Simple bubble sort for now
	for i := 0; i < len(logs)-1; i++ {
		for j := 0; j < len(logs)-i-1; j++ {
			time1 := logs[j]["modified"].(time.Time)
			time2 := logs[j+1]["modified"].(time.Time)
			if time1.Before(time2) {
				logs[j], logs[j+1] = logs[j+1], logs[j]
			}
		}
	}

	return logs, nil
}

// GetLog returns a specific import session log
func (s *ImportService) GetLog(sessionID string) (*ImportSession, error) {
	logPath := filepath.Join(s.logDir, sessionID+".json")
	data, err := ioutil.ReadFile(logPath)
	if err != nil {
		return nil, err
	}

	var session ImportSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// cleanupOldLogs removes old session logs to maintain the max log count
func (s *ImportService) cleanupOldLogs() {
	files, err := ioutil.ReadDir(s.logDir)
	if err != nil {
		return
	}

	// Sort by modification time (oldest first)
	var logFiles []os.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			logFiles = append(logFiles, file)
		}
	}

	// Remove oldest logs if we exceed the limit
	if len(logFiles) > s.maxLogs {
		// Sort by modification time (oldest first)
		for i := 0; i < len(logFiles)-s.maxLogs; i++ {
			oldLogPath := filepath.Join(s.logDir, logFiles[i].Name())
			os.Remove(oldLogPath)
		}
	}
}
