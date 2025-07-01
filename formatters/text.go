package formatters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"lock-analyser/i18n"
	"lock-analyser/lockanalyzer"
)

// TextFormatter formats data as text with multilingual support
type TextFormatter struct {
	translator *i18n.Translator
}

// NewTextFormatter creates a new text formatter for the specified language
func NewTextFormatter(lang string) *TextFormatter {
	return &TextFormatter{
		translator: i18n.NewTranslator(lang),
	}
}

// Format implements the LockReportFormatter interface
func (f *TextFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	t := f.translator
	var content strings.Builder

	// Header
	content.WriteString(strings.Repeat("=", 80) + "\n")
	content.WriteString(t.T("report_title") + "\n")
	content.WriteString(strings.Repeat("=", 80) + "\n")
	content.WriteString(fmt.Sprintf("%s: %s\n\n", t.T("generated_at"), data.Timestamp.Format("2006-01-02 15:04:05")))

	// Summary
	content.WriteString(t.T("summary_title") + "\n")
	content.WriteString(strings.Repeat("-", 40) + "\n")
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("total_locks"), data.Summary.TotalLocks))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("blocked_transactions"), data.Summary.BlockedTxns))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("long_transactions"), data.Summary.LongTxns))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("deadlocks_detected"), data.Summary.Deadlocks))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("object_conflicts"), data.Summary.ObjectConflicts))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("critical_issues"), data.Summary.CriticalIssues))
	content.WriteString(fmt.Sprintf("%s: %d\n", t.T("warnings"), data.Summary.Warnings))
	content.WriteString(fmt.Sprintf("%s: %d\n\n", t.T("recommendations"), data.Summary.Recommendations))

	// Active locks
	if len(data.Locks) > 0 {
		content.WriteString(t.T("active_locks") + "\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, lock := range data.Locks {
			content.WriteString(t.TWithData("lock_info_format", map[string]interface{}{
				"arg1": lock.PID,
				"arg2": lock.Mode,
				"arg3": lock.Granted,
				"arg4": lock.Type,
				"arg5": lock.Object,
			}) + "\n")
		}
		content.WriteString("\n")
	}

	// Blocked transactions
	if len(data.BlockedTxns) > 0 {
		content.WriteString(t.T("blocked_transactions_section") + "\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, txn := range data.BlockedTxns {
			content.WriteString(t.TWithData("blocked_transaction_format", map[string]interface{}{
				"arg1": txn.PID,
				"arg2": txn.Duration,
				"arg3": txn.Query,
			}) + "\n")
		}
		content.WriteString("\n")
	}

	// Long transactions
	if len(data.LongTxns) > 0 {
		content.WriteString(t.T("long_transactions_section") + "\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, txn := range data.LongTxns {
			content.WriteString(t.TWithData("long_transaction_format", map[string]interface{}{
				"arg1": txn.PID,
				"arg2": txn.Duration,
				"arg3": txn.Query,
			}) + "\n")
		}
		content.WriteString("\n")
	}

	// Suggestions
	if len(data.Suggestions) > 0 {
		content.WriteString(t.T("improvement_suggestions") + "\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for i, suggestion := range data.Suggestions {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
		content.WriteString("\n")
	}

	// Footer
	content.WriteString(t.T("report_footer") + "\n")

	_, err := output.Write([]byte(content.String()))
	return err
}

// GetFileExtension returns the file extension for this formatter
func (f *TextFormatter) GetFileExtension() string {
	return "txt"
}

// FormatText formats data as text and writes to a Writer (legacy version)
func FormatText(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter := NewTextFormatter("fr") // French as default for compatibility
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
