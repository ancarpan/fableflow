package conversion

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetKindlegenPath returns the platform-specific path to the kindlegen executable.
// This function automatically detects the current platform and returns the appropriate
// kindlegen binary path, similar to FB2Converter's approach.
func GetKindlegenPath() (string, error) {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("unable to detect program path: %w", err)
	}

	// Get the directory containing the executable
	execDir, err := filepath.Abs(filepath.Dir(execPath))
	if err != nil {
		return "", fmt.Errorf("unable to calculate program path: %w", err)
	}

	// Determine the platform-specific kindlegen binary
	var kindlegenPath string
	switch runtime.GOOS {
	case "darwin":
		kindlegenPath = filepath.Join(execDir, "kindlegen", "darwin", "kindlegen")
	case "linux":
		kindlegenPath = filepath.Join(execDir, "kindlegen", "linux", "kindlegen")
	case "windows":
		kindlegenPath = filepath.Join(execDir, "kindlegen", "windows", "kindlegen.exe")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Check if the kindlegen binary exists
	if _, err := os.Stat(kindlegenPath); err != nil {
		return "", fmt.Errorf("unable to find kindlegen binary at %s: %w", kindlegenPath, err)
	}

	return kindlegenPath, nil
}

// GetKindlegenPathFromConfig allows specifying a custom kindlegen path.
// If the path is relative, it's resolved relative to the executable directory.
func GetKindlegenPathFromConfig(customPath string) (string, error) {
	if customPath == "" {
		return GetKindlegenPath()
	}

	// If path is absolute, use it as-is
	if filepath.IsAbs(customPath) {
		if _, err := os.Stat(customPath); err != nil {
			return "", fmt.Errorf("unable to find kindlegen at %s: %w", customPath, err)
		}
		return customPath, nil
	}

	// If path is relative, resolve it relative to the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("unable to detect program path: %w", err)
	}

	execDir, err := filepath.Abs(filepath.Dir(execPath))
	if err != nil {
		return "", fmt.Errorf("unable to calculate program path: %w", err)
	}

	kindlegenPath := filepath.Join(execDir, customPath)
	if _, err := os.Stat(kindlegenPath); err != nil {
		return "", fmt.Errorf("unable to find kindlegen at %s: %w", kindlegenPath, err)
	}

	return kindlegenPath, nil
}
