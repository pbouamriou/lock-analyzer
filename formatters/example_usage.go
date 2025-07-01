package formatters

import (
	"concurrent-db/lockanalyzer"
	"fmt"
	"io"
	"log"

	"github.com/uptrace/bun"
)

// ExampleUsage montre comment utiliser la nouvelle architecture des formatters
func ExampleUsage(db *bun.DB) {
	// Créer les formatters
	textFormatter := NewTextFormatter("")
	jsonFormatter := NewJSONFormatter("")
	markdownFormatter := NewMarkdownFormatter("")

	// Exemple 1: Générer et afficher un rapport en Markdown
	fmt.Println("=== Affichage du rapport en Markdown ===")
	if err := GenerateAndDisplayReport(db, markdownFormatter); err != nil {
		log.Printf("Erreur: %v", err)
	}

	// Exemple 2: Générer et écrire un rapport JSON dans un fichier
	fmt.Println("\n=== Génération du rapport JSON ===")
	if err := GenerateAndWriteReport(db, jsonFormatter, "example_report.json"); err != nil {
		log.Printf("Erreur: %v", err)
	} else {
		fmt.Println("Rapport JSON généré: example_report.json")
	}

	// Exemple 3: Générer et écrire un rapport texte dans un fichier
	fmt.Println("\n=== Génération du rapport texte ===")
	if err := GenerateAndWriteReport(db, textFormatter, "example_report.txt"); err != nil {
		log.Printf("Erreur: %v", err)
	} else {
		fmt.Println("Rapport texte généré: example_report.txt")
	}

	// Exemple 4: Utilisation avec des extensions de fichiers automatiques
	baseFilename := "my_report"
	textFile := baseFilename + textFormatter.GetFileExtension()
	jsonFile := baseFilename + jsonFormatter.GetFileExtension()
	markdownFile := baseFilename + markdownFormatter.GetFileExtension()

	fmt.Println("\n=== Génération avec extensions automatiques ===")
	GenerateAndWriteReport(db, textFormatter, textFile)
	GenerateAndWriteReport(db, jsonFormatter, jsonFile)
	GenerateAndWriteReport(db, markdownFormatter, markdownFile)

	fmt.Printf("Fichiers générés: %s, %s, %s\n", textFile, jsonFile, markdownFile)
}

// ExampleWithCustomFormatter montre comment créer un formatter personnalisé
func ExampleWithCustomFormatter(db *bun.DB) {
	// Créer un formatter personnalisé qui utilise le format texte
	customFormatter := &CustomTextFormatter{
		prefix: "CUSTOM_REPORT: ",
	}

	// Utiliser le formatter personnalisé
	if err := GenerateAndWriteReport(db, customFormatter, "custom_report.txt"); err != nil {
		log.Printf("Erreur: %v", err)
	} else {
		fmt.Println("Rapport personnalisé généré: custom_report.txt")
	}
}

// CustomTextFormatter est un exemple de formatter personnalisé
type CustomTextFormatter struct {
	prefix string
}

// Format implémente l'interface LockReportFormatter
func (f *CustomTextFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	// Ajouter un préfixe personnalisé au rapport
	fmt.Fprintf(output, "%s\n", f.prefix)
	return FormatText(data, output)
}

// GetFileExtension retourne l'extension pour ce formatter
func (f *CustomTextFormatter) GetFileExtension() string {
	return ".custom.txt"
}
