package formatters

import (
	"concurrent-db/lockanalyzer"
	"io"
)

// JSONFormatter implémente LockReportFormatter pour le format JSON
type JSONFormatter struct{}

// NewJSONFormatter crée une nouvelle instance de JSONFormatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format implémente l'interface LockReportFormatter pour le format JSON
func (f *JSONFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	return FormatJSON(data, output)
}

// GetFileExtension retourne l'extension de fichier pour le format JSON
func (f *JSONFormatter) GetFileExtension() string {
	return ".json"
}
