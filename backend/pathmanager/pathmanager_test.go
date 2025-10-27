package pathmanager

import (
	"testing"
)

func TestPathManager_GenerateBookPath(t *testing.T) {
	config := DefaultPathConfig("/test/library")
	pm := NewPathManager(config)

	tests := []struct {
		author   string
		title    string
		format   string
		expected string
		hasError bool
	}{
		{
			author:   "John Doe",
			title:    "The Great Book",
			format:   "epub",
			expected: "/test/library/J/John Doe/The Great Book/The Great Book - John Doe.epub",
			hasError: false,
		},
		{
			author:   "José María",
			title:    "El Libro Español",
			format:   "epub",
			expected: "/test/library/J/José María/El Libro Español/El Libro Español - José María.epub",
			hasError: false,
		},
		{
			author:   "Author/With\\Invalid:Chars",
			title:    "Title*With?Invalid\"Chars",
			format:   "epub",
			expected: "/test/library/A/Author/With\\Invalid:Chars/Title*With?Invalid\"Chars/Title*With?Invalid\"Chars - Author/With\\Invalid:Chars.epub",
			hasError: false,
		},
		{
			author:   "",
			title:    "Some Title",
			format:   "epub",
			expected: "",
			hasError: true,
		},
		{
			author:   "Some Author",
			title:    "",
			format:   "epub",
			expected: "",
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := pm.GenerateBookPath(test.author, test.title, test.format)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for author='%s', title='%s', but got none", test.author, test.title)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for author='%s', title='%s': %v", test.author, test.title, err)
			continue
		}

		if result != test.expected {
			t.Errorf("Expected '%s', got '%s' for author='%s', title='%s'", test.expected, result, test.author, test.title)
		}
	}
}

func TestPathManager_CleanForFilesystem(t *testing.T) {
	config := DefaultPathConfig("/test")
	pm := NewPathManager(config)

	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Text", "Normal Text"},
		{"José María", "Jose Maria"},
		{"Café & Résumé", "Cafe & Resume"},
		{"Text/With\\Invalid:Chars", "TextWithInvalidChars"},
		{"  Multiple   Spaces  ", "Multiple Spaces"},
		{"", "Unknown"},
		{"   ", "Unknown"},
		{"Only/Invalid:Chars", "OnlyInvalidChars"},
	}

	for _, test := range tests {
		result := pm.CleanForFilesystem(test.input)
		if result != test.expected {
			t.Errorf("Expected '%s', got '%s' for input '%s'", test.expected, result, test.input)
		}
	}
}

func TestPathManager_GenerateDuplicatePath(t *testing.T) {
	config := DefaultPathConfig("/test/library")
	pm := NewPathManager(config)

	result, err := pm.GenerateDuplicatePath("John Doe", "The Great Book", "epub", 2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "/test/library/J/John Doe/The Great Book (2)/The Great Book (2) - John Doe.epub"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestPathManager_OptionalCleaning(t *testing.T) {
	// Test with cleaning enabled
	configWithCleaning := &PathConfig{
		BaseDirectory:    "/test/library",
		AuthorSubdir:     true,
		TitleSubdir:      true,
		FilenamePattern:  "{title} - {author}.{format}",
		CleanFilenames:   true, // Enable cleaning
		UseAuthorHashing: true,
	}
	pmWithCleaning := NewPathManager(configWithCleaning)

	result1, err := pmWithCleaning.GenerateBookPath("José María", "El Libro Español", "epub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected1 := "/test/library/J/Jose Maria/El Libro Espanol/El Libro Espanol - Jose Maria.epub"
	if result1 != expected1 {
		t.Errorf("With cleaning: Expected '%s', got '%s'", expected1, result1)
	}

	// Test with cleaning disabled (default)
	configNoCleaning := DefaultPathConfigNoCleaning("/test/library")
	pmNoCleaning := NewPathManager(configNoCleaning)

	result2, err := pmNoCleaning.GenerateBookPath("José María", "El Libro Español", "epub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected2 := "/test/library/J/José María/El Libro Español/El Libro Español - José María.epub"
	if result2 != expected2 {
		t.Errorf("Without cleaning: Expected '%s', got '%s'", expected2, result2)
	}
}

func TestPathManager_AuthorHashing(t *testing.T) {
	// Test with hashing enabled (default)
	configWithHashing := DefaultPathConfig("/test/library")
	pmWithHashing := NewPathManager(configWithHashing)

	tests := []struct {
		author   string
		title    string
		expected string
	}{
		{
			author:   "John Doe",
			title:    "The Great Book",
			expected: "/test/library/J/John Doe/The Great Book/The Great Book - John Doe.epub",
		},
		{
			author:   "Alice Smith",
			title:    "Another Book",
			expected: "/test/library/A/Alice Smith/Another Book/Another Book - Alice Smith.epub",
		},
		{
			author:   "Zoe Wilson",
			title:    "Z Book",
			expected: "/test/library/Z/Zoe Wilson/Z Book/Z Book - Zoe Wilson.epub",
		},
		{
			author:   "123 Author",
			title:    "Number Book",
			expected: "/test/library/0-9/123 Author/Number Book/Number Book - 123 Author.epub",
		},
		{
			author:   "José María",
			title:    "Spanish Book",
			expected: "/test/library/J/José María/Spanish Book/Spanish Book - José María.epub",
		},
	}

	for _, test := range tests {
		result, err := pmWithHashing.GenerateBookPath(test.author, test.title, "epub")
		if err != nil {
			t.Errorf("Unexpected error for author='%s': %v", test.author, err)
			continue
		}
		if result != test.expected {
			t.Errorf("Expected '%s', got '%s' for author='%s'", test.expected, result, test.author)
		}
	}
}

func TestPathManager_NoHashing(t *testing.T) {
	// Test with hashing disabled
	config := &PathConfig{
		BaseDirectory:    "/test/library",
		AuthorSubdir:     true,
		TitleSubdir:      true,
		FilenamePattern:  "{title} - {author}.{format}",
		CleanFilenames:   false,
		UseAuthorHashing: false, // Disable hashing
	}
	pm := NewPathManager(config)

	result, err := pm.GenerateBookPath("John Doe", "The Great Book", "epub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "/test/library/John Doe/The Great Book/The Great Book - John Doe.epub"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
