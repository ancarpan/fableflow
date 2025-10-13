package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// EPUBEditor handles loading, editing, and saving EPUB files
type EPUBEditor struct {
	filePath string
	opfData  *OPFDocument
	zipFiles map[string][]byte // Store all files from the EPUB
}

// OPFDocument represents the structure of an EPUB OPF file
type OPFDocument struct {
	XMLName  xml.Name `xml:"package"`
	Version  string   `xml:"version,attr"`
	Metadata Metadata `xml:"metadata"`
	Manifest Manifest `xml:"manifest"`
	Spine    Spine    `xml:"spine"`
}

// Metadata represents the metadata section of an OPF file
type Metadata struct {
	Title       []DCElement `xml:"title"`
	Creator     []DCElement `xml:"creator"`
	Language    []DCElement `xml:"language"`
	Description []DCElement `xml:"description"`
	Publisher   []DCElement `xml:"publisher"`
	Date        []DCElement `xml:"date"`
	Subject     []DCElement `xml:"subject"`
	Rights      []DCElement `xml:"rights"`
	Identifier  []DCElement `xml:"identifier"`
}

// DCElement represents a Dublin Core element with optional attributes
type DCElement struct {
	XMLName xml.Name `xml:""`
	Value   string   `xml:",chardata"`
	ID      string   `xml:"id,attr,omitempty"`
	Scheme  string   `xml:"scheme,attr,omitempty"`
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

// NewEPUBEditor creates a new EPUB editor
func NewEPUBEditor(filePath string) *EPUBEditor {
	return &EPUBEditor{
		filePath: filePath,
		zipFiles: make(map[string][]byte),
	}
}

// Load loads an existing EPUB file for editing
func (e *EPUBEditor) Load() error {
	// Open EPUB file (which is a ZIP archive)
	reader, err := zip.OpenReader(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB file: %v", err)
	}
	defer reader.Close()

	// Store all files from the EPUB
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", file.Name, err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", file.Name, err)
		}

		e.zipFiles[file.Name] = data
	}

	// Find and parse the OPF file
	opfFile, err := e.findOPFFile()
	if err != nil {
		return fmt.Errorf("failed to find OPF file: %v", err)
	}

	// Parse OPF content
	opf, err := e.parseOPF(opfFile)
	if err != nil {
		return fmt.Errorf("failed to parse OPF file: %v", err)
	}

	e.opfData = opf
	return nil
}

// findOPFFile locates the OPF file in the EPUB
func (e *EPUBEditor) findOPFFile() ([]byte, error) {
	// First, try to find META-INF/container.xml
	containerData, exists := e.zipFiles["META-INF/container.xml"]
	if !exists {
		return nil, fmt.Errorf("META-INF/container.xml not found")
	}

	// Parse container.xml to find OPF file
	var container struct {
		XMLName   xml.Name `xml:"container"`
		RootFiles []struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}

	if err := xml.Unmarshal(containerData, &container); err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %v", err)
	}

	if len(container.RootFiles) == 0 {
		return nil, fmt.Errorf("no rootfile found in container.xml")
	}

	opfPath := container.RootFiles[0].FullPath
	opfData, exists := e.zipFiles[opfPath]
	if !exists {
		return nil, fmt.Errorf("OPF file not found: %s", opfPath)
	}

	return opfData, nil
}

// parseOPF parses the OPF XML content
func (e *EPUBEditor) parseOPF(opfData []byte) (*OPFDocument, error) {
	var opf OPFDocument
	if err := xml.Unmarshal(opfData, &opf); err != nil {
		return nil, fmt.Errorf("failed to parse OPF XML: %v", err)
	}
	return &opf, nil
}

// UpdateMetadata updates the metadata in the OPF document
func (e *EPUBEditor) UpdateMetadata(title, author, isbn, publisher string) error {
	if e.opfData == nil {
		return fmt.Errorf("no OPF data loaded")
	}

	// Update title
	if title != "" {
		if len(e.opfData.Metadata.Title) == 0 {
			e.opfData.Metadata.Title = []DCElement{{Value: title}}
		} else {
			e.opfData.Metadata.Title[0].Value = title
		}
	}

	// Update creator (author)
	if author != "" {
		if len(e.opfData.Metadata.Creator) == 0 {
			e.opfData.Metadata.Creator = []DCElement{{Value: author}}
		} else {
			e.opfData.Metadata.Creator[0].Value = author
		}
	}

	// Update publisher
	if publisher != "" {
		if len(e.opfData.Metadata.Publisher) == 0 {
			e.opfData.Metadata.Publisher = []DCElement{{Value: publisher}}
		} else {
			e.opfData.Metadata.Publisher[0].Value = publisher
		}
	}

	// Update ISBN (identifier)
	if isbn != "" {
		// Find existing ISBN identifier
		found := false
		for i, id := range e.opfData.Metadata.Identifier {
			if id.Scheme == "ISBN" || strings.Contains(strings.ToLower(id.Scheme), "isbn") {
				e.opfData.Metadata.Identifier[i].Value = isbn
				found = true
				break
			}
		}

		// If no ISBN identifier found, add one
		if !found {
			e.opfData.Metadata.Identifier = append(e.opfData.Metadata.Identifier, DCElement{
				Value:  isbn,
				ID:     "isbn",
				Scheme: "ISBN",
			})
		}
	}

	return nil
}

// Save saves the modified EPUB file
func (e *EPUBEditor) Save() error {
	if e.opfData == nil {
		return fmt.Errorf("no OPF data loaded")
	}

	// Marshal the updated OPF data
	opfXML, err := xml.MarshalIndent(e.opfData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OPF XML: %v", err)
	}

	// Add XML declaration
	opfXML = []byte(xml.Header + string(opfXML))

	// Update the OPF file in our stored files
	// Find the OPF file path
	containerData := e.zipFiles["META-INF/container.xml"]
	var container struct {
		XMLName   xml.Name `xml:"container"`
		RootFiles []struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}

	if err := xml.Unmarshal(containerData, &container); err != nil {
		return fmt.Errorf("failed to parse container.xml: %v", err)
	}

	opfPath := container.RootFiles[0].FullPath
	e.zipFiles[opfPath] = opfXML

	// Create new EPUB file
	return e.writeEPUB()
}

// writeEPUB writes the EPUB file with all stored files
func (e *EPUBEditor) writeEPUB() error {
	// Create new EPUB file
	file, err := os.Create(e.filePath)
	if err != nil {
		return fmt.Errorf("failed to create EPUB file: %v", err)
	}
	defer file.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Write all files to the new EPUB
	for fileName, fileData := range e.zipFiles {
		writer, err := zipWriter.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file %s in ZIP: %v", fileName, err)
		}

		if _, err := writer.Write(fileData); err != nil {
			return fmt.Errorf("failed to write file %s to ZIP: %v", fileName, err)
		}
	}

	return nil
}

// GetCurrentMetadata returns the current metadata from the EPUB
func (e *EPUBEditor) GetCurrentMetadata() (title, author, isbn, publisher string) {
	if e.opfData == nil {
		return "", "", "", ""
	}

	// Get title
	if len(e.opfData.Metadata.Title) > 0 {
		title = e.opfData.Metadata.Title[0].Value
	}

	// Get author
	if len(e.opfData.Metadata.Creator) > 0 {
		author = e.opfData.Metadata.Creator[0].Value
	}

	// Get publisher
	if len(e.opfData.Metadata.Publisher) > 0 {
		publisher = e.opfData.Metadata.Publisher[0].Value
	}

	// Get ISBN
	for _, id := range e.opfData.Metadata.Identifier {
		if id.Scheme == "ISBN" || strings.Contains(strings.ToLower(id.Scheme), "isbn") {
			isbn = id.Value
			break
		}
	}

	return title, author, isbn, publisher
}
