package formatters

import (
	"concurrent-db/lockanalyzer"
	"io"
)

// MarkdownFormatter implémente LockReportFormatter pour le format Markdown
type MarkdownFormatter struct{}

// NewMarkdownFormatter crée une nouvelle instance de MarkdownFormatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format implémente l'interface LockReportFormatter pour le format Markdown
func (f *MarkdownFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	return FormatMarkdown(data, output)
}

// GetFileExtension retourne l'extension de fichier pour le format Markdown
func (f *MarkdownFormatter) GetFileExtension() string {
	return ".md"
}
