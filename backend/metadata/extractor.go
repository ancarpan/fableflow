package metadata

import (
	"archive/zip"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"fableflow/backend/conversion"
)

// Note: OPF and Metadata types are now imported from conversion package

// BookMetadata represents extracted book metadata
type BookMetadata struct {
	Title       string
	Author      string
	Publisher   string
	Language    string
	Description string
	ISBN        string
	Date        string
	Subject     string
	Rights      string
}

// Extractor handles metadata extraction from various ebook formats
type Extractor struct{}

// NewExtractor creates a new metadata extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractMetadata extracts metadata from an ebook file
func (e *Extractor) ExtractMetadata(filePath string) (*BookMetadata, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return e.extractEPUBMetadata(filePath)
	case ".pdf":
		return e.extractPDFMetadata(filePath)
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}
}

// extractEPUBMetadata extracts metadata from EPUB files using smart OPF finding
func (e *Extractor) extractEPUBMetadata(filePath string) (*BookMetadata, error) {
	// EPUB files are ZIP archives
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB as ZIP: %v", err)
	}
	defer reader.Close()

	// Use smart OPF finding logic from conversion package
	parser := conversion.NewEPUBParser()
	opfFile, err := parser.FindOPFFile(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to find OPF file: %v", err)
	}

	// Parse the OPF file using conversion package
	opf, err := parser.ParseOPF(opfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF file: %v", err)
	}

	// Convert to BookMetadata format
	metadata := e.convertOPFToBookMetadata(opf)

	// Fallback to filename if no title found
	if metadata.Title == "" {
		filename := filepath.Base(filePath)
		metadata.Title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	log.Printf("Extracted EPUB metadata - Title: %s, Author: %s", metadata.Title, metadata.Author)
	return metadata, nil
}

// extractPDFMetadata extracts metadata from PDF files
func (e *Extractor) extractPDFMetadata(filePath string) (*BookMetadata, error) {
	// For now, PDF metadata extraction is not implemented
	// Fallback to filename parsing
	log.Printf("PDF metadata extraction not implemented, using filename parsing for: %s", filePath)
	return e.ExtractFromFilename(filePath), nil
}

// ExtractFromFilename is a fallback method that tries to parse metadata from filename
func (e *Extractor) ExtractFromFilename(filePath string) *BookMetadata {
	filename := filepath.Base(filePath)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	metadata := &BookMetadata{
		Title:  nameWithoutExt,
		Author: "Unknown",
	}

	// Try to extract author and title from filename patterns
	if strings.Contains(nameWithoutExt, " - ") {
		// Remove any ID suffix (e.g., _1234)
		baseName := nameWithoutExt
		if lastUnderscore := strings.LastIndex(nameWithoutExt, "_"); lastUnderscore != -1 {
			baseName = nameWithoutExt[:lastUnderscore]
		}

		// Split by " - " to get title and author
		if strings.Contains(baseName, " - ") {
			parts := strings.SplitN(baseName, " - ", 2)
			if len(parts) == 2 {
				metadata.Title = strings.TrimSpace(parts[0])
				metadata.Author = strings.TrimSpace(parts[1])
			}
		}
	}

	return metadata
}

// convertOPFToBookMetadata converts conversion package OPF to BookMetadata
func (e *Extractor) convertOPFToBookMetadata(opf *conversion.OPF) *BookMetadata {
	metadata := &BookMetadata{}

	// Extract metadata from OPF
	if len(opf.Metadata.Title) > 0 {
		metadata.Title = strings.TrimSpace(opf.Metadata.Title[0])
	}
	if len(opf.Metadata.Creator) > 0 {
		metadata.Author = strings.TrimSpace(opf.Metadata.Creator[0])
	}
	if len(opf.Metadata.Publisher) > 0 {
		metadata.Publisher = strings.TrimSpace(opf.Metadata.Publisher[0])
	}
	if len(opf.Metadata.Language) > 0 {
		metadata.Language = strings.TrimSpace(opf.Metadata.Language[0])
	}
	if len(opf.Metadata.Description) > 0 {
		metadata.Description = strings.TrimSpace(opf.Metadata.Description[0])
	}
	if len(opf.Metadata.Date) > 0 {
		metadata.Date = strings.TrimSpace(opf.Metadata.Date[0])
	}
	if len(opf.Metadata.Subject) > 0 {
		metadata.Subject = strings.TrimSpace(opf.Metadata.Subject[0])
	}
	if len(opf.Metadata.Rights) > 0 {
		metadata.Rights = strings.TrimSpace(opf.Metadata.Rights[0])
	}

	// Fallback to "Unknown" if no author found
	if metadata.Author == "" {
		metadata.Author = "Unknown"
	}

	return metadata
}

// isISBN checks if a string looks like an ISBN number
func isISBN(identifier string) bool {
	// Remove common prefixes and clean the string
	clean := strings.ToLower(identifier)
	clean = strings.TrimPrefix(clean, "isbn:")
	clean = strings.TrimPrefix(clean, "urn:isbn:")
	clean = strings.TrimSpace(clean)

	// Remove hyphens and spaces
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")

	// Check if it's a 10 or 13 digit number
	if len(clean) == 10 || len(clean) == 13 {
		// Check if all characters are digits
		matched, _ := regexp.MatchString(`^\d+$`, clean)
		return matched
	}

	// Check if it contains ISBN-like pattern with hyphens (e.g., 978-1-234-56789-0)
	isbnPattern := regexp.MustCompile(`^\d{3}-\d{1}-\d{3}-\d{5}-\d{1}$|^\d{1}-\d{3}-\d{5}-\d{1}$`)
	return isbnPattern.MatchString(identifier)
}
