package formatters

import (
	"fmt"
	"io"
	"os"

	"concurrent-db/i18n"
	"concurrent-db/lockanalyzer"

	"github.com/uptrace/bun"
)

// LockReportFormatter définit l'interface pour formater et écrire des rapports de locks
type LockReportFormatter interface {
	Format(data *lockanalyzer.ReportData, output io.Writer) error
	GetFileExtension() string
}

// NewFormatter crée un nouveau formatter pour le format et la langue spécifiés
func NewFormatter(format, lang string) (LockReportFormatter, error) {
	// Valider la langue
	if !i18n.IsValidLanguage(lang) {
		lang = "fr" // Français par défaut
	}

	switch format {
	case "text", "txt":
		return NewTextFormatter(lang), nil
	case "markdown", "md":
		return NewMarkdownFormatter(lang), nil
	case "json":
		return NewJSONFormatter(lang), nil
	default:
		return nil, fmt.Errorf("format non supporté: %s", format)
	}
}

// GetAvailableFormats retourne la liste des formats disponibles
func GetAvailableFormats() []string {
	return []string{"text", "markdown", "json"}
}

// GetAvailableLanguages retourne la liste des langues disponibles
func GetAvailableLanguages() []string {
	return i18n.GetAvailableLanguages()
}

// GenerateAndWriteReport génère un rapport et l'écrit en utilisant le formatter spécifié
func GenerateAndWriteReport(db *bun.DB, formatter LockReportFormatter, filename string) error {
	// Générer les données du rapport
	reportData, err := lockanalyzer.GenerateLocksReport(db)
	if err != nil {
		return fmt.Errorf("erreur lors de la génération des données du rapport: %v", err)
	}

	// Créer le fichier de sortie
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier %s: %v", filename, err)
	}
	defer file.Close()

	// Formater et écrire le rapport
	if err := formatter.Format(reportData, file); err != nil {
		return fmt.Errorf("erreur lors du formatage du rapport: %v", err)
	}

	return nil
}

// GenerateAndDisplayReport génère un rapport et l'affiche sur stdout en utilisant le formatter spécifié
func GenerateAndDisplayReport(db *bun.DB, formatter LockReportFormatter) error {
	// Générer les données du rapport
	reportData, err := lockanalyzer.GenerateLocksReport(db)
	if err != nil {
		return fmt.Errorf("erreur lors de la génération des données du rapport: %v", err)
	}

	// Formater et afficher le rapport
	if err := formatter.Format(reportData, os.Stdout); err != nil {
		return fmt.Errorf("erreur lors du formatage du rapport: %v", err)
	}

	return nil
}
