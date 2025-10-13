package conversion

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// KindlegenConverter handles EPUB to AZW3 conversion using Amazon's kindlegen tool.
// This follows FB2Converter's approach: kindlegen creates MOBI, then we rename to AZW3.
type KindlegenConverter struct {
	kindlegenPath string
	verbose       bool
}

// NewKindlegenConverter creates a new kindlegen-based converter.
func NewKindlegenConverter() (*KindlegenConverter, error) {
	kindlegenPath, err := GetKindlegenPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get kindlegen path: %w", err)
	}

	return &KindlegenConverter{
		kindlegenPath: kindlegenPath,
		verbose:       false, // Set to true for debugging
	}, nil
}

// ConvertEPUBToAZW3 converts an EPUB file to AZW3 format using kindlegen.
// This follows FB2Converter's approach: kindlegen creates MOBI, then we rename it to AZW3.
func (kc *KindlegenConverter) ConvertEPUBToAZW3(inputPath, outputPath string) error {
	fmt.Printf("KindlegenConverter: Starting conversion %s -> %s\n", inputPath, outputPath)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	fmt.Printf("KindlegenConverter: Output directory: %s\n", outputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate intermediate MOBI file using kindlegen
	fmt.Printf("KindlegenConverter: Generating intermediate MOBI file\n")
	mobiPath, err := kc.generateIntermediateMOBI(inputPath, outputDir)
	if err != nil {
		return fmt.Errorf("failed to generate intermediate MOBI: %w", err)
	}

	// Rename MOBI to AZW3 (AZW3 is essentially MOBI format)
	fmt.Printf("KindlegenConverter: Renaming %s to %s\n", mobiPath, outputPath)
	if err := os.Rename(mobiPath, outputPath); err != nil {
		return fmt.Errorf("failed to rename MOBI to AZW3: %w", err)
	}

	fmt.Printf("KindlegenConverter: Successfully converted %s to %s\n", inputPath, outputPath)
	return nil
}

// generateIntermediateMOBI uses kindlegen to create a MOBI file from EPUB.
// This follows FB2Converter's approach.
func (kc *KindlegenConverter) generateIntermediateMOBI(inputPath, outputDir string) (string, error) {
	// Create output filename (MOBI format) - keep original filename, just change extension
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	mobiFile := baseName + ".mobi"
	mobiPath := filepath.Join(outputDir, mobiFile)

	// Prepare kindlegen arguments
	args := []string{
		inputPath,       // Input EPUB file
		"-c1",           // Compression level 1 (fastest)
		"-locale", "en", // Locale
		"-o", mobiFile, // Output filename only (no path)
	}

	if kc.verbose {
		args = append(args, "-verbose")
	}

	// Create command and set working directory to output directory
	cmd := exec.Command(kc.kindlegenPath, args...)
	cmd.Dir = outputDir // Set working directory so kindlegen creates file there

	fmt.Printf("Running kindlegen: %s %s\n", kc.kindlegenPath, strings.Join(args, " "))
	fmt.Printf("Expected output file: %s\n", mobiPath)

	// Capture stdout for logging
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("unable to redirect kindlegen stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("unable to start kindlegen: %w", err)
	}

	// Read and log kindlegen output
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		fmt.Printf("kindlegen: %s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("kindlegen stdout pipe broken: %w", err)
	}

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if len(ee.Stderr) > 0 {
				fmt.Printf("kindlegen stderr: %s\n", string(ee.Stderr))
			}
			ws := ee.Sys().(syscall.WaitStatus)
			switch ws.ExitStatus() {
			case 1:
				// Warnings - kindlegen sometimes returns 1 for warnings but still succeeds
				fmt.Printf("kindlegen completed with warnings\n")
			case 0:
				// Success
				fmt.Printf("kindlegen completed successfully\n")
			case 2:
				// Error - unable to create mobi
				return "", fmt.Errorf("kindlegen returned error: %w", err)
			default:
				return "", fmt.Errorf("kindlegen returned error: %w", err)
			}
		} else {
			return "", fmt.Errorf("kindlegen returned error: %w", err)
		}
	}

	// Check if MOBI file was created in the expected location
	if _, err := os.Stat(mobiPath); err != nil {
		// Kindlegen might have created the file in the same directory as the input file
		inputDir := filepath.Dir(inputPath)
		actualMobiPath := filepath.Join(inputDir, mobiFile)

		if _, err := os.Stat(actualMobiPath); err == nil {
			fmt.Printf("Found MOBI file in input directory: %s\n", actualMobiPath)
			// Move it to the expected location
			if err := os.Rename(actualMobiPath, mobiPath); err != nil {
				return "", fmt.Errorf("failed to move MOBI file to expected location: %w", err)
			}
			fmt.Printf("Moved MOBI file to: %s\n", mobiPath)
		} else {
			return "", fmt.Errorf("kindlegen did not create MOBI file %s: %w", mobiPath, err)
		}
	}

	return mobiPath, nil
}

// ConvertEPUBToMOBI converts an EPUB file to MOBI format using kindlegen.
// This is useful for testing or when MOBI format is preferred.
func (kc *KindlegenConverter) ConvertEPUBToMOBI(inputPath, outputPath string) error {
	// Change output extension to .mobi
	mobiPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".mobi"
	return kc.ConvertEPUBToAZW3(inputPath, mobiPath)
}

// SetVerbose enables or disables verbose output from kindlegen.
func (kc *KindlegenConverter) SetVerbose(verbose bool) {
	kc.verbose = verbose
}

// GetKindlegenVersion returns the version of the kindlegen binary.
func (kc *KindlegenConverter) GetKindlegenVersion() (string, error) {
	cmd := exec.Command(kc.kindlegenPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get kindlegen version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
