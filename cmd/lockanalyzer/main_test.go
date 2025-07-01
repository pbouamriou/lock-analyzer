package main

import (
	"concurrent-db/locales"
	"os"
	"testing"
)

func TestMainWithEmbeddedLocales(t *testing.T) {
	// Test that the main function can be called without errors
	// This verifies that embedded locales are working correctly

	// Save original args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Test help flag
	os.Args = []string{"lockanalyzer", "-help"}

	// We can't easily test main() directly, but we can verify that
	// the embedded locales are accessible
	localesFS := locales.GetLocalesFS()
	if localesFS == nil {
		t.Fatal("GetLocalesFS() returned nil")
	}

	// Test that we can list locale files
	files, err := locales.ListLocaleFiles()
	if err != nil {
		t.Fatalf("ListLocaleFiles() failed: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No locale files found in embedded filesystem")
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
			t.Errorf("Expected locale file %s not found in embedded filesystem", file)
		}
	}

	// Test that we can read locale files
	for file := range expectedFiles {
		content, err := locales.GetLocaleFile(file)
		if err != nil {
			t.Errorf("GetLocaleFile('%s') failed: %v", file, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Locale file %s is empty", file)
		}
	}
}
