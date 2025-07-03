package formatters

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pbouamriou/lock-analyzer/lockanalyzer"
)

// MarkdownFormatter formats data as Markdown using Go templates
type MarkdownFormatter struct {
	formatter *TemplateFormatter
}

// NewMarkdownFormatter creates a new Markdown formatter for the specified language
func NewMarkdownFormatter(lang string) (*MarkdownFormatter, error) {
	formatter, err := NewTemplateFormatter(lang, "markdown")
	if err != nil {
		return nil, err
	}

	return &MarkdownFormatter{
		formatter: formatter,
	}, nil
}

// Format implements the LockReportFormatter interface
func (f *MarkdownFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	return f.formatter.Format(data, output)
}

// GetFileExtension returns the file extension for this formatter
func (f *MarkdownFormatter) GetFileExtension() string {
	return "md"
}

// FormatMarkdown formats data as Markdown and writes to a Writer (legacy version)
func FormatMarkdown(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter, err := NewMarkdownFormatter("fr") // French as default for compatibility
	if err != nil {
		return err
	}
	return formatter.Format(data, output)
}

// WriteMarkdownFile writes the Markdown report to a file (legacy version)
func WriteMarkdownFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.md", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return FormatMarkdown(data, file)
}
