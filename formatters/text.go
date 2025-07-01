package formatters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"concurrent-db/lockanalyzer"
)

// FormatText formate les données en texte et les écrit vers un Writer
func FormatText(data *lockanalyzer.ReportData, output io.Writer) error {
	var content strings.Builder

	// En-tête
	content.WriteString(strings.Repeat("=", 80) + "\n")
	content.WriteString("RAPPORT D'ANALYSE DES LOCKS POSTGRESQL\n")
	content.WriteString(strings.Repeat("=", 80) + "\n")
	content.WriteString(fmt.Sprintf("Généré le: %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05")))

	// Résumé
	content.WriteString("RÉSUMÉ EXÉCUTIF\n")
	content.WriteString(strings.Repeat("-", 40) + "\n")
	content.WriteString(fmt.Sprintf("Total locks actifs: %d\n", data.Summary.TotalLocks))
	content.WriteString(fmt.Sprintf("Transactions bloquées: %d\n", data.Summary.BlockedTxns))
	content.WriteString(fmt.Sprintf("Transactions longues: %d\n", data.Summary.LongTxns))
	content.WriteString(fmt.Sprintf("Deadlocks détectés: %d\n", data.Summary.Deadlocks))
	content.WriteString(fmt.Sprintf("Conflits d'objets: %d\n", data.Summary.ObjectConflicts))
	content.WriteString(fmt.Sprintf("Problèmes critiques: %d\n", data.Summary.CriticalIssues))
	content.WriteString(fmt.Sprintf("Avertissements: %d\n", data.Summary.Warnings))
	content.WriteString(fmt.Sprintf("Recommandations: %d\n\n", data.Summary.Recommendations))

	// Locks actifs
	if len(data.Locks) > 0 {
		content.WriteString("LOCKS ACTIFS\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, lock := range data.Locks {
			content.WriteString(fmt.Sprintf("PID: %d, Mode: %s, Granted: %t, Type: %s",
				lock.PID, lock.Mode, lock.Granted, lock.Type))
			if lock.Object != "" {
				content.WriteString(fmt.Sprintf(", Object: %s", lock.Object))
			}
			if lock.Page != "" {
				content.WriteString(fmt.Sprintf(", Page: %s, Tuple: %s", lock.Page, lock.Tuple))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Transactions bloquées
	if len(data.BlockedTxns) > 0 {
		content.WriteString("TRANSACTIONS BLOQUÉES\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, txn := range data.BlockedTxns {
			content.WriteString(fmt.Sprintf("PID: %s, Durée: %s, Query: %s\n",
				txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Transactions longues
	if len(data.LongTxns) > 0 {
		content.WriteString("TRANSACTIONS LONGUES\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for _, txn := range data.LongTxns {
			content.WriteString(fmt.Sprintf("PID: %s, Durée: %s, Query: %s\n",
				txn.PID, txn.Duration, txn.Query))
		}
		content.WriteString("\n")
	}

	// Suggestions
	if len(data.Suggestions) > 0 {
		content.WriteString("SUGGESTIONS D'AMÉLIORATION\n")
		content.WriteString(strings.Repeat("-", 40) + "\n")
		for i, suggestion := range data.Suggestions {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
		content.WriteString("\n")
	}

	_, err := output.Write([]byte(content.String()))
	return err
}

// WriteTextFile écrit le rapport texte dans un fichier
func WriteTextFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.txt", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer file.Close()

	return FormatText(data, file)
}
