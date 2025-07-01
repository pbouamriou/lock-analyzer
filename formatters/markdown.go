package formatters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"concurrent-db/i18n"
	"concurrent-db/lockanalyzer"
)

// MarkdownFormatter formate les données en Markdown avec support multilingue
type MarkdownFormatter struct {
	translator *i18n.Translator
}

// NewMarkdownFormatter crée un nouveau formatter Markdown pour la langue spécifiée
func NewMarkdownFormatter(lang string) *MarkdownFormatter {
	return &MarkdownFormatter{
		translator: i18n.NewTranslator(lang),
	}
}

// Format implémente l'interface LockReportFormatter
func (f *MarkdownFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	t := f.translator
	var content strings.Builder

	// En-tête
	content.WriteString(fmt.Sprintf("# %s\n\n", t.T("report_title")))
	content.WriteString(fmt.Sprintf("**%s:** %s\n\n", t.T("generated_at"), data.Timestamp.Format("2006-01-02 15:04:05")))

	// Résumé
	content.WriteString(fmt.Sprintf("## 📊 %s\n\n", t.T("summary_title")))
	content.WriteString("| Métrique | Valeur |\n")
	content.WriteString("|----------|--------|\n")
	content.WriteString(fmt.Sprintf("| 🔒 %s | %d |\n", t.T("total_locks"), data.Summary.TotalLocks))
	content.WriteString(fmt.Sprintf("| ⏳ %s | %d |\n", t.T("blocked_transactions"), data.Summary.BlockedTxns))
	content.WriteString(fmt.Sprintf("| ⏰ %s | %d |\n", t.T("long_transactions"), data.Summary.LongTxns))
	content.WriteString(fmt.Sprintf("| 💀 %s | %d |\n", t.T("deadlocks_detected"), data.Summary.Deadlocks))
	content.WriteString(fmt.Sprintf("| ⚠️ %s | %d |\n", t.T("object_conflicts"), data.Summary.ObjectConflicts))
	content.WriteString(fmt.Sprintf("| 🚨 %s | %d |\n", t.T("critical_issues"), data.Summary.CriticalIssues))
	content.WriteString(fmt.Sprintf("| ⚡ %s | %d |\n", t.T("warnings"), data.Summary.Warnings))
	content.WriteString(fmt.Sprintf("| 💡 %s | %d |\n\n", t.T("recommendations"), data.Summary.Recommendations))

	// Locks actifs
	if len(data.Locks) > 0 {
		content.WriteString(fmt.Sprintf("## 🔒 %s\n\n", t.T("active_locks")))
		content.WriteString("| PID | Mode | Granted | Type | Object | Page | Tuple |\n")
		content.WriteString("|-----|------|---------|------|--------|------|-------|\n")
		for _, lock := range data.Locks {
			content.WriteString(fmt.Sprintf("| %d | %s | %t | %s | %s | %s | %s |\n",
				lock.PID, lock.Mode, lock.Granted, lock.Type, lock.Object, lock.Page, lock.Tuple))
		}
		content.WriteString("\n")
	}

	// Transactions bloquées
	if len(data.BlockedTxns) > 0 {
		content.WriteString(fmt.Sprintf("## ⏳ %s\n\n", t.T("blocked_transactions_section")))
		content.WriteString("| PID | Durée | Query |\n")
		content.WriteString("|-----|-------|-------|\n")
		for _, txn := range data.BlockedTxns {
			content.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Transactions longues
	if len(data.LongTxns) > 0 {
		content.WriteString(fmt.Sprintf("## ⏰ %s\n\n", t.T("long_transactions_section")))
		content.WriteString("| PID | Durée | Query |\n")
		content.WriteString("|-----|-------|-------|\n")
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

	// Pied de page
	content.WriteString(fmt.Sprintf("---\n*%s*\n", t.T("report_footer")))

	_, err := output.Write([]byte(content.String()))
	return err
}

// GetFileExtension retourne l'extension de fichier pour ce formatter
func (f *MarkdownFormatter) GetFileExtension() string {
	return "md"
}

// FormatMarkdown formate les données en Markdown et les écrit vers un Writer (version legacy)
func FormatMarkdown(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter := NewMarkdownFormatter("fr") // Français par défaut pour la compatibilité
	return formatter.Format(data, output)
}

// WriteMarkdownFile écrit le rapport Markdown dans un fichier (version legacy)
func WriteMarkdownFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.md", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer file.Close()

	return FormatMarkdown(data, file)
}
