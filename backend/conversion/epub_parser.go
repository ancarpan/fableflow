package conversion

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// EPUBBook represents the parsed content of an EPUB file
type EPUBBook struct {
	Title       string
	Author      string
	Language    string
	Description string
	Chapters    []Chapter
	CSS         []string
	Images      map[string][]byte
	CoverImage  []byte
	CoverPath   string
}

// Chapter represents a single chapter in the EPUB
type Chapter struct {
	Title   string
	Content string
	Order   int
}

// OPF represents the structure of an EPUB OPF file
type OPF struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
	Manifest Manifest `xml:"manifest"`
	Spine    Spine    `xml:"spine"`
}

// Metadata represents the metadata section of an OPF file
type Metadata struct {
	Title       []string `xml:"title"`
	Creator     []string `xml:"creator"`
	Language    []string `xml:"language"`
	Description []string `xml:"description"`
	Publisher   []string `xml:"publisher"`
	Date        []string `xml:"date"`
	Subject     []string `xml:"subject"`
	Rights      []string `xml:"rights"`
}

// Manifest represents the manifest section of an OPF file
type Manifest struct {
	Items []Item `xml:"item"`
}

// Item represents an item in the manifest
type Item struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

// Spine represents the spine section of an OPF file
type Spine struct {
	ItemRefs []ItemRef `xml:"itemref"`
}

// ItemRef represents a reference to an item in the spine
type ItemRef struct {
	IDRef string `xml:"idref,attr"`
}

// EPUBParser handles parsing of EPUB files
type EPUBParser struct{}

// NewEPUBParser creates a new EPUB parser
func NewEPUBParser() *EPUBParser {
	return &EPUBParser{}
}

// ParseEPUB parses an EPUB file and extracts its content
func (p *EPUBParser) ParseEPUB(filePath string) (*EPUBBook, error) {
	// Open EPUB file (which is a ZIP archive)
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %v", err)
	}
	defer reader.Close()

	book := &EPUBBook{
		Images: make(map[string][]byte),
	}

	// Find and parse the OPF file
	opfFile, err := p.FindOPFFile(reader)
	if err != nil {
		fmt.Printf("Warning: Could not find OPF file: %v\n", err)
		// Fallback to filename-based metadata
		p.extractMetadataFromFilename(filePath, book)
	} else {
		// Parse OPF content
		opf, err := p.ParseOPF(opfFile)
		if err != nil {
			fmt.Printf("Warning: Could not parse OPF file: %v\n", err)
			// Fallback to filename-based metadata
			p.extractMetadataFromFilename(filePath, book)
		} else {
			// Extract metadata from OPF
			p.extractMetadata(opf, book)
		}
	}

	// Extract cover image
	p.extractCoverImage(reader, book)

	// Extract content - try OPF-based extraction first, then fallback
	if opfFile != nil {
		opf, err := p.ParseOPF(opfFile)
		if err == nil {
			err = p.extractContent(reader, opf, book)
			if err != nil {
				fmt.Printf("Warning: OPF-based content extraction failed: %v\n", err)
				// Fallback to simple content extraction
				p.extractContentSimple(reader, book)
			}
		} else {
			// Fallback to simple content extraction
			p.extractContentSimple(reader, book)
		}
	} else {
		// Fallback to simple content extraction
		p.extractContentSimple(reader, book)
	}

	return book, nil
}

// FindOPFFile locates the OPF file in the EPUB (public method for reuse)
func (p *EPUBParser) FindOPFFile(reader *zip.ReadCloser) (*zip.File, error) {
	// First, try to find META-INF/container.xml to locate the OPF file
	containerFile, err := p.findContainerFile(reader)
	if err == nil {
		opfPath, err := p.parseContainerFile(containerFile)
		if err == nil {
			// Look for the OPF file specified in container.xml
			for _, file := range reader.File {
				if file.Name == opfPath || strings.HasSuffix(file.Name, opfPath) {
					return file, nil
				}
			}
		}
	}

	// Fallback: Look for common OPF file names
	opfNames := []string{"content.opf", "package.opf", "book.opf", "metadata.opf"}
	for _, name := range opfNames {
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, name) {
				return file, nil
			}
		}
	}

	// Last resort: Look for any .opf file
	for _, file := range reader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".opf") {
			return file, nil
		}
	}

	return nil, fmt.Errorf("OPF file not found in EPUB")
}

// findContainerFile locates the META-INF/container.xml file
func (p *EPUBParser) findContainerFile(reader *zip.ReadCloser) (*zip.File, error) {
	for _, file := range reader.File {
		if file.Name == "META-INF/container.xml" {
			return file, nil
		}
	}
	return nil, fmt.Errorf("container.xml not found")
}

// parseContainerFile parses the container.xml to find the OPF file path
func (p *EPUBParser) parseContainerFile(containerFile *zip.File) (string, error) {
	rc, err := containerFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	// Simple XML parsing to find the OPF file path
	// Look for <rootfile full-path="..." media-type="application/oebps-package+xml"/>
	contentStr := string(content)

	// Find the rootfile element
	start := strings.Index(contentStr, "<rootfile")
	if start == -1 {
		return "", fmt.Errorf("rootfile element not found in container.xml")
	}

	end := strings.Index(contentStr[start:], "/>")
	if end == -1 {
		return "", fmt.Errorf("rootfile element not properly closed")
	}

	rootfileElement := contentStr[start : start+end+2]

	// Extract full-path attribute
	fullPathStart := strings.Index(rootfileElement, `full-path="`)
	if fullPathStart == -1 {
		return "", fmt.Errorf("full-path attribute not found")
	}

	fullPathStart += len(`full-path="`)
	fullPathEnd := strings.Index(rootfileElement[fullPathStart:], `"`)
	if fullPathEnd == -1 {
		return "", fmt.Errorf("full-path attribute not properly quoted")
	}

	return rootfileElement[fullPathStart : fullPathStart+fullPathEnd], nil
}

// ParseOPF parses the OPF XML content (public method for reuse)
func (p *EPUBParser) ParseOPF(opfFile *zip.File) (*OPF, error) {
	rc, err := opfFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open OPF file: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read OPF file: %v", err)
	}

	var opf OPF
	err = xml.Unmarshal(content, &opf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OPF XML: %v", err)
	}

	return &opf, nil
}

// extractMetadata extracts metadata from OPF
func (p *EPUBParser) extractMetadata(opf *OPF, book *EPUBBook) {
	if len(opf.Metadata.Title) > 0 {
		book.Title = strings.TrimSpace(opf.Metadata.Title[0])
	}
	if len(opf.Metadata.Creator) > 0 {
		book.Author = strings.TrimSpace(opf.Metadata.Creator[0])
	}
	if len(opf.Metadata.Language) > 0 {
		book.Language = strings.TrimSpace(opf.Metadata.Language[0])
	}
	if len(opf.Metadata.Description) > 0 {
		book.Description = strings.TrimSpace(opf.Metadata.Description[0])
	}

	// Fallback to filename if no title found
	if book.Title == "" {
		book.Title = "Unknown Title"
	}
	if book.Author == "" {
		book.Author = "Unknown Author"
	}
}

// extractContent extracts content from EPUB based on spine order
func (p *EPUBParser) extractContent(reader *zip.ReadCloser, opf *OPF, book *EPUBBook) error {
	// Create a map of items by ID for quick lookup
	itemMap := make(map[string]Item)
	for _, item := range opf.Manifest.Items {
		itemMap[item.ID] = item
	}

	// Process spine items in order
	for i, itemRef := range opf.Spine.ItemRefs {
		item, exists := itemMap[itemRef.IDRef]
		if !exists {
			continue
		}

		// Handle different media types
		switch {
		case strings.Contains(item.MediaType, "html") || strings.HasSuffix(item.Href, ".html") || strings.HasSuffix(item.Href, ".xhtml"):
			content, err := p.extractHTMLContent(reader, item.Href)
			if err != nil {
				continue // Skip problematic files
			}

			chapter := Chapter{
				Title:   p.extractTitleFromHTML(content),
				Content: content,
				Order:   i,
			}
			book.Chapters = append(book.Chapters, chapter)

		case strings.Contains(item.MediaType, "css"):
			css, err := p.extractCSSContent(reader, item.Href)
			if err == nil {
				book.CSS = append(book.CSS, css)
			}

		case strings.Contains(item.MediaType, "image"):
			imageData, err := p.extractImageContent(reader, item.Href)
			if err == nil {
				book.Images[item.Href] = imageData
			}
		}
	}

	return nil
}

// extractHTMLContent extracts HTML content from a file
func (p *EPUBParser) extractHTMLContent(reader *zip.ReadCloser, href string) (string, error) {
	for _, file := range reader.File {
		if file.Name == href || strings.HasSuffix(file.Name, href) {
			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}

			return string(content), nil
		}
	}
	return "", fmt.Errorf("HTML file not found: %s", href)
}

// extractCSSContent extracts CSS content from a file
func (p *EPUBParser) extractCSSContent(reader *zip.ReadCloser, href string) (string, error) {
	for _, file := range reader.File {
		if file.Name == href || strings.HasSuffix(file.Name, href) {
			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}

			return string(content), nil
		}
	}
	return "", fmt.Errorf("CSS file not found: %s", href)
}

// extractImageContent extracts image content from a file
func (p *EPUBParser) extractImageContent(reader *zip.ReadCloser, href string) ([]byte, error) {
	for _, file := range reader.File {
		if file.Name == href || strings.HasSuffix(file.Name, href) {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			return content, nil
		}
	}
	return nil, fmt.Errorf("image file not found: %s", href)
}

// extractTitleFromHTML extracts title from HTML content
func (p *EPUBParser) extractTitleFromHTML(html string) string {
	// Simple title extraction - look for <title> tag or first <h1>
	// This is a basic implementation
	if strings.Contains(html, "<title>") {
		start := strings.Index(html, "<title>")
		end := strings.Index(html, "</title>")
		if start != -1 && end != -1 && end > start {
			title := html[start+7 : end]
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}

	// Fallback to first h1
	if strings.Contains(html, "<h1") {
		start := strings.Index(html, "<h1")
		if start != -1 {
			start = strings.Index(html[start:], ">")
			if start != -1 {
				start += len(html[:strings.Index(html, "<h1")]) + start + 1
				end := strings.Index(html[start:], "</h1>")
				if end != -1 {
					title := html[start : start+end]
					title = strings.TrimSpace(title)
					if title != "" {
						return title
					}
				}
			}
		}
	}

	return "Chapter"
}

// extractMetadataFromFilename extracts metadata from the filename as fallback
func (p *EPUBParser) extractMetadataFromFilename(filePath string, book *EPUBBook) {
	filename := filepath.Base(filePath)
	// Remove extension
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try to parse "Author - Title" format
	if strings.Contains(nameWithoutExt, " - ") {
		parts := strings.Split(nameWithoutExt, " - ")
		if len(parts) >= 2 {
			book.Author = strings.TrimSpace(parts[0])
			book.Title = strings.TrimSpace(parts[1])
		} else {
			book.Title = nameWithoutExt
			book.Author = "Unknown Author"
		}
	} else {
		book.Title = nameWithoutExt
		book.Author = "Unknown Author"
	}

	book.Language = "en" // Default language
	book.Description = ""
}

// extractCoverImage extracts the cover image from the EPUB
func (p *EPUBParser) extractCoverImage(reader *zip.ReadCloser, book *EPUBBook) {
	// First, try to find cover from OPF metadata
	opfFile, err := p.FindOPFFile(reader)
	if err == nil {
		opf, err := p.ParseOPF(opfFile)
		if err == nil {
			// Look for cover metadata in OPF
			for _, item := range opf.Manifest.Items {
				if item.ID == "cover" || strings.Contains(strings.ToLower(item.Href), "cover") {
					imageData, err := p.extractImageContent(reader, item.Href)
					if err == nil && len(imageData) > 0 {
						book.CoverImage = imageData
						book.CoverPath = item.Href
						fmt.Printf("Found cover image from OPF: %s (%d bytes)\n", item.Href, len(imageData))
						return
					}
				}
			}
		}
	}

	// Look for common cover image filenames
	coverNames := []string{
		"cover.jpg", "cover.jpeg", "cover.png", "cover.gif",
		"cover-image.jpg", "cover-image.jpeg", "cover-image.png",
		"titlepage.jpg", "titlepage.jpeg", "titlepage.png",
		"front-cover.jpg", "front-cover.jpeg", "front-cover.png",
		"immagine_png.jpeg", "immagine_png.jpg", // Italian naming
	}

	// Search for cover images
	for _, coverName := range coverNames {
		for _, file := range reader.File {
			if strings.HasSuffix(strings.ToLower(file.Name), coverName) {
				imageData, err := p.extractImageContent(reader, file.Name)
				if err == nil && len(imageData) > 0 {
					book.CoverImage = imageData
					book.CoverPath = file.Name
					fmt.Printf("Found cover image: %s (%d bytes)\n", file.Name, len(imageData))
					return
				}
			}
		}
	}

	// Fallback: Look for any image in images directory
	for _, file := range reader.File {
		if strings.Contains(strings.ToLower(file.Name), "images/") {
			ext := strings.ToLower(filepath.Ext(file.Name))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
				imageData, err := p.extractImageContent(reader, file.Name)
				if err == nil && len(imageData) > 0 {
					book.CoverImage = imageData
					book.CoverPath = file.Name
					fmt.Printf("Found potential cover image: %s (%d bytes)\n", file.Name, len(imageData))
					return
				}
			}
		}
	}

	fmt.Printf("No cover image found\n")
}

// extractContentSimple provides a fallback content extraction when OPF parsing fails
func (p *EPUBParser) extractContentSimple(reader *zip.ReadCloser, book *EPUBBook) {
	// Look for HTML files in the EPUB
	htmlFiles := make([]*zip.File, 0)
	for _, file := range reader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".html") ||
			strings.HasSuffix(strings.ToLower(file.Name), ".xhtml") {
			htmlFiles = append(htmlFiles, file)
		}
	}

	// Sort files by name to get a consistent order
	sort.Slice(htmlFiles, func(i, j int) bool {
		return htmlFiles[i].Name < htmlFiles[j].Name
	})

	// Extract content from HTML files
	for i, file := range htmlFiles {
		content, err := p.extractHTMLContent(reader, file.Name)
		if err != nil {
			continue // Skip problematic files
		}

		chapter := Chapter{
			Title:   p.extractTitleFromHTML(content),
			Content: content,
			Order:   i,
		}
		book.Chapters = append(book.Chapters, chapter)
	}

	// If no chapters were found, create a single chapter with basic content
	if len(book.Chapters) == 0 {
		chapter := Chapter{
			Title:   book.Title,
			Content: fmt.Sprintf("<h1>%s</h1><p>by %s</p><p>Content could not be extracted from this EPUB file.</p>", book.Title, book.Author),
			Order:   0,
		}
		book.Chapters = append(book.Chapters, chapter)
	}
}
