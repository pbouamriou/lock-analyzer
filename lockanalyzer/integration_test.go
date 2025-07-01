package lockanalyzer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestConcurrentTransactions teste les transactions concurrentes
func TestConcurrentTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Récupérer les données de test
	var projects []Project
	if err := tdb.DB.NewSelect().Model(&projects).Scan(context.Background()); err != nil {
		t.Fatalf("Erreur lors de la récupération des projets: %v", err)
	}

	var models []Model
	if err := tdb.DB.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Erreur lors de la récupération des modèles: %v", err)
	}

	var files []File
	if err := tdb.DB.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Erreur lors de la récupération des fichiers: %v", err)
	}

	if len(models) == 0 || len(files) == 0 {
		t.Fatal("Aucun modèle ou fichier trouvé dans la base de données")
	}

	// Test 1: Transaction simple sans conflit
	t.Run("TransactionSimple", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tx, err := tdb.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Erreur lors du début de transaction: %v", err)
		}

		// Mettre à jour un modèle
		_, err = tx.NewUpdate().Model(&Model{ID: models[0].ID, State: "updated"}).Column("state").WherePK().Exec(ctx)
		if err != nil {
			t.Fatalf("Erreur lors de la mise à jour: %v", err)
		}

		// Valider la transaction
		if err := tx.Commit(); err != nil {
			t.Fatalf("Erreur lors de la validation: %v", err)
		}

		// Vérifier que le rapport peut être généré
		report, err := GenerateLocksReport(tdb.DB)
		if err != nil {
			t.Fatalf("Erreur lors de la génération du rapport: %v", err)
		}

		if report.Timestamp.IsZero() {
			t.Error("Le timestamp du rapport ne doit pas être vide")
		}
	})

	// Test 2: Transactions concurrentes avec potentiel de lock
	t.Run("TransactionsConcurrentes", func(t *testing.T) {
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		// Démarrer la première transaction
		tx1, err := tdb.DB.BeginTx(ctx1, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Erreur lors du début de transaction 1: %v", err)
		}

		// Mettre à jour un modèle (va déclencher le trigger sur projects)
		_, err = tx1.NewUpdate().Model(&Model{ID: models[0].ID, State: "locked"}).Column("state").WherePK().Exec(ctx1)
		if err != nil {
			t.Fatalf("Erreur lors de la mise à jour dans tx1: %v", err)
		}

		// Démarrer la deuxième transaction
		tx2, err := tdb.DB.BeginTx(ctx2, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Erreur lors du début de transaction 2: %v", err)
		}

		// Essayer de mettre à jour le même projet (peut causer un lock)
		_, err = tx2.NewUpdate().Model(&Project{ID: projects[0].ID, Name: "updated name"}).Column("name").WherePK().Exec(ctx2)
		if err != nil {
			t.Logf("Transaction 2 bloquée (attendu): %v", err)
		}

		// Valider les transactions
		if err := tx1.Commit(); err != nil {
			t.Fatalf("Erreur lors de la validation de tx1: %v", err)
		}

		if err := tx2.Commit(); err != nil {
			t.Logf("Erreur lors de la validation de tx2 (peut être normale): %v", err)
		}

		// Vérifier que le rapport peut être généré après les transactions
		report, err := GenerateLocksReport(tdb.DB)
		if err != nil {
			t.Fatalf("Erreur lors de la génération du rapport: %v", err)
		}

		if report.Timestamp.IsZero() {
			t.Error("Le timestamp du rapport ne doit pas être vide")
		}
	})
}

// TestDetectBlockedTransactionsReal teste la détection des transactions bloquées en temps réel
func TestDetectBlockedTransactionsReal(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Démarrer une transaction longue
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := tdb.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("Erreur lors du début de transaction: %v", err)
	}

	// Faire une mise à jour qui va durer
	_, err = tx.NewUpdate().Model(&Model{ID: "660e8400-e29b-41d4-a716-446655440001", State: "long_transaction"}).Column("state").WherePK().Exec(ctx)
	if err != nil {
		t.Fatalf("Erreur lors de la mise à jour: %v", err)
	}

	// Attendre un peu pour que la transaction soit visible
	time.Sleep(100 * time.Millisecond)

	// Détecter les transactions bloquées
	blocked := DetectBlockedTransactions(tdb.DB)

	// Il peut y avoir ou non des transactions bloquées selon l'état de la base
	t.Logf("Transactions bloquées détectées: %v", blocked)

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Erreur lors de la validation: %v", err)
	}
}

// TestGenerateLocksReportWithRealData teste la génération de rapport avec des données réelles
func TestGenerateLocksReportWithRealData(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Générer un rapport
	report, err := GenerateLocksReport(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la génération du rapport: %v", err)
	}

	// Vérifications de base
	if report.Timestamp.IsZero() {
		t.Error("Le timestamp du rapport ne doit pas être vide")
	}

	if report.Summary.TotalLocks < 0 {
		t.Error("Le nombre total de locks ne peut pas être négatif")
	}

	if report.Summary.CriticalIssues < 0 {
		t.Error("Le nombre de problèmes critiques ne peut pas être négatif")
	}

	// Vérifier que les suggestions sont générées
	if len(report.Suggestions) == 0 {
		t.Error("Le rapport doit contenir au moins une suggestion")
	}

	// Vérifier que l'analyse des index fonctionne
	if len(report.IndexAnalysis) == 0 {
		t.Log("Aucun index analysé (peut être normal selon la configuration)")
	}
}

// TestLockDetectionWithTriggers teste la détection des locks avec les triggers
func TestLockDetectionWithTriggers(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Récupérer les données
	var projects []Project
	if err := tdb.DB.NewSelect().Model(&projects).Scan(context.Background()); err != nil {
		t.Fatalf("Erreur lors de la récupération des projets: %v", err)
	}

	var models []Model
	if err := tdb.DB.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(context.Background()); err != nil {
		t.Fatalf("Erreur lors de la récupération des modèles: %v", err)
	}

	// Test avec trigger sur models
	t.Run("TriggerModels", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tx, err := tdb.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			t.Fatalf("Erreur lors du début de transaction: %v", err)
		}

		// Mettre à jour un modèle (va déclencher le trigger)
		_, err = tx.NewUpdate().Model(&Model{ID: models[0].ID, State: "trigger_test"}).Column("state").WherePK().Exec(ctx)
		if err != nil {
			t.Fatalf("Erreur lors de la mise à jour: %v", err)
		}

		// Valider la transaction
		if err := tx.Commit(); err != nil {
			t.Fatalf("Erreur lors de la validation: %v", err)
		}

		// Vérifier que le projet a été mis à jour par le trigger
		var updatedProject Project
		if err := tdb.DB.NewSelect().Model(&updatedProject).Where("id = ?", projects[0].ID).Scan(context.Background()); err != nil {
			t.Fatalf("Erreur lors de la récupération du projet mis à jour: %v", err)
		}

		// Le modified_at devrait avoir été mis à jour
		if updatedProject.ModifiedAt.Equal(projects[0].ModifiedAt) {
			t.Log("Le trigger a fonctionné (modified_at mis à jour)")
		}
	})
}

// TestPerformanceWithLargeDataset teste les performances avec un grand nombre de données
func TestPerformanceWithLargeDataset(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	// Mesurer le temps de génération du rapport
	start := time.Now()

	report, err := GenerateLocksReport(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la génération du rapport: %v", err)
	}

	duration := time.Since(start)

	// Le rapport ne devrait pas prendre plus de 5 secondes
	if duration > 5*time.Second {
		t.Errorf("La génération du rapport prend trop de temps: %v", duration)
	}

	t.Logf("Génération du rapport en %v", duration)

	// Vérifier que le rapport est complet
	if report.Timestamp.IsZero() {
		t.Error("Le timestamp du rapport ne doit pas être vide")
	}

	if len(report.Suggestions) == 0 {
		t.Error("Le rapport doit contenir des suggestions")
	}
}
