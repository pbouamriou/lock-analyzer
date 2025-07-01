package formatters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"concurrent-db/i18n"
	"concurrent-db/lockanalyzer"
)

// JSONFormatter formats data as JSON with multilingual support
type JSONFormatter struct {
	translator *i18n.Translator
}

// NewJSONFormatter creates a new JSON formatter for the specified language
func NewJSONFormatter(lang string) *JSONFormatter {
	return &JSONFormatter{
		translator: i18n.NewTranslator(lang),
	}
}

// Format implements the LockReportFormatter interface
func (f *JSONFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	// For JSON, we add translations in the metadata
	jsonData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"language":                           f.translator.T("report_title"),
			"generated_at_label":                 f.translator.T("generated_at"),
			"summary_title":                      f.translator.T("summary_title"),
			"total_locks_label":                  f.translator.T("total_locks"),
			"blocked_transactions_label":         f.translator.T("blocked_transactions"),
			"long_transactions_label":            f.translator.T("long_transactions"),
			"deadlocks_detected_label":           f.translator.T("deadlocks_detected"),
			"object_conflicts_label":             f.translator.T("object_conflicts"),
			"critical_issues_label":              f.translator.T("critical_issues"),
			"warnings_label":                     f.translator.T("warnings"),
			"recommendations_label":              f.translator.T("recommendations"),
			"active_locks_label":                 f.translator.T("active_locks"),
			"blocked_transactions_section_label": f.translator.T("blocked_transactions_section"),
			"long_transactions_section_label":    f.translator.T("long_transactions_section"),
			"improvement_suggestions_label":      f.translator.T("improvement_suggestions"),
			"report_footer":                      f.translator.T("report_footer"),
		},
		"data": data,
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error during JSON serialization: %v", err)
	}

	_, err = output.Write(jsonBytes)
	return err
}

// GetFileExtension returns the file extension for this formatter
func (f *JSONFormatter) GetFileExtension() string {
	return "json"
}

// FormatJSON formats data as JSON and writes to a Writer (legacy version)
func FormatJSON(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter := NewJSONFormatter("fr") // French as default for compatibility
	return formatter.Format(data, output)
}

// WriteJSONFile writes the JSON report to a file (legacy version)
func WriteJSONFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.json", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return FormatJSON(data, file)
}
