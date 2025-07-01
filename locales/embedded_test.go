package locales

import (
	"testing"
)

func TestGetLocalesFS(t *testing.T) {
	fs := GetLocalesFS()
	if fs == nil {
		t.Fatal("GetLocalesFS() returned nil")
	}
}

func TestListLocaleFiles(t *testing.T) {
	files, err := ListLocaleFiles()
	if err != nil {
		t.Fatalf("ListLocaleFiles() failed: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No locale files found")
	}

	// Check that we have the expected files
	expectedFiles := map[string]bool{
		"en.json": false,
		"fr.json": false,
	}

	for _, file := range files {
		if _, exists := expectedFiles[file]; exists {
			expectedFiles[file] = true
		}
	}

	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected locale file %s not found", file)
		}
	}
}

func TestGetLocaleFile(t *testing.T) {
	// Test reading English file
	content, err := GetLocaleFile("en.json")
	if err != nil {
		t.Fatalf("GetLocaleFile('en.json') failed: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("English locale file is empty")
	}

	// Test reading French file
	content, err = GetLocaleFile("fr.json")
	if err != nil {
		t.Fatalf("GetLocaleFile('fr.json') failed: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("French locale file is empty")
	}

	// Test reading non-existent file
	_, err = GetLocaleFile("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error when reading non-existent file")
	}
}
