package metadata

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// OPF represents the structure of an EPUB OPF file
type OPF struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
}

// Metadata represents the metadata section of an OPF file
type Metadata struct {
	Title       []string `xml:"title"`
	Creator     []string `xml:"creator"`
	Publisher   []string `xml:"publisher"`
	Language    []string `xml:"language"`
	Description []string `xml:"description"`
	Identifier  []string `xml:"identifier"`
	Date        []string `xml:"date"`
	Subject     []string `xml:"subject"`
	Rights      []string `xml:"rights"`
}

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

// extractEPUBMetadata extracts metadata from EPUB files
func (e *Extractor) extractEPUBMetadata(filePath string) (*BookMetadata, error) {
	// EPUB files are ZIP archives
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB as ZIP: %v", err)
	}
	defer reader.Close()

	// Look for content.opf file
	var opfFile *zip.File
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "content.opf") {
			opfFile = file
			break
		}
	}

	if opfFile == nil {
		return nil, fmt.Errorf("content.opf not found in EPUB")
	}

	// Read the OPF file
	rc, err := opfFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open content.opf: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read content.opf: %v", err)
	}

	// Parse the OPF XML
	var opf OPF
	err = xml.Unmarshal(content, &opf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF XML: %v", err)
	}

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
	// Look for ISBN in identifiers (skip UUIDs and other non-ISBN identifiers)
	for _, identifier := range opf.Metadata.Identifier {
		cleanID := strings.TrimSpace(identifier)
		// Check if it looks like an ISBN (contains numbers and hyphens, or is 10/13 digits)
		if isISBN(cleanID) {
			metadata.ISBN = cleanID
			break
		}
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

	// Fallback to filename if no title found
	if metadata.Title == "" {
		filename := filepath.Base(filePath)
		metadata.Title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	// Fallback to "Unknown" if no author found
	if metadata.Author == "" {
		metadata.Author = "Unknown"
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
