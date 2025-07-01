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
		{"texte français", "text", "fr", false},
		{"markdown anglais", "markdown", "en", false},
		{"json espagnol", "json", "es", false},
		{"format invalide", "invalid", "fr", true},
		{"langue invalide", "text", "invalid", false}, // Devrait utiliser le français par défaut
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, tt.lang)
			if tt.expectError {
				if err == nil {
					t.Error("Attendait une erreur mais n'en a pas eu")
				}
				return
			}
			if err != nil {
				t.Errorf("Erreur inattendue: %v", err)
			}
			if formatter == nil {
				t.Error("Formatter ne devrait pas être nil")
			}
		})
	}
}

func TestGetAvailableFormats(t *testing.T) {
	formats := GetAvailableFormats()
	expected := []string{"text", "markdown", "json"}

	if len(formats) != len(expected) {
		t.Errorf("Nombre de formats incorrect: got %d, want %d", len(formats), len(expected))
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
			t.Errorf("Format %s manquant dans la liste", format)
		}
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	languages := GetAvailableLanguages()
	expected := []string{"fr", "en", "es", "de"}

	if len(languages) != len(expected) {
		t.Errorf("Nombre de langues incorrect: got %d, want %d", len(languages), len(expected))
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
			t.Errorf("Langue %s manquante dans la liste", lang)
		}
	}
}

func TestTextFormatterI18n(t *testing.T) {
	// Test avec différentes langues
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("langue_"+lang, func(t *testing.T) {
			formatter := NewTextFormatter(lang)
			if formatter == nil {
				t.Fatal("NewTextFormatter ne devrait pas retourner nil")
			}

			// Créer des données de test
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
				t.Errorf("Erreur lors du formatage: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Le contenu ne devrait pas être vide")
			}

			// Vérifier que le titre est traduit
			if lang == "fr" && !strings.Contains(content, "RAPPORT D'ANALYSE DES LOCKS POSTGRESQL") {
				t.Error("Le titre français devrait être présent")
			}
			if lang == "en" && !strings.Contains(content, "POSTGRESQL LOCK ANALYSIS REPORT") {
				t.Error("Le titre anglais devrait être présent")
			}
		})
	}
}

func TestMarkdownFormatterI18n(t *testing.T) {
	// Test avec différentes langues
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("langue_"+lang, func(t *testing.T) {
			formatter := NewMarkdownFormatter(lang)
			if formatter == nil {
				t.Fatal("NewMarkdownFormatter ne devrait pas retourner nil")
			}

			// Créer des données de test
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
				t.Errorf("Erreur lors du formatage: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Le contenu ne devrait pas être vide")
			}

			if lang == "fr" && !strings.Contains(strings.ToLower(content), "rapport d'analyse des locks postgresql") {
				t.Error("Le titre français devrait être présent (robuste à la casse et emoji)")
			}
			if lang == "en" && !strings.Contains(content, "POSTGRESQL LOCK ANALYSIS REPORT") {
				t.Error("Le titre anglais devrait être présent")
			}
		})
	}
}

func TestJSONFormatterI18n(t *testing.T) {
	// Test avec différentes langues
	languages := []string{"fr", "en"}

	for _, lang := range languages {
		t.Run("langue_"+lang, func(t *testing.T) {
			formatter := NewJSONFormatter(lang)
			if formatter == nil {
				t.Fatal("NewJSONFormatter ne devrait pas retourner nil")
			}

			// Créer des données de test
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
				t.Errorf("Erreur lors du formatage: %v", err)
			}

			content := buf.String()
			if content == "" {
				t.Error("Le contenu ne devrait pas être vide")
			}

			// Vérifier que les métadonnées sont présentes
			if !strings.Contains(content, "metadata") {
				t.Error("Les métadonnées devraient être présentes")
			}

			// Vérifier que les traductions sont incluses
			if lang == "fr" && !strings.Contains(content, "RAPPORT D'ANALYSE DES LOCKS POSTGRESQL") {
				t.Error("La traduction française devrait être présente")
			}
			if lang == "en" && !strings.Contains(content, "POSTGRESQL LOCK ANALYSIS REPORT") {
				t.Error("La traduction anglaise devrait être présente")
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
		{"texte", "text", "fr", "txt"},
		{"markdown", "markdown", "en", "md"},
		{"json", "json", "es", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, tt.lang)
			if err != nil {
				t.Fatalf("Erreur lors de la création du formatter: %v", err)
			}

			ext := formatter.GetFileExtension()
			if ext != tt.expected {
				t.Errorf("Extension incorrecte: got %s, want %s", ext, tt.expected)
			}
		})
	}
}
