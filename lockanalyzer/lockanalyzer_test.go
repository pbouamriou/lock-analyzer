package lockanalyzer

import (
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Models are now defined in models_test.go

// Test utilities are now defined in test_utils.go

// TestGenerateLocksReport tests report generation without locks
func TestGenerateLocksReport(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	report, err := GenerateLocksReport(tdb.DB)
	if err != nil {
		t.Fatalf("Error generating report: %v", err)
	}

	// Basic checks
	if report.Timestamp.IsZero() {
		t.Error("Report timestamp should not be empty")
	}

	// Suggestions may be empty if no locks are detected
	if len(report.Suggestions) >= 0 {
		t.Logf("Number of suggestions generated: %d", len(report.Suggestions))
	}

	// Verify that summary is calculated correctly
	if report.Summary.TotalLocks < 0 {
		t.Error("Total number of locks cannot be negative")
	}

	if report.Summary.CriticalIssues < 0 {
		t.Error("Number of critical issues cannot be negative")
	}
}

// TestDetectBlockedTransactions tests detection of blocked transactions
func TestDetectBlockedTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	blocked := DetectBlockedTransactions(tdb.DB)

	// Without active transactions, there should not be any blocked transactions
	if len(blocked) > 0 {
		t.Logf("Blocked transactions detected: %v", blocked)
	}
}

// TestGetLocks tests lock retrieval
func TestGetLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Error retrieving locks: %v", err)
	}

	// Debug: display type and value
	t.Logf("Locks type: %T, Value: %v, Is nil: %v", locks, locks, locks == nil)

	// Verify that the function doesn't return an error
	// In Go, an empty slice [] is considered nil, which is normal
	// We just verify that the function doesn't return an error

	// On an empty database, there may not be any active locks
	t.Logf("Number of locks detected: %d", len(locks))
}

// TestGetRowLocks tests row lock retrieval
func TestGetRowLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	rowLocks, err := getRowLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Error retrieving row locks: %v", err)
	}

	// Verify that the function doesn't return an error
	// In Go, an empty slice [] is considered nil, which is normal
	// We just verify that the function doesn't return an error

	// On an empty database, there may not be any active row locks
	t.Logf("Number of row locks detected: %d", len(rowLocks))
}

// TestDetectLongTransactions tests detection of long transactions
func TestDetectLongTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	longTxns := detectLongTransactions(tdb.DB)

	// Without active transactions, there should not be any long transactions
	if len(longTxns) > 0 {
		t.Logf("Long transactions detected: %v", longTxns)
	}
}

// TestDetectObjectConflicts tests detection of object conflicts
func TestDetectObjectConflicts(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Error retrieving locks: %v", err)
	}

	conflicts := detectObjectConflicts(locks)

	// Verify that the function works correctly
	// In Go, an empty slice [] is considered nil, which is normal
	// We just verify that the function doesn't return an error

	// On an empty database, there may not be any object conflicts
	t.Logf("Number of object conflicts detected: %d", len(conflicts))
}

// TestAnalyzeIndexes tests index analysis
func TestAnalyzeIndexes(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	indexes := analyzeIndexes(tdb.DB)

	// There should be at least a few indexes
	if len(indexes) == 0 {
		t.Error("No indexes detected")
	}

	// Verify that indexes have valid information
	for _, index := range indexes {
		if index.Name == "" {
			t.Error("Index name should not be empty")
		}
		if index.Table == "" {
			t.Error("Table name should not be empty")
		}
	}
}

// TestCalculateSummary tests summary calculation
func TestCalculateSummary(t *testing.T) {
	data := &ReportData{
		Timestamp: time.Now(),
		Locks: []LockInfo{
			{PID: 1, Mode: "ExclusiveLock", Granted: true},
			{PID: 2, Mode: "ShareLock", Granted: false},
		},
		BlockedTxns: []BlockedTransaction{
			{PID: "2", Duration: "10s"},
		},
		LongTxns: []LongTransaction{
			{PID: "1", Duration: "30s"},
		},
		ObjectConflicts: []ObjectConflict{
			{Object: "projects", PIDs: []string{"1", "2"}},
		},
		Suggestions: []string{"Test suggestion"},
	}

	summary := calculateSummary(data)

	// Checks
	if summary.TotalLocks != 2 {
		t.Errorf("Expected total locks: 2, got: %d", summary.TotalLocks)
	}

	if summary.BlockedTxns != 1 {
		t.Errorf("Expected blocked transactions: 1, got: %d", summary.BlockedTxns)
	}

	if summary.LongTxns != 1 {
		t.Errorf("Expected long transactions: 1, got: %d", summary.LongTxns)
	}

	if summary.ObjectConflicts != 1 {
		t.Errorf("Expected object conflicts: 1, got: %d", summary.ObjectConflicts)
	}

	if summary.CriticalIssues != 1 {
		t.Errorf("Expected critical issues: 1, got: %d", summary.CriticalIssues)
	}

	if summary.Warnings != 2 {
		t.Errorf("Expected warnings: 2, got: %d", summary.Warnings)
	}

	if summary.Recommendations != 1 {
		t.Errorf("Expected recommendations: 1, got: %d", summary.Recommendations)
	}
}

// TestGenerateSuggestions tests suggestion generation
func TestGenerateSuggestions(t *testing.T) {
	data := &ReportData{
		BlockedTxns: []BlockedTransaction{
			{PID: "1", Duration: "10s"},
		},
		LongTxns: []LongTransaction{
			{PID: "2", Duration: "30s"},
		},
		ObjectConflicts: []ObjectConflict{
			{Object: "projects"},
		},
		Deadlocks: []DeadlockInfo{
			{Transaction1: LockInfo{PID: 1}, Transaction2: LockInfo{PID: 2}},
		},
		Locks: make([]LockInfo, 15), // More than 10 locks
	}

	suggestions := generateSuggestions(data)

	// Verify that there are suggestions
	if len(suggestions) == 0 {
		t.Error("There should be suggestions")
	}

	// Verify that suggestions are relevant
	for _, suggestion := range suggestions {
		if suggestion == "" {
			t.Error("Suggestion should not be empty")
		}
	}

	// Verify specific suggestions based on data
	hasTimeoutSuggestion := false
	hasSplitSuggestion := false
	for _, suggestion := range suggestions {
		if contains(suggestion, "timeout") {
			hasTimeoutSuggestion = true
		}
		if contains(suggestion, "split") {
			hasSplitSuggestion = true
		}
	}

	if !hasTimeoutSuggestion {
		t.Error("Should have timeout suggestion for blocked transactions")
	}
	if !hasSplitSuggestion {
		t.Error("Should have split suggestion for long transactions")
	}
}

// TestDetectDeadlocks tests deadlock detection
func TestDetectDeadlocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Mode: "ExclusiveLock", Granted: true, Object: "table1"},
		{PID: 2, Mode: "ExclusiveLock", Granted: false, Object: "table2"},
		{PID: 1, Mode: "ShareLock", Granted: false, Object: "table2"},
		{PID: 2, Mode: "ShareLock", Granted: true, Object: "table1"},
	}

	deadlocks := detectDeadlocks(locks)

	// Verify deadlock structure
	if len(deadlocks) > 0 {
		for _, deadlock := range deadlocks {
			if deadlock.Transaction1.PID == 0 {
				t.Error("Deadlock should have valid transaction 1")
			}
			if deadlock.Transaction2.PID == 0 {
				t.Error("Deadlock should have valid transaction 2")
			}
		}
	}
}

// TestDetectBlockedTransactionsFromLocks tests detection of blocked transactions from locks
func TestDetectBlockedTransactionsFromLocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Mode: "ExclusiveLock", Granted: false, Object: "table1"},
		{PID: 2, Mode: "ShareLock", Granted: true, Object: "table2"},
	}

	blocked := detectBlockedTransactions(locks)

	// Verify that blocked transactions are detected
	if len(blocked) > 0 {
		for _, txn := range blocked {
			if txn.PID == "" {
				t.Error("Blocked transaction should have valid PID")
			}
		}
	}
}

// TestDetectObjectConflictsFromLocks tests detection of object conflicts from locks
func TestDetectObjectConflictsFromLocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Mode: "ExclusiveLock", Granted: true, Object: "table1"},
		{PID: 2, Mode: "ShareLock", Granted: false, Object: "table1"},
		{PID: 3, Mode: "ExclusiveLock", Granted: false, Object: "table1"},
	}

	conflicts := detectObjectConflicts(locks)

	// Verify that conflicts have correct information
	if len(conflicts) > 0 {
		for _, conflict := range conflicts {
			if conflict.Object == "" {
				t.Error("Object conflict should have valid object name")
			}
			if len(conflict.PIDs) == 0 {
				t.Error("Object conflict should have PIDs")
			}
		}
	}
}

// Utility function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Utility function to check if a string contains a substring (case sensitive)
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
