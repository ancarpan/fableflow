package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConvertEPUBToAZW3 is the main conversion function using Amazon's kindlegen tool.
// This follows FB2Converter's approach for high-quality EPUB to AZW3 conversion.
func ConvertEPUBToAZW3(inputPath, outputPath string) error {
	// Validate input file
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input file not found: %w", err)
	}

	// Check if input is EPUB
	if !strings.HasSuffix(strings.ToLower(inputPath), ".epub") {
		return fmt.Errorf("input file must be an EPUB file")
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use kindlegen for conversion
	converter, err := NewKindlegenConverter()
	if err != nil {
		return fmt.Errorf("failed to create kindlegen converter: %w", err)
	}

	// Enable verbose output for debugging
	converter.SetVerbose(true)

	// Convert EPUB to AZW3 using kindlegen
	if err := converter.ConvertEPUBToAZW3(inputPath, outputPath); err != nil {
		return fmt.Errorf("kindlegen conversion failed: %w", err)
	}

	fmt.Printf("Successfully converted using kindlegen: %s -> %s\n", inputPath, outputPath)
	return nil
}
