package formatters

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"concurrent-db/lockanalyzer"
)

// TestMarkdownFormatter teste le formatter Markdown
func TestMarkdownFormatter(t *testing.T) {
	formatter := NewMarkdownFormatter()

	// Vérifier l'extension
	if formatter.GetFileExtension() != ".md" {
		t.Errorf("Extension attendue: .md, obtenue: %s", formatter.GetFileExtension())
	}

	// Tester le formatage
	var buf bytes.Buffer
	data := createTestReportData()

	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Erreur lors du formatage Markdown: %v", err)
	}

	output := buf.String()

	// Vérifications de base
	if !strings.Contains(output, "# Rapport d'Analyse des Locks PostgreSQL") {
		t.Error("Le rapport Markdown doit contenir le titre principal")
	}

	if !strings.Contains(output, "## 📊 Résumé Exécutif") {
		t.Error("Le rapport Markdown doit contenir la section résumé")
	}

	if !strings.Contains(output, "## 🔒 Locks Actifs") {
		t.Error("Le rapport Markdown doit contenir la section locks actifs")
	}

	if !strings.Contains(output, "## 💡 Suggestions d'Amélioration") {
		t.Error("Le rapport Markdown doit contenir la section suggestions")
	}

	// Vérifier les métriques
	if !strings.Contains(output, "🔒 Total locks actifs | 2") {
		t.Error("Le rapport doit afficher le nombre total de locks")
	}

	if !strings.Contains(output, "⏳ Transactions bloquées | 1") {
		t.Error("Le rapport doit afficher le nombre de transactions bloquées")
	}
}

// TestJSONFormatter teste le formatter JSON
func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()

	// Vérifier l'extension
	if formatter.GetFileExtension() != ".json" {
		t.Errorf("Extension attendue: .json, obtenue: %s", formatter.GetFileExtension())
	}

	// Tester le formatage
	var buf bytes.Buffer
	data := createTestReportData()

	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Erreur lors du formatage JSON: %v", err)
	}

	output := buf.String()

	// Vérifier que c'est du JSON valide
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsedData); err != nil {
		t.Fatalf("Le formatage JSON n'est pas valide: %v", err)
	}

	// Vérifier les champs requis
	requiredFields := []string{"Timestamp", "Locks", "BlockedTxns", "LongTxns", "Suggestions", "Summary"}
	for _, field := range requiredFields {
		if _, exists := parsedData[field]; !exists {
			t.Errorf("Le JSON doit contenir le champ: %s", field)
		}
	}

	// Vérifier le résumé
	if summary, exists := parsedData["Summary"].(map[string]interface{}); exists {
		if totalLocks, ok := summary["TotalLocks"].(float64); ok {
			if totalLocks != 2 {
				t.Errorf("TotalLocks attendu: 2, obtenu: %f", totalLocks)
			}
		} else {
			t.Error("TotalLocks doit être un nombre")
		}
	} else {
		t.Error("Le résumé doit être un objet")
	}
}

// TestTextFormatter teste le formatter texte
func TestTextFormatter(t *testing.T) {
	formatter := NewTextFormatter()

	// Vérifier l'extension
	if formatter.GetFileExtension() != ".txt" {
		t.Errorf("Extension attendue: .txt, obtenue: %s", formatter.GetFileExtension())
	}

	// Tester le formatage
	var buf bytes.Buffer
	data := createTestReportData()

	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Erreur lors du formatage texte: %v", err)
	}

	output := buf.String()
	t.Logf("Contenu généré par TextFormatter :\n%s", output)

	// Adapter les assertions au contenu réel généré (majuscules et accents)
	if !strings.Contains(output, "RAPPORT D'ANALYSE DES LOCKS POSTGRESQL") {
		t.Error("Le rapport texte doit contenir le titre principal exact")
	}
	if !strings.Contains(output, "RÉSUMÉ EXÉCUTIF") {
		t.Error("Le rapport texte doit contenir la section résumé exacte")
	}
	if !strings.Contains(output, "LOCKS ACTIFS") {
		t.Error("Le rapport texte doit contenir la section locks actifs exacte")
	}
	if !strings.Contains(output, "SUGGESTIONS D'AMÉLIORATION") {
		t.Error("Le rapport texte doit contenir la section suggestions exacte")
	}

	// Vérifier les métriques
	if !strings.Contains(output, "Total locks actifs: 2") {
		t.Error("Le rapport doit afficher le nombre total de locks")
	}

	if !strings.Contains(output, "Transactions bloquées: 1") {
		t.Error("Le rapport doit afficher le nombre de transactions bloquées")
	}
}

// TestGenerateAndWriteReport teste la fonction utilitaire
func TestGenerateAndWriteReport(t *testing.T) {
	// Créer un formatter de test
	formatter := &TestFormatter{}

	// Tester avec un fichier temporaire
	filename := "test_report.txt"

	// Créer des données de test
	data := createTestReportData()

	// Simuler la fonction GenerateLocksReport
	err := GenerateAndWriteReportWithData(data, formatter, filename)
	if err != nil {
		t.Fatalf("Erreur lors de la génération du rapport: %v", err)
	}

	// Vérifier que le fichier a été créé
	// Note: Dans un vrai test, on vérifierait le contenu du fichier
}

// TestGenerateAndDisplayReport teste la fonction d'affichage
func TestGenerateAndDisplayReport(t *testing.T) {
	// Créer un formatter de test
	formatter := &TestFormatter{}

	// Créer des données de test
	data := createTestReportData()

	// Tester l'affichage vers stdout
	var buf bytes.Buffer
	err := GenerateAndDisplayReportWithData(data, formatter, &buf)
	if err != nil {
		t.Fatalf("Erreur lors de l'affichage du rapport: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "TEST_FORMAT") {
		t.Error("L'affichage doit contenir le contenu formaté")
	}
}

// TestFormatter est un formatter de test pour les tests unitaires
type TestFormatter struct{}

func (f *TestFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	output.Write([]byte("TEST_FORMAT\n"))
	output.Write([]byte("Timestamp: " + data.Timestamp.Format(time.RFC3339) + "\n"))
	output.Write([]byte("Total Locks: " + string(rune(data.Summary.TotalLocks)) + "\n"))
	return nil
}

func (f *TestFormatter) GetFileExtension() string {
	return ".test"
}

// createTestReportData crée des données de test pour les formatters
func createTestReportData() *lockanalyzer.ReportData {
	return &lockanalyzer.ReportData{
		Timestamp: time.Now(),
		Locks: []lockanalyzer.LockInfo{
			{
				PID:     1,
				Mode:    "ExclusiveLock",
				Granted: true,
				Object:  "projects",
			},
			{
				PID:     2,
				Mode:    "ShareLock",
				Granted: false,
				Object:  "models",
			},
		},
		BlockedTxns: []lockanalyzer.BlockedTransaction{
			{
				PID:      "2",
				Duration: "10s",
				Query:    "SELECT * FROM models",
			},
		},
		LongTxns: []lockanalyzer.LongTransaction{
			{
				PID:      "1",
				Duration: "30s",
				Query:    "UPDATE projects SET name = 'test'",
			},
		},
		ObjectConflicts: []lockanalyzer.ObjectConflict{
			{
				Object:         "projects",
				PIDs:           []string{"1", "2"},
				Mode:           "multiple",
				Recommendation: "Review access patterns",
			},
		},
		IndexAnalysis: []lockanalyzer.IndexInfo{
			{
				Name:  "projects_pkey",
				Table: "projects",
				Size:  "16 kB",
			},
		},
		Suggestions: []string{
			"Considérer l'ajout de timeouts sur les transactions longues",
			"Diviser les transactions longues en transactions plus petites",
		},
		Summary: lockanalyzer.ReportSummary{
			TotalLocks:      2,
			BlockedTxns:     1,
			LongTxns:        1,
			Deadlocks:       0,
			ObjectConflicts: 1,
			CriticalIssues:  1,
			Warnings:        2,
			Recommendations: 2,
		},
	}
}

// TestCustomFormatter teste un formatter personnalisé
func TestCustomFormatter(t *testing.T) {
	customFormatter := &CustomTextFormatter{
		prefix: "CUSTOM_REPORT: ",
	}

	// Vérifier l'extension
	if customFormatter.GetFileExtension() != ".custom.txt" {
		t.Errorf("Extension attendue: .custom.txt, obtenue: %s", customFormatter.GetFileExtension())
	}

	// Tester le formatage
	var buf bytes.Buffer
	data := createTestReportData()

	err := customFormatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Erreur lors du formatage personnalisé: %v", err)
	}

	output := buf.String()
	t.Logf("Contenu généré par CustomFormatter :\n%s", output)

	// Vérifier le préfixe personnalisé
	if !strings.HasPrefix(output, "CUSTOM_REPORT: ") {
		t.Error("Le rapport personnalisé doit commencer par le préfixe")
	}

	// Adapter l'assertion au contenu réel généré
	if !strings.Contains(output, "Généré le:") {
		t.Error("Le rapport personnalisé doit contenir la date de génération")
	}
}

// TestFormatterInterface teste que tous les formatters implémentent l'interface
func TestFormatterInterface(t *testing.T) {
	formatters := []lockanalyzer.LockReportFormatter{
		NewMarkdownFormatter(),
		NewJSONFormatter(),
		NewTextFormatter(),
		&CustomTextFormatter{prefix: "TEST"},
	}

	for i, formatter := range formatters {
		// Vérifier que GetFileExtension fonctionne
		ext := formatter.GetFileExtension()
		if ext == "" {
			t.Errorf("Formatter %d: GetFileExtension ne doit pas retourner une chaîne vide", i)
		}

		// Vérifier que Format fonctionne
		var buf bytes.Buffer
		data := createTestReportData()
		err := formatter.Format(data, &buf)
		if err != nil {
			t.Errorf("Formatter %d: Format ne doit pas retourner d'erreur: %v", i, err)
		}

		// Vérifier que le formatage produit du contenu
		if buf.Len() == 0 {
			t.Errorf("Formatter %d: Format doit produire du contenu", i)
		}
	}
}

// TestEmptyData teste le formatage avec des données vides
func TestEmptyData(t *testing.T) {
	formatters := []lockanalyzer.LockReportFormatter{
		NewMarkdownFormatter(),
		NewJSONFormatter(),
		NewTextFormatter(),
	}

	emptyData := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
		Summary:   lockanalyzer.ReportSummary{},
	}

	for i, formatter := range formatters {
		var buf bytes.Buffer
		err := formatter.Format(emptyData, &buf)
		if err != nil {
			t.Errorf("Formatter %d: Format avec données vides ne doit pas retourner d'erreur: %v", i, err)
		}

		// Même avec des données vides, il doit y avoir du contenu
		if buf.Len() == 0 {
			t.Errorf("Formatter %d: Format avec données vides doit produire du contenu", i)
		}
	}
}

// TestLargeData teste le formatage avec beaucoup de données
func TestLargeData(t *testing.T) {
	formatter := NewMarkdownFormatter()

	// Créer beaucoup de locks
	locks := make([]lockanalyzer.LockInfo, 100)
	for i := range locks {
		locks[i] = lockanalyzer.LockInfo{
			PID:     i + 1,
			Mode:    "ExclusiveLock",
			Granted: i%2 == 0,
			Object:  "table_" + string(rune(i%10+'a')),
		}
	}

	data := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
		Locks:     locks,
		Summary: lockanalyzer.ReportSummary{
			TotalLocks: len(locks),
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Erreur lors du formatage de données volumineuses: %v", err)
	}

	output := buf.String()

	// Vérifier que le rapport contient les données
	if !strings.Contains(output, "Total locks actifs | 100") {
		t.Error("Le rapport doit afficher le bon nombre de locks")
	}
}

// Fonctions utilitaires pour les tests
func GenerateAndWriteReportWithData(data *lockanalyzer.ReportData, formatter lockanalyzer.LockReportFormatter, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return formatter.Format(data, file)
}

func GenerateAndDisplayReportWithData(data *lockanalyzer.ReportData, formatter lockanalyzer.LockReportFormatter, output io.Writer) error {
	return formatter.Format(data, output)
}
