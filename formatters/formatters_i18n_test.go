package formatters

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"concurrent-db/lockanalyzer"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		lang        string
		expectError bool
	}{
		{"French text", "text", "fr", false},
		{"English markdown", "markdown", "en", false},
		{"Spanish json", "json", "es", false},
		{"Invalid format", "invalid", "fr", true},
		{"Invalid language", "text", "invalid", false}, // Should use French as default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, tt.lang)
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but didn't get one")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if formatter == nil {
				t.Error("Formatter should not be nil")
			}
		})
	}
}

func TestGetAvailableFormats(t *testing.T) {
	formats := GetAvailableFormats()
	expected := []string{"text", "markdown", "json"}

	if len(formats) != len(expected) {
		t.Errorf("Incorrect number of formats: got %d, want %d", len(formats), len(expected))
	}

	for _, format := range expected {
		found := false
		for _, available := range formats {
			if available == format {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Format %s missing from list", format)
		}
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	languages := GetAvailableLanguages()
	expected := []string{"fr", "en", "es", "de"}

	if len(languages) != len(expected) {
		t.Errorf("Incorrect number of languages: got %d, want %d", len(languages), len(expected))
	}

	for _, lang := range expected {
		found := false
		for _, available := range languages {
			if available == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Language %s missing from list", lang)
		}
	}
}

func TestTextFormatterI18n(t *testing.T) {
	// Test with different languages
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("language_"+lang, func(t *testing.T) {
			formatter := NewTextFormatter(lang)
			if formatter == nil {
				t.Fatal("NewTextFormatter should not return nil")
			}

			// Create test data
			data := &lockanalyzer.ReportData{
				Timestamp: time.Now(),
				Summary: lockanalyzer.ReportSummary{
					TotalLocks:      5,
					BlockedTxns:     2,
					LongTxns:        1,
					Deadlocks:       0,
					ObjectConflicts: 1,
					CriticalIssues:  0,
					Warnings:        3,
					Recommendations: 2,
				},
				Locks: []lockanalyzer.LockInfo{
					{PID: 123, Mode: "AccessShareLock", Granted: true, Type: "relation", Object: "test_table"},
				},
				Suggestions: []string{"Test suggestion"},
			}

			var buf bytes.Buffer
			err := formatter.Format(data, &buf)
			if err != nil {
				t.Errorf("Error during formatting: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Content should not be empty")
			}

			// Check that title is translated
			if lang == "fr" && !strings.Contains(content, "RAPPORT D'ANALYSE DES LOCKS POSTGRESQL") {
				t.Error("French title should be present")
			}
			if lang == "en" && !strings.Contains(content, "POSTGRESQL LOCK ANALYSIS REPORT") {
				t.Error("English title should be present")
			}
		})
	}
}

func TestMarkdownFormatterI18n(t *testing.T) {
	// Test with different languages
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("language_"+lang, func(t *testing.T) {
			formatter := NewMarkdownFormatter(lang)
			if formatter == nil {
				t.Fatal("NewMarkdownFormatter should not return nil")
			}

			// Create test data
			data := &lockanalyzer.ReportData{
				Timestamp: time.Now(),
				Summary: lockanalyzer.ReportSummary{
					TotalLocks:      5,
					BlockedTxns:     2,
					LongTxns:        1,
					Deadlocks:       0,
					ObjectConflicts: 1,
					CriticalIssues:  0,
					Warnings:        3,
					Recommendations: 2,
				},
				Locks: []lockanalyzer.LockInfo{
					{PID: 123, Mode: "AccessShareLock", Granted: true, Type: "relation", Object: "test_table"},
				},
				Suggestions: []string{"Test suggestion"},
			}

			var buf bytes.Buffer
			err := formatter.Format(data, &buf)
			if err != nil {
				t.Errorf("Error during formatting: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Content should not be empty")
			}

			if lang == "fr" && !strings.Contains(strings.ToLower(content), "rapport d'analyse des locks postgresql") {
				t.Error("French title should be present (case and emoji robust)")
			}
			if lang == "en" && !strings.Contains(content, "POSTGRESQL LOCK ANALYSIS REPORT") {
				t.Error("English title should be present")
			}
		})
	}
}

func TestJSONFormatterI18n(t *testing.T) {
	// Test with different languages
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("language_"+lang, func(t *testing.T) {
			formatter := NewJSONFormatter(lang)
			if formatter == nil {
				t.Fatal("NewJSONFormatter should not return nil")
			}

			// Create test data
			data := &lockanalyzer.ReportData{
				Timestamp: time.Now(),
				Summary: lockanalyzer.ReportSummary{
					TotalLocks:      5,
					BlockedTxns:     2,
					LongTxns:        1,
					Deadlocks:       0,
					ObjectConflicts: 1,
					CriticalIssues:  0,
					Warnings:        3,
					Recommendations: 2,
				},
				Locks: []lockanalyzer.LockInfo{
					{PID: 123, Mode: "AccessShareLock", Granted: true, Type: "relation", Object: "test_table"},
				},
				Suggestions: []string{"Test suggestion"},
			}

			var buf bytes.Buffer
			err := formatter.Format(data, &buf)
			if err != nil {
				t.Errorf("Error during formatting: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Content should not be empty")
			}

			// Check that metadata is present
			if !strings.Contains(content, "metadata") {
				t.Error("JSON should contain metadata section")
			}

			// Check that translations are included
			if !strings.Contains(content, "language") {
				t.Error("JSON should include language information")
			}
		})
	}
}

func TestFormatterFileExtensions(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		lang     string
		expected string
	}{
		{"Text French", "text", "fr", "txt"},
		{"Markdown English", "markdown", "en", "md"},
		{"JSON Spanish", "json", "es", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, tt.lang)
			if err != nil {
				t.Fatalf("Error creating formatter: %v", err)
			}

			ext := formatter.GetFileExtension()
			if ext != tt.expected {
				t.Errorf("Expected extension %s, got %s", tt.expected, ext)
			}
		})
	}
}
