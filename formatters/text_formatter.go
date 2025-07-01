package formatters

import (
	"concurrent-db/lockanalyzer"
	"io"
)

// TextFormatter implémente LockReportFormatter pour le format texte
type TextFormatter struct{}

// NewTextFormatter crée une nouvelle instance de TextFormatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format implémente l'interface LockReportFormatter pour le format texte
func (f *TextFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	return FormatText(data, output)
}

// GetFileExtension retourne l'extension de fichier pour le format texte
func (f *TextFormatter) GetFileExtension() string {
	return ".txt"
}
