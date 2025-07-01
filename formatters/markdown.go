package formatters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"concurrent-db/lockanalyzer"
)

// FormatMarkdown formate les données en Markdown et les écrit vers un Writer
func FormatMarkdown(data *lockanalyzer.ReportData, output io.Writer) error {
	var content strings.Builder

	// En-tête
	content.WriteString("# Rapport d'Analyse des Locks PostgreSQL\n\n")
	content.WriteString(fmt.Sprintf("**Généré le:** %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05")))

	// Résumé
	content.WriteString("## 📊 Résumé Exécutif\n\n")
	content.WriteString("| Métrique | Valeur |\n")
	content.WriteString("|----------|--------|\n")
	content.WriteString(fmt.Sprintf("| 🔒 Total locks actifs | %d |\n", data.Summary.TotalLocks))
	content.WriteString(fmt.Sprintf("| ⏳ Transactions bloquées | %d |\n", data.Summary.BlockedTxns))
	content.WriteString(fmt.Sprintf("| ⏰ Transactions longues | %d |\n", data.Summary.LongTxns))
	content.WriteString(fmt.Sprintf("| 💀 Deadlocks détectés | %d |\n", data.Summary.Deadlocks))
	content.WriteString(fmt.Sprintf("| ⚠️ Conflits d'objets | %d |\n", data.Summary.ObjectConflicts))
	content.WriteString(fmt.Sprintf("| 🚨 Problèmes critiques | %d |\n", data.Summary.CriticalIssues))
	content.WriteString(fmt.Sprintf("| ⚡ Avertissements | %d |\n", data.Summary.Warnings))
	content.WriteString(fmt.Sprintf("| 💡 Recommandations | %d |\n\n", data.Summary.Recommendations))

	// Locks actifs
	if len(data.Locks) > 0 {
		content.WriteString("## 🔒 Locks Actifs\n\n")
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
		content.WriteString("## ⏳ Transactions Bloquées\n\n")
		content.WriteString("| PID | Durée | Query |\n")
		content.WriteString("|-----|-------|-------|\n")
		for _, txn := range data.BlockedTxns {
			content.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Transactions longues
	if len(data.LongTxns) > 0 {
		content.WriteString("## ⏰ Transactions Longues\n\n")
		content.WriteString("| PID | Durée | Query |\n")
		content.WriteString("|-----|-------|-------|\n")
		for _, txn := range data.LongTxns {
			content.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Suggestions
	if len(data.Suggestions) > 0 {
		content.WriteString("## 💡 Suggestions d'Amélioration\n\n")
		for i, suggestion := range data.Suggestions {
			content.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, suggestion))
		}
	}

	_, err := output.Write([]byte(content.String()))
	return err
}

// WriteMarkdownFile écrit le rapport Markdown dans un fichier
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
