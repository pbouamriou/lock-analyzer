package formatters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"concurrent-db/i18n"
	"concurrent-db/lockanalyzer"
)

// JSONFormatter formate les données en JSON avec support multilingue
type JSONFormatter struct {
	translator *i18n.Translator
}

// NewJSONFormatter crée un nouveau formatter JSON pour la langue spécifiée
func NewJSONFormatter(lang string) *JSONFormatter {
	return &JSONFormatter{
		translator: i18n.NewTranslator(lang),
	}
}

// Format implémente l'interface LockReportFormatter
func (f *JSONFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	// Pour JSON, nous ajoutons les traductions dans les métadonnées
	jsonData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"language":                           f.translator.T("report_title"),
			"generated_at_label":                 f.translator.T("generated_at"),
			"summary_title":                      f.translator.T("summary_title"),
			"total_locks_label":                  f.translator.T("total_locks"),
			"blocked_transactions_label":         f.translator.T("blocked_transactions"),
			"long_transactions_label":            f.translator.T("long_transactions"),
			"deadlocks_detected_label":           f.translator.T("deadlocks_detected"),
			"object_conflicts_label":             f.translator.T("object_conflicts"),
			"critical_issues_label":              f.translator.T("critical_issues"),
			"warnings_label":                     f.translator.T("warnings"),
			"recommendations_label":              f.translator.T("recommendations"),
			"active_locks_label":                 f.translator.T("active_locks"),
			"blocked_transactions_section_label": f.translator.T("blocked_transactions_section"),
			"long_transactions_section_label":    f.translator.T("long_transactions_section"),
			"improvement_suggestions_label":      f.translator.T("improvement_suggestions"),
			"report_footer":                      f.translator.T("report_footer"),
		},
		"data": data,
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON: %v", err)
	}

	_, err = output.Write(jsonBytes)
	return err
}

// GetFileExtension retourne l'extension de fichier pour ce formatter
func (f *JSONFormatter) GetFileExtension() string {
	return "json"
}

// FormatJSON formate les données en JSON et les écrit vers un Writer (version legacy)
func FormatJSON(data *lockanalyzer.ReportData, output io.Writer) error {
	formatter := NewJSONFormatter("fr") // Français par défaut pour la compatibilité
	return formatter.Format(data, output)
}

// WriteJSONFile écrit le rapport JSON dans un fichier (version legacy)
func WriteJSONFile(data *lockanalyzer.ReportData, filename string) error {
	if filename == "" {
		filename = fmt.Sprintf("lock_analysis_%s.json", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer file.Close()

	return FormatJSON(data, file)
}
