package formatters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"concurrent-db/lockanalyzer"
)

// FormatJSON formate les données en JSON et les écrit vers un Writer
func FormatJSON(data *lockanalyzer.ReportData, output io.Writer) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON: %v", err)
	}

	_, err = output.Write(jsonData)
	return err
}

// WriteJSONFile écrit le rapport JSON dans un fichier
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
