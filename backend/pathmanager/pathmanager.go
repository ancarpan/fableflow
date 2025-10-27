package pathmanager

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// PathConfig holds configuration for path generation
type PathConfig struct {
	BaseDirectory    string
	AuthorSubdir     bool
	TitleSubdir      bool
	FilenamePattern  string // "{title} - {author}.{format}"
	CleanFilenames   bool   // Whether to clean filenames for filesystem compatibility
	UseAuthorHashing bool   // Whether to use first letter hashing for author directories
}

// DefaultPathConfig returns a sensible default configuration
func DefaultPathConfig(baseDir string) *PathConfig {
	return &PathConfig{
		BaseDirectory:    baseDir,
		AuthorSubdir:     true,
		TitleSubdir:      true,
		FilenamePattern:  "{title} - {author}.{format}",
		CleanFilenames:   false, // Don't clean filenames by default
		UseAuthorHashing: true,  // Use first letter hashing for author directories
	}
}

// DefaultPathConfigNoCleaning returns a configuration without filename cleaning
func DefaultPathConfigNoCleaning(baseDir string) *PathConfig {
	return &PathConfig{
		BaseDirectory:    baseDir,
		AuthorSubdir:     true,
		TitleSubdir:      true,
		FilenamePattern:  "{title} - {author}.{format}",
		CleanFilenames:   false, // Don't clean filenames
		UseAuthorHashing: true,  // Use first letter hashing for author directories
	}
}

// PathManager handles book path generation
type PathManager struct {
	config *PathConfig
}

// NewPathManager creates a new PathManager with the given configuration
func NewPathManager(config *PathConfig) *PathManager {
	return &PathManager{config: config}
}

// GenerateBookPath creates a complete file path for a book
func (p *PathManager) GenerateBookPath(author, title, format string) (string, error) {
	if author == "" || title == "" {
		return "", errors.New("author and title cannot be empty")
	}

	// Clean the inputs if configured to do so
	var cleanAuthor, cleanTitle string
	if p.config.CleanFilenames {
		cleanAuthor = p.CleanForFilesystem(author)
		cleanTitle = p.CleanForFilesystem(title)

		// Validate after cleaning
		if cleanAuthor == "" || cleanTitle == "" {
			return "", errors.New("author and title cannot be empty after cleaning")
		}
	} else {
		cleanAuthor = author
		cleanTitle = title
	}

	// Generate directory path
	var dirPath string
	if p.config.AuthorSubdir && p.config.TitleSubdir {
		if p.config.UseAuthorHashing {
			// Use first letter hashing for author directory
			authorHash := p.getFirstLetterHash(cleanAuthor)
			dirPath = filepath.Join(p.config.BaseDirectory, authorHash, cleanAuthor, cleanTitle)
		} else {
			// Original flat structure
			dirPath = filepath.Join(p.config.BaseDirectory, cleanAuthor, cleanTitle)
		}
	} else if p.config.AuthorSubdir {
		if p.config.UseAuthorHashing {
			// Use first letter hashing for author directory
			authorHash := p.getFirstLetterHash(cleanAuthor)
			dirPath = filepath.Join(p.config.BaseDirectory, authorHash, cleanAuthor)
		} else {
			// Original flat structure
			dirPath = filepath.Join(p.config.BaseDirectory, cleanAuthor)
		}
	} else {
		dirPath = p.config.BaseDirectory
	}

	// Generate filename
	filename := p.generateFilename(cleanTitle, cleanAuthor, format)

	// Join directory and filename
	fullPath := filepath.Join(dirPath, filename)

	return fullPath, nil
}

// generateFilename creates a filename based on the configured pattern
func (p *PathManager) generateFilename(title, author, format string) string {
	// Use the configured pattern
	filename := strings.ReplaceAll(p.config.FilenamePattern, "{title}", title)
	filename = strings.ReplaceAll(filename, "{author}", author)
	filename = strings.ReplaceAll(filename, "{format}", format)

	return filename
}

// CleanForFilesystem removes or replaces characters that are invalid for filesystem paths
func (p *PathManager) CleanForFilesystem(s string) string {
	if s == "" {
		return "Unknown"
	}

	// Normalize unicode (é -> e, ñ -> n, etc.)
	s = p.normalizeUnicode(s)

	// Remove invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := s
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "")
	}

	// Replace multiple spaces with single space
	result = strings.Join(strings.Fields(result), " ")

	// Trim whitespace
	result = strings.TrimSpace(result)

	// Ensure it's not empty
	if result == "" {
		return "Unknown"
	}

	return result
}

// normalizeUnicode normalizes unicode characters for better filesystem compatibility
func (p *PathManager) normalizeUnicode(s string) string {
	// Create a transformer that removes diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// Apply the transformation
	result, _, err := transform.String(t, s)
	if err != nil {
		// If transformation fails, return original string
		return s
	}

	return result
}

// GetRelativePath returns the path relative to the base directory
func (p *PathManager) GetRelativePath(fullPath, baseDir string) string {
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return fullPath
	}
	return relPath
}

// ValidatePath checks if a path is valid and safe
func (p *PathManager) ValidatePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return errors.New("path contains invalid traversal")
	}

	return nil
}

// getFirstLetterHash returns the first letter of a string, uppercased, for directory hashing
func (p *PathManager) getFirstLetterHash(s string) string {
	if s == "" {
		return "Unknown"
	}

	// Get the first character
	firstChar := strings.ToUpper(string(s[0]))

	// If it's not a letter, put it in a special directory
	if !unicode.IsLetter(rune(firstChar[0])) {
		return "0-9"
	}

	return firstChar
}

// GenerateDuplicatePath creates a path for a duplicate book by adding a number
func (p *PathManager) GenerateDuplicatePath(author, title, format string, duplicateNumber int) (string, error) {
	if duplicateNumber < 2 {
		return "", errors.New("duplicate number must be 2 or greater")
	}

	// Clean the inputs
	cleanAuthor := p.CleanForFilesystem(author)
	cleanTitle := p.CleanForFilesystem(title)

	// Add number to title
	numberedTitle := fmt.Sprintf("%s (%d)", cleanTitle, duplicateNumber)

	// Generate path with numbered title
	return p.GenerateBookPath(cleanAuthor, numberedTitle, format)
}
