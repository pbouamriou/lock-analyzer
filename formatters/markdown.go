package formatters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pbouamriou/lock-analyzer/i18n"
	"github.com/pbouamriou/lock-analyzer/lockanalyzer"
)

// MarkdownFormatter formats data as Markdown with multilingual support
type MarkdownFormatter struct {
	translator *i18n.Translator
}

// NewMarkdownFormatter creates a new Markdown formatter for the specified language
func NewMarkdownFormatter(lang string) *MarkdownFormatter {
	return &MarkdownFormatter{
		translator: i18n.NewTranslator(lang),
	}
}

// Format implements the LockReportFormatter interface
func (f *MarkdownFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	t := f.translator
	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("# %s\n\n", t.T("report_title")))
	content.WriteString(fmt.Sprintf("**%s:** %s\n\n", t.T("generated_at"), data.Timestamp.Format("2006-01-02 15:04:05")))

	// Summary
	content.WriteString(fmt.Sprintf("## 📊 %s\n\n", t.T("summary_title")))
	content.WriteString(fmt.Sprintf("| %s | %s |\n", t.T("table_metric"), t.T("table_value")))
	content.WriteString("|--------|-------|\n")
	content.WriteString(fmt.Sprintf("| 🔒 %s | %d |\n", t.T("total_locks"), data.Summary.TotalLocks))
	content.WriteString(fmt.Sprintf("| ⏳ %s | %d |\n", t.T("blocked_transactions"), data.Summary.BlockedTxns))
	content.WriteString(fmt.Sprintf("| ⏰ %s | %d |\n", t.T("long_transactions"), data.Summary.LongTxns))
	content.WriteString(fmt.Sprintf("| 💀 %s | %d |\n", t.T("deadlocks_detected"), data.Summary.Deadlocks))
	content.WriteString(fmt.Sprintf("| ⚠️ %s | %d |\n", t.T("object_conflicts"), data.Summary.ObjectConflicts))
	content.WriteString(fmt.Sprintf("| 🚨 %s | %d |\n", t.T("critical_issues"), data.Summary.CriticalIssues))
	content.WriteString(fmt.Sprintf("| ⚡ %s | %d |\n", t.T("warnings"), data.Summary.Warnings))
	content.WriteString(fmt.Sprintf("| 💡 %s | %d |\n\n", t.T("recommendations"), data.Summary.Recommendations))

	// Active locks
	if len(data.Locks) > 0 {
		content.WriteString(fmt.Sprintf("## 🔒 %s\n\n", t.T("active_locks")))
		content.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n",
			t.T("table_pid"), t.T("table_mode"), t.T("table_granted"), t.T("table_type"), t.T("table_object"), t.T("table_page"), t.T("table_tuple")))
		content.WriteString("|-----|------|---------|------|--------|------|-------|\n")
		for _, lock := range data.Locks {
			content.WriteString(fmt.Sprintf("| %d | %s | %t | %s | %s | %s | %s |\n",
				lock.PID, lock.Mode, lock.Granted, lock.Type, lock.Object, lock.Page, lock.Tuple))
		}
		content.WriteString("\n")
	}

	// Blocked transactions
	if len(data.BlockedTxns) > 0 {
		content.WriteString(fmt.Sprintf("## ⏳ %s\n\n", t.T("blocked_transactions_section")))
		content.WriteString(fmt.Sprintf("| %s | %s | %s |\n", t.T("table_pid"), t.T("table_duration"), t.T("table_query")))
		content.WriteString("|-----|----------|-------|\n")
		for _, txn := range data.BlockedTxns {
			content.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Long transactions
	if len(data.LongTxns) > 0 {
		content.WriteString(fmt.Sprintf("## ⏰ %s\n\n", t.T("long_transactions_section")))
		content.WriteString(fmt.Sprintf("| %s | %s | %s |\n", t.T("table_pid"), t.T("table_duration"), t.T("table_query")))
		content.WriteString("|-----|----------|-------|\n")
		for _, txn := range data.LongTxns {
			content.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Suggestions
	if len(data.Suggestions) > 0 {
		content.WriteString(fmt.Sprintf("## 💡 %s\n\n", t.T("improvement_suggestions")))
		for i, suggestion := range data.Suggestions {
			content.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, suggestion))
		}
	}

	// Footer
	content.WriteString(fmt.Sprintf("---\n*%s*\n", t.T("report_footer")))

	_, err := output.Write([]byte(content.String()))
	return err
}

// GetFileExtension returns the file extension for this formatter
func (f *MarkdownFormatter) GetFileExtension() string {
	return "md"
}

// FormatMarkdown formats data as Markdown and writes to a Writer (legacy version)
func FormatMarkdown(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter := NewMarkdownFormatter("fr") // French as default for compatibility
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
