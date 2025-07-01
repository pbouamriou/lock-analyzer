package formatters

import (
	"fmt"
	"io"
	"lock-analyser/lockanalyzer"
	"log"

	"github.com/uptrace/bun"
)

// ExampleUsage shows how to use the new formatter architecture
func ExampleUsage(db *bun.DB) {
	// Create formatters
	textFormatter := NewTextFormatter("")
	jsonFormatter := NewJSONFormatter("")
	markdownFormatter := NewMarkdownFormatter("")

	// Example 1: Generate and display a Markdown report
	fmt.Println("=== Displaying Markdown report ===")
	if err := GenerateAndDisplayReport(db, markdownFormatter); err != nil {
		log.Printf("Error: %v", err)
	}

	// Example 2: Generate and write a JSON report to a file
	fmt.Println("\n=== Generating JSON report ===")
	if err := GenerateAndWriteReport(db, jsonFormatter, "example_report.json"); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("JSON report generated: example_report.json")
	}

	// Example 3: Generate and write a text report to a file
	fmt.Println("\n=== Generating text report ===")
	if err := GenerateAndWriteReport(db, textFormatter, "example_report.txt"); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Text report generated: example_report.txt")
	}

	// Example 4: Usage with automatic file extensions
	baseFilename := "my_report"
	textFile := baseFilename + textFormatter.GetFileExtension()
	jsonFile := baseFilename + jsonFormatter.GetFileExtension()
	markdownFile := baseFilename + markdownFormatter.GetFileExtension()

	fmt.Println("\n=== Generation with automatic extensions ===")
	GenerateAndWriteReport(db, textFormatter, textFile)
	GenerateAndWriteReport(db, jsonFormatter, jsonFile)
	GenerateAndWriteReport(db, markdownFormatter, markdownFile)

	fmt.Printf("Generated files: %s, %s, %s\n", textFile, jsonFile, markdownFile)
}

// ExampleWithCustomFormatter shows how to create a custom formatter
func ExampleWithCustomFormatter(db *bun.DB) {
	// Create a custom formatter that uses the text format
	customFormatter := &CustomTextFormatter{
		prefix: "CUSTOM_REPORT: ",
	}

	// Use the custom formatter
	if err := GenerateAndWriteReport(db, customFormatter, "custom_report.txt"); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Custom report generated: custom_report.txt")
	}
}

// CustomTextFormatter is an example of a custom formatter
type CustomTextFormatter struct {
	prefix string
}

// Format implements the LockReportFormatter interface
func (f *CustomTextFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	// Add a custom prefix to the report
	fmt.Fprintf(output, "%s\n", f.prefix)
	return FormatText(data, output)
}

// GetFileExtension returns the extension for this formatter
func (f *CustomTextFormatter) GetFileExtension() string {
	return ".custom.txt"
}
