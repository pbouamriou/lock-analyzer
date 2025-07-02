package formatters

import (
	"fmt"
	"io"
	"os"

	"github.com/pbouamriou/lock-analyzer/i18n"
	"github.com/pbouamriou/lock-analyzer/lockanalyzer"

	"github.com/uptrace/bun"
)

// LockReportFormatter defines the interface for formatting and writing lock reports
type LockReportFormatter interface {
	Format(data *lockanalyzer.ReportData, output io.Writer) error
	GetFileExtension() string
}

// NewFormatter creates a new formatter for the specified format and language
func NewFormatter(format, lang string) (LockReportFormatter, error) {
	// Validate language
	if !i18n.IsValidLanguage(lang) {
		lang = "fr" // French as default
	}

	switch format {
	case "text", "txt":
		return NewTextFormatter(lang), nil
	case "markdown", "md":
		return NewMarkdownFormatter(lang), nil
	case "json":
		return NewJSONFormatter(lang), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetAvailableFormats returns the list of available formats
func GetAvailableFormats() []string {
	return []string{"text", "markdown", "json"}
}

// GetAvailableLanguages returns the list of available languages
func GetAvailableLanguages() []string {
	return i18n.GetAvailableLanguages()
}

// GenerateAndWriteReport generates a report and writes it using the specified formatter
func GenerateAndWriteReport(db *bun.DB, formatter LockReportFormatter, filename string) error {
	// Generate report data
	reportData, err := lockanalyzer.GenerateLocksReport(db)
	if err != nil {
		return fmt.Errorf("error generating report data: %v", err)
	}

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", filename, err)
	}
	defer file.Close()

	// Format and write the report
	if err := formatter.Format(reportData, file); err != nil {
		return fmt.Errorf("error formatting report: %v", err)
	}

	return nil
}

// GenerateAndDisplayReport generates a report and displays it on stdout using the specified formatter
func GenerateAndDisplayReport(db *bun.DB, formatter LockReportFormatter) error {
	// Generate report data
	reportData, err := lockanalyzer.GenerateLocksReport(db)
	if err != nil {
		return fmt.Errorf("error generating report data: %v", err)
	}

	// Format and display the report
	if err := formatter.Format(reportData, os.Stdout); err != nil {
		return fmt.Errorf("error formatting report: %v", err)
	}

	return nil
}
