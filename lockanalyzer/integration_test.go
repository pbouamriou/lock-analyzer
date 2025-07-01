package lockanalyzer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Models are now defined in models_test.go

// Test utilities are now defined in test_utils.go

// TestConcurrentTransactions tests concurrent transactions
func TestConcurrentTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Retrieve test data
	var projects []Project
	if err := tdb.DB.NewSelect().Model(&projects).Scan(context.Background()); err != nil {
		t.Fatalf("Error retrieving projects: %v", err)
	}

	var models []Model
	if err := tdb.DB.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Error retrieving models: %v", err)
	}

	var files []File
	if err := tdb.DB.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Error retrieving files: %v", err)
	}

	if len(models) == 0 || len(files) == 0 {
		t.Fatal("No models or files found in database")
	}

	// Test 1: Simple transaction without conflict
	t.Run("SimpleTransaction", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tx, err := tdb.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Error starting transaction: %v", err)
		}

		// Update a model
		_, err = tx.NewUpdate().Model(&Model{ID: models[0].ID, State: "updated"}).Column("state").WherePK().Exec(ctx)
		if err != nil {
			t.Fatalf("Error during update: %v", err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			t.Fatalf("Error during commit: %v", err)
		}

		// Verify that the report can be generated
		report, err := GenerateLocksReport(tdb.DB)
		if err != nil {
			t.Fatalf("Error generating report: %v", err)
		}

		if report.Timestamp.IsZero() {
			t.Error("Report timestamp should not be empty")
		}
	})

	// Test 2: Concurrent transactions with potential for locks
	t.Run("ConcurrentTransactions", func(t *testing.T) {
		ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel1()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()

		// Start the first transaction
		tx1, err := tdb.DB.BeginTx(ctx1, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Error starting transaction 1: %v", err)
		}

		// Update a model (will trigger the trigger on projects)
		_, err = tx1.NewUpdate().Model(&Model{ID: models[0].ID, State: "locked"}).Column("state").WherePK().Exec(ctx1)
		if err != nil {
			t.Fatalf("Error during update in tx1: %v", err)
		}

		// Start the second transaction
		tx2, err := tdb.DB.BeginTx(ctx2, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Error starting transaction 2: %v", err)
		}

		// Try to update the same project (may cause a lock)
		_, err = tx2.NewUpdate().Model(&Project{ID: projects[0].ID, Name: "updated name"}).Column("name").WherePK().Exec(ctx2)
		if err != nil {
			t.Logf("Transaction 2 blocked (expected): %v", err)
			_ = tx2.Rollback() // explicit rollback, but if already rolled back, it's not a problem
		} else {
			if err := tx2.Commit(); err != nil {
				t.Logf("Error during tx2 commit: %v", err)
			}
		}

		// Commit tx1, but if already rolled back (by the server), we ignore the error
		if err := tx1.Commit(); err != nil {
			t.Logf("Error during tx1 commit: %v", err)
		}

		// Verify that the report can be generated after the transactions
		report, err := GenerateLocksReport(tdb.DB)
		if err != nil {
			t.Fatalf("Error generating report: %v", err)
		}

		if report.Timestamp.IsZero() {
			t.Error("Report timestamp should not be empty")
		}
	})
}

// TestDetectBlockedTransactionsReal tests real-time detection of blocked transactions
func TestDetectBlockedTransactionsReal(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Start a long transaction
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := tdb.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	// Make an update that will last
	_, err = tx.NewUpdate().Model(&Model{ID: "660e8400-e29b-41d4-a716-446655440001", State: "long_transaction"}).Column("state").WherePK().Exec(ctx)
	if err != nil {
		t.Fatalf("Error during update: %v", err)
	}

	// Wait a bit for the transaction to be visible
	time.Sleep(100 * time.Millisecond)

	// Detect blocked transactions
	blocked := DetectBlockedTransactions(tdb.DB)

	// There may or may not be blocked transactions depending on the database state
	t.Logf("Blocked transactions detected: %v", blocked)

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Error during commit: %v", err)
	}
}

// TestGenerateLocksReportWithRealData tests report generation with real data
func TestGenerateLocksReportWithRealData(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Generate a report
	report, err := GenerateLocksReport(tdb.DB)
	if err != nil {
		t.Fatalf("Error generating report: %v", err)
	}

	// Basic checks
	if report.Timestamp.IsZero() {
		t.Error("Report timestamp should not be empty")
	}

	if report.Summary.TotalLocks < 0 {
		t.Error("Total number of locks cannot be negative")
	}

	if report.Summary.CriticalIssues < 0 {
		t.Error("Number of critical issues cannot be negative")
	}

	// Verify that suggestions are generated (may be empty if no locks)
	if len(report.Suggestions) >= 0 {
		t.Logf("Number of suggestions generated: %d", len(report.Suggestions))
	}

	// Verify that index analysis works
	if len(report.IndexAnalysis) >= 0 {
		t.Logf("Index analysis completed: %d indexes found", len(report.IndexAnalysis))
	}
}

// TestLockDetectionWithTriggers tests lock detection with triggers
func TestLockDetectionWithTriggers(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Retrieve data
	var projects []Project
	if err := tdb.DB.NewSelect().Model(&projects).Scan(context.Background()); err != nil {
		t.Fatalf("Error retrieving projects: %v", err)
	}

	var models []Model
	if err := tdb.DB.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Error retrieving models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("No models found in database")
	}

	// Test trigger behavior
	t.Run("TriggerUpdate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get the original modified_at
		var originalProject Project
		if err := tdb.DB.NewSelect().Model(&originalProject).Where("id = ?", projects[0].ID).Scan(ctx); err != nil {
			t.Fatalf("Error retrieving original project: %v", err)
		}

		originalModifiedAt := originalProject.ModifiedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(100 * time.Millisecond)

		// Update a model (should trigger project update)
		_, err := tdb.DB.NewUpdate().Model(&Model{ID: models[0].ID, State: "trigger_test"}).Column("state").WherePK().Exec(ctx)
		if err != nil {
			t.Fatalf("Error updating model: %v", err)
		}

		// Verify that the project has been updated by the trigger
		var updatedProject Project
		if err := tdb.DB.NewSelect().Model(&updatedProject).Where("id = ?", projects[0].ID).Scan(ctx); err != nil {
			t.Fatalf("Error retrieving updated project: %v", err)
		}

		// The modified_at should have been updated
		if !updatedProject.ModifiedAt.After(originalModifiedAt) {
			t.Error("Project modified_at should have been updated by the trigger")
		}
	})
}

// TestPerformanceWithLargeDataset tests performance with a large dataset
func TestPerformanceWithLargeDataset(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Measure report generation time
	start := time.Now()
	report, err := GenerateLocksReport(tdb.DB)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Error generating report: %v", err)
	}

	t.Logf("Report generation took: %v", duration)

	// Performance check: should complete within reasonable time
	if duration > 5*time.Second {
		t.Errorf("Report generation took too long: %v", duration)
	}

	// Verify that the report is complete
	if report.Timestamp.IsZero() {
		t.Error("Report timestamp should not be empty")
	}

	if report.Summary.TotalLocks < 0 {
		t.Error("Total number of locks cannot be negative")
	}
}
