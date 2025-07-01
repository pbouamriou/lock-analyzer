package formatters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"concurrent-db/lockanalyzer"
)

// TestMarkdownFormatter tests the Markdown formatter
func TestMarkdownFormatter(t *testing.T) {
	formatter := NewMarkdownFormatter("fr")

	// Check extension
	if formatter.GetFileExtension() != "md" {
		t.Errorf("Expected extension: md, got: %s", formatter.GetFileExtension())
	}

	// Create test data
	data := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
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
		Locks: []lockanalyzer.LockInfo{
			{PID: 1, Mode: "ExclusiveLock", Granted: true, Type: "relation", Object: "projects"},
			{PID: 2, Mode: "ShareLock", Granted: false, Type: "relation", Object: "models"},
		},
		BlockedTxns: []lockanalyzer.BlockedTransaction{
			{PID: "2", Duration: "10s", Query: "SELECT * FROM models"},
		},
		LongTxns: []lockanalyzer.LongTransaction{
			{PID: "1", Duration: "30s", Query: "UPDATE projects SET name = 'test'"},
		},
		Suggestions: []string{
			"Consider adding timeouts on long transactions",
			"Split long transactions into smaller ones",
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Error during formatting: %v", err)
	}

	content := strings.ToLower(buf.String())
	if !strings.Contains(content, "rapport d'analyse des locks postgresql") {
		t.Error("Markdown report must contain main title (case and emoji robust)")
	}
	if !strings.Contains(content, "résumé exécutif") {
		t.Error("Markdown report must contain summary section (case and emoji robust)")
	}
	if !strings.Contains(content, "locks actifs") {
		t.Error("Markdown report must contain active locks section (case and emoji robust)")
	}
	if !strings.Contains(content, "suggestions d'amélioration") {
		t.Error("Markdown report must contain suggestions section (case and emoji robust)")
	}

	// Check data
	if !strings.Contains(content, "2") {
		t.Error("Report must display total number of locks")
	}
	if !strings.Contains(content, "1") {
		t.Error("Report must display number of blocked transactions")
	}
}

// TestJSONFormatter tests the JSON formatter
func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter("fr")

	// Check extension
	if formatter.GetFileExtension() != "json" {
		t.Errorf("Expected extension: json, got: %s", formatter.GetFileExtension())
	}

	// Create test data
	data := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
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
		Locks: []lockanalyzer.LockInfo{
			{PID: 1, Mode: "ExclusiveLock", Granted: true, Type: "relation", Object: "projects"},
			{PID: 2, Mode: "ShareLock", Granted: false, Type: "relation", Object: "models"},
		},
		BlockedTxns: []lockanalyzer.BlockedTransaction{
			{PID: "2", Duration: "10s", Query: "SELECT * FROM models"},
		},
		LongTxns: []lockanalyzer.LongTransaction{
			{PID: "1", Duration: "30s", Query: "UPDATE projects SET name = 'test'"},
		},
		Suggestions: []string{
			"Consider adding timeouts on long transactions",
			"Split long transactions into smaller ones",
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Error during formatting: %v", err)
	}

	content := buf.String()
	requiredFields := []string{"Timestamp", "Locks", "BlockedTxns", "LongTxns", "Suggestions", "Summary"}
	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("JSON must contain field: %s", field)
		}
	}

	// Check that it's valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		t.Fatalf("Generated content is not valid JSON: %v", err)
	}

	// Check structure
	if data, exists := jsonData["data"]; exists {
		if summary, ok := data.(map[string]interface{})["Summary"]; ok {
			if summaryObj, ok := summary.(map[string]interface{}); ok {
				if summaryObj["TotalLocks"] != float64(2) {
					t.Error("Summary must be an object")
				}
			}
		}
	}
}

// TestTextFormatter tests the text formatter
func TestTextFormatter(t *testing.T) {
	formatter := NewTextFormatter("fr")

	// Check extension
	if formatter.GetFileExtension() != "txt" {
		t.Errorf("Expected extension: txt, got: %s", formatter.GetFileExtension())
	}

	// Create test data
	data := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
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
		Locks: []lockanalyzer.LockInfo{
			{PID: 1, Mode: "ExclusiveLock", Granted: true, Type: "relation", Object: "projects"},
			{PID: 2, Mode: "ShareLock", Granted: false, Type: "relation", Object: "models"},
		},
		BlockedTxns: []lockanalyzer.BlockedTransaction{
			{PID: "2", Duration: "10s", Query: "SELECT * FROM models"},
		},
		LongTxns: []lockanalyzer.LongTransaction{
			{PID: "1", Duration: "30s", Query: "UPDATE projects SET name = 'test'"},
		},
		Suggestions: []string{
			"Consider adding timeouts on long transactions",
			"Split long transactions into smaller ones",
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Error during formatting: %v", err)
	}

	content := buf.String()
	t.Logf("Content generated by TextFormatter:\n%s", content)

	// Check content in French
	if !strings.Contains(content, "RAPPORT D'ANALYSE DES LOCKS POSTGRESQL") {
		t.Error("Text report must contain main title")
	}
	if !strings.Contains(content, "RÉSUMÉ EXÉCUTIF") {
		t.Error("Text report must contain summary section")
	}
	if !strings.Contains(content, "LOCKS ACTIFS") {
		t.Error("Text report must contain active locks section")
	}
	if !strings.Contains(content, "SUGGESTIONS D'AMÉLIORATION") {
		t.Error("Text report must contain suggestions section")
	}

	// Check data
	if !strings.Contains(content, "2") {
		t.Error("Report must display total number of locks")
	}
	if !strings.Contains(content, "1") {
		t.Error("Report must display number of blocked transactions")
	}
}

// TestGenerateAndWriteReport tests report generation and file writing
func TestGenerateAndWriteReport(t *testing.T) {
	// Create a test formatter
	formatter := &TestFormatter{}

	// Create test data
	data := createTestReportData()

	// Test file writing
	filename := "test_report.txt"
	err := GenerateAndWriteReportWithData(data, formatter, filename)
	if err != nil {
		t.Fatalf("Error writing report: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Report file was not created")
	}

	// Clean up
	os.Remove(filename)
}

// TestGenerateAndDisplayReport tests report generation and display
func TestGenerateAndDisplayReport(t *testing.T) {
	// Create a test formatter
	formatter := &TestFormatter{}

	// Create test data
	data := createTestReportData()

	// Test display to buffer
	var buf bytes.Buffer
	err := GenerateAndDisplayReportWithData(data, formatter, &buf)
	if err != nil {
		t.Fatalf("Error displaying report: %v", err)
	}

	// Check that content was written
	if buf.Len() == 0 {
		t.Error("No content was written to buffer")
	}
}

// TestFormatter is a test implementation of LockReportFormatter
type TestFormatter struct{}

func (f *TestFormatter) Format(data *lockanalyzer.ReportData, output io.Writer) error {
	_, err := output.Write([]byte("Test report content"))
	return err
}

func (f *TestFormatter) GetFileExtension() string {
	return "test"
}

// createTestReportData creates test data for formatters
func createTestReportData() *lockanalyzer.ReportData {
	return &lockanalyzer.ReportData{
		Timestamp: time.Now(),
		Summary: lockanalyzer.ReportSummary{
			TotalLocks:      5,
			BlockedTxns:     2,
			LongTxns:        1,
			Deadlocks:       0,
			ObjectConflicts: 1,
			CriticalIssues:  2,
			Warnings:        3,
			Recommendations: 4,
		},
		Locks: []lockanalyzer.LockInfo{
			{PID: 1, Mode: "ExclusiveLock", Granted: true, Type: "relation", Object: "projects"},
			{PID: 2, Mode: "ShareLock", Granted: false, Type: "relation", Object: "models"},
			{PID: 3, Mode: "RowShareLock", Granted: true, Type: "tuple", Object: "files"},
		},
		BlockedTxns: []lockanalyzer.BlockedTransaction{
			{PID: "2", Duration: "15s", Query: "SELECT * FROM models WHERE id = 1"},
			{PID: "4", Duration: "30s", Query: "UPDATE projects SET name = 'updated'"},
		},
		LongTxns: []lockanalyzer.LongTransaction{
			{PID: "1", Duration: "2m", Query: "UPDATE projects SET modified_at = NOW()"},
		},
		Suggestions: []string{
			"Consider adding timeouts on long transactions",
			"Split long transactions into smaller ones",
			"Review lock acquisition strategy",
			"Optimize queries to reduce execution time",
		},
	}
}

// TestFormatterInterface tests that all formatters implement the interface
func TestFormatterInterface(t *testing.T) {
	// Test all formatters
	formatters := []lockanalyzer.LockReportFormatter{
		NewTextFormatter("en"),
		NewMarkdownFormatter("en"),
		NewJSONFormatter("en"),
	}

	for _, formatter := range formatters {
		// Check that GetFileExtension works
		ext := formatter.GetFileExtension()
		if ext == "" {
			t.Error("GetFileExtension must return a non-empty string")
		}

		// Check that Format works
		data := createTestReportData()
		var buf bytes.Buffer
		err := formatter.Format(data, &buf)
		if err != nil {
			t.Errorf("Format method failed: %v", err)
		}

		// Check that formatting produces content
		if buf.Len() == 0 {
			t.Error("Format method must produce content")
		}
	}
}

// TestEmptyData tests formatting with empty data
func TestEmptyData(t *testing.T) {
	formatter := NewTextFormatter("en")

	// Create empty data
	data := &lockanalyzer.ReportData{
		Timestamp:   time.Now(),
		Summary:     lockanalyzer.ReportSummary{},
		Locks:       []lockanalyzer.LockInfo{},
		Suggestions: []string{},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Error formatting empty data: %v", err)
	}

	// Even with empty data, there must be content
	if buf.Len() == 0 {
		t.Error("Formatter must produce content even with empty data")
	}
}

// TestLargeData tests formatting with large amounts of data
func TestLargeData(t *testing.T) {
	formatter := NewMarkdownFormatter("en")

	// Create many locks
	var locks []lockanalyzer.LockInfo
	for i := 0; i < 100; i++ {
		locks = append(locks, lockanalyzer.LockInfo{
			PID:     i,
			Mode:    "ExclusiveLock",
			Granted: i%2 == 0,
			Type:    "relation",
			Object:  fmt.Sprintf("table_%d", i),
		})
	}

	data := &lockanalyzer.ReportData{
		Timestamp: time.Now(),
		Summary: lockanalyzer.ReportSummary{
			TotalLocks: len(locks),
		},
		Locks: locks,
		Suggestions: []string{
			"Consider reviewing transaction patterns",
			"Monitor lock wait times regularly",
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(data, &buf)
	if err != nil {
		t.Fatalf("Error formatting large data: %v", err)
	}

	content := buf.String()
	// Check that the number of locks is mentioned in a table row
	if !strings.Contains(content, "100") {
		t.Error("Report must mention the number of locks in a table row")
	}
}

// Helper functions for testing

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
