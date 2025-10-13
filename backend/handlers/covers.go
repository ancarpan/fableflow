package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"fableflow/backend/database"
)

// OPF document structures for XML parsing
type OPFDocument struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		Meta []MetaTag `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []ManifestItem `xml:"item"`
	} `xml:"manifest"`
}

type MetaTag struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type ManifestItem struct {
	ID   string `xml:"id,attr"`
	Href string `xml:"href,attr"`
}

// CoversHandler handles cover image requests
type CoversHandler struct {
	db *database.Manager
}

// NewCoversHandler creates a new covers handler
func NewCoversHandler(db *database.Manager) *CoversHandler {
	return &CoversHandler{db: db}
}

// ServeCover serves a book's cover image
func (h *CoversHandler) ServeCover(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract book ID from URL path
	idStr := r.URL.Path[len("/api/covers/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Get book details
	book, err := h.db.GetBookByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Check if it's an EPUB file
	if !strings.HasSuffix(strings.ToLower(book.FilePath), ".epub") {
		http.Error(w, "Cover extraction only supported for EPUB files", http.StatusBadRequest)
		return
	}

	// Open the EPUB file
	reader, err := zip.OpenReader(book.FilePath)
	if err != nil {
		http.Error(w, "Failed to open EPUB file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Find cover image
	coverPath, err := h.findCoverInOPF(reader)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cover not found: %v", err), http.StatusNotFound)
		return
	}

	// Read cover image
	coverFile, err := reader.Open(coverPath)
	if err != nil {
		http.Error(w, "Failed to open cover image", http.StatusInternalServerError)
		return
	}
	defer coverFile.Close()

	// Read image data
	imageData, err := io.ReadAll(coverFile)
	if err != nil {
		http.Error(w, "Failed to read cover image", http.StatusInternalServerError)
		return
	}

	// Check for size parameter
	size := r.URL.Query().Get("size")
	if size == "thumbnail" {
		// Generate thumbnail
		thumbnailData, contentType, err := h.generateThumbnail(imageData, "image/jpeg", 200, 280)
		if err != nil {
			http.Error(w, "Failed to generate thumbnail", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(thumbnailData)
		return
	}

	// Serve full image
	contentType := http.DetectContentType(imageData)
	w.Header().Set("Content-Type", contentType)
	w.Write(imageData)
}

// findCoverInOPF finds the cover image path in the OPF file using XML parsing
func (h *CoversHandler) findCoverInOPF(reader *zip.ReadCloser) (string, error) {
	// Find the OPF file
	var opfPath string
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".opf") {
			opfPath = file.Name
			break
		}
	}

	if opfPath == "" {
		return "", fmt.Errorf("no OPF file found")
	}

	// Read and parse the OPF file
	opfFile, err := reader.Open(opfPath)
	if err != nil {
		return "", err
	}
	defer opfFile.Close()

	// Parse XML using Go's standard library
	opfData, err := io.ReadAll(opfFile)
	if err != nil {
		return "", err
	}

	var opf OPFDocument
	if err := xml.Unmarshal(opfData, &opf); err != nil {
		return "", fmt.Errorf("failed to parse OPF XML: %v", err)
	}

	// Step 1: Find cover metadata
	var coverID string
	for _, meta := range opf.Metadata.Meta {
		if meta.Name == "cover" {
			coverID = meta.Content
			fmt.Printf("Found cover metadata: <meta name=\"cover\" content=\"%s\"/>\n", coverID)
			break
		}
	}

	if coverID == "" {
		// Fallback: look for direct cover references in manifest
		for _, item := range opf.Manifest.Items {
			if item.ID == "cover" || strings.Contains(item.ID, "cover") {
				fmt.Printf("Found direct cover reference: %s\n", item.Href)
				// Make path relative to OPF file location
				opfDir := filepath.Dir(opfPath)
				if opfDir != "." {
					return filepath.Join(opfDir, item.Href), nil
				}
				return item.Href, nil
			}
		}
		return "", fmt.Errorf("no cover metadata found in OPF")
	}

	// Step 2: Find manifest item by cover ID
	var coverPath string
	for _, item := range opf.Manifest.Items {
		if item.ID == coverID {
			coverPath = item.Href
			fmt.Printf("Found cover image in manifest: %s\n", coverPath)
			break
		}
	}

	if coverPath == "" {
		return "", fmt.Errorf("cover ID '%s' not found in manifest", coverID)
	}

	// Step 3: Make path relative to OPF file location
	opfDir := filepath.Dir(opfPath)
	if opfDir != "." {
		coverPath = filepath.Join(opfDir, coverPath)
	}

	fmt.Printf("Resolved cover path: %s\n", coverPath)
	return coverPath, nil
}

// generateThumbnail creates a thumbnail version of the image
func (h *CoversHandler) generateThumbnail(imageData []byte, contentType string, maxWidth, maxHeight int) ([]byte, string, error) {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Calculate thumbnail dimensions (maintain aspect ratio)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling factor
	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Resize the image
	resized := h.resizeImage(img, newWidth, newHeight)

	// Encode as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return nil, "", fmt.Errorf("failed to encode thumbnail: %v", err)
	}

	return buf.Bytes(), "image/jpeg", nil
}

// resizeImage resizes an image to the specified dimensions
func (h *CoversHandler) resizeImage(img image.Image, width, height int) image.Image {
	// Simple nearest-neighbor resize
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Create new image
	resized := image.NewRGBA(image.Rect(0, 0, width, height))

	// Scale factors
	xScale := float64(srcWidth) / float64(width)
	yScale := float64(srcHeight) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate source coordinates
			srcX := int(float64(x) * xScale)
			srcY := int(float64(y) * yScale)

			// Clamp to bounds
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			// Copy pixel
			resized.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return resized
}
