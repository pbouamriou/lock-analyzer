package formatters

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pbouamriou/lock-analyzer/lockanalyzer"
)

// TextFormatter formats data as text using Go templates
type TextFormatter struct {
	formatter *TemplateFormatter
}

// NewTextFormatter creates a new text formatter for the specified language
func NewTextFormatter(lang string) (*TextFormatter, error) {
	formatter, err := NewTemplateFormatter(lang, "text")
	if err != nil {
		return nil, err
	}

	return &TextFormatter{
		formatter: formatter,
	}, nil
}

// Format implements the LockReportFormatter interface
func (f *TextFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	return f.formatter.Format(data, output)
}

// GetFileExtension returns the file extension for this formatter
func (f *TextFormatter) GetFileExtension() string {
	return "txt"
}

// FormatText formats data as text and writes to a Writer (legacy version)
func FormatText(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter, err := NewTextFormatter("fr") // French as default for compatibility
	if err != nil {
		return err
	}
	return formatter.Format(data, output)
}

// WriteTextFile writes the text report to a file (legacy version)
func WriteTextFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.txt", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	return FormatText(data, file)
}
