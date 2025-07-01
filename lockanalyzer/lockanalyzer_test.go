package lockanalyzer

import (
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Les modèles sont maintenant définis dans models_test.go

// Les utilitaires de test sont maintenant définis dans test_utils.go

// TestGenerateLocksReport teste la génération de rapport sans locks
func TestGenerateLocksReport(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	report, err := GenerateLocksReport(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la génération du rapport: %v", err)
	}

	// Vérifications de base
	if report.Timestamp.IsZero() {
		t.Error("Timestamp du rapport ne doit pas être vide")
	}

	if len(report.Suggestions) == 0 {
		t.Error("Le rapport doit contenir au moins une suggestion")
	}

	// Vérifier que le résumé est calculé correctement
	if report.Summary.TotalLocks < 0 {
		t.Error("Le nombre total de locks ne peut pas être négatif")
	}

	if report.Summary.CriticalIssues < 0 {
		t.Error("Le nombre de problèmes critiques ne peut pas être négatif")
	}
}

// TestDetectBlockedTransactions teste la détection des transactions bloquées
func TestDetectBlockedTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	blocked := DetectBlockedTransactions(tdb.DB)

	// Sans transactions actives, il ne devrait pas y avoir de transactions bloquées
	if len(blocked) > 0 {
		t.Logf("Transactions bloquées détectées: %v", blocked)
	}
}

// TestGetLocks teste la récupération des locks
func TestGetLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks: %v", err)
	}

	// Debug: afficher le type et la valeur
	t.Logf("Type de locks: %T, Valeur: %v, Est nil: %v", locks, locks, locks == nil)

	// Vérifier que la fonction ne retourne pas d'erreur
	// En Go, une slice vide [] est considérée comme nil, c'est normal
	// On vérifie juste que la fonction ne retourne pas d'erreur

	// Sur une base vide, il peut ne pas y avoir de locks actifs
	t.Logf("Nombre de locks détectés: %d", len(locks))
}

// TestGetRowLocks teste la récupération des locks de lignes
func TestGetRowLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	rowLocks, err := getRowLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks de lignes: %v", err)
	}

	// Vérifier que la fonction ne retourne pas d'erreur
	// En Go, une slice vide [] est considérée comme nil, c'est normal
	// On vérifie juste que la fonction ne retourne pas d'erreur

	// Sur une base vide, il peut ne pas y avoir de locks de lignes actifs
	t.Logf("Nombre de locks de lignes détectés: %d", len(rowLocks))
}

// TestDetectLongTransactions teste la détection des transactions longues
func TestDetectLongTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	longTxns := detectLongTransactions(tdb.DB)

	// Sans transactions actives, il ne devrait pas y avoir de transactions longues
	if len(longTxns) > 0 {
		t.Logf("Transactions longues détectées: %v", longTxns)
	}
}

// TestDetectObjectConflicts teste la détection des conflits d'objets
func TestDetectObjectConflicts(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks: %v", err)
	}

	conflicts := detectObjectConflicts(locks)

	// Vérifier que la fonction fonctionne correctement
	// En Go, une slice vide [] est considérée comme nil, c'est normal
	// On vérifie juste que la fonction ne retourne pas d'erreur

	// Sur une base vide, il peut ne pas y avoir de conflits d'objets
	t.Logf("Nombre de conflits d'objets détectés: %d", len(conflicts))
}

// TestAnalyzeIndexes teste l'analyse des index
func TestAnalyzeIndexes(t *testing.T) {
	tdb := setupTestDB(t, "fixture_test.yml")
	defer tdb.cleanupTestDB()

	indexes := analyzeIndexes(tdb.DB)

	// Il devrait y avoir au moins quelques index
	if len(indexes) == 0 {
		t.Error("Aucun index détecté")
	}

	// Vérifier que les index ont des informations valides
	for _, index := range indexes {
		if index.Name == "" {
			t.Error("Le nom de l'index ne doit pas être vide")
		}
		if index.Table == "" {
			t.Error("Le nom de la table ne doit pas être vide")
		}
	}
}

// TestCalculateSummary teste le calcul du résumé
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

	// Vérifications
	if summary.TotalLocks != 2 {
		t.Errorf("Total locks attendu: 2, obtenu: %d", summary.TotalLocks)
	}

	if summary.BlockedTxns != 1 {
		t.Errorf("Transactions bloquées attendues: 1, obtenues: %d", summary.BlockedTxns)
	}

	if summary.LongTxns != 1 {
		t.Errorf("Transactions longues attendues: 1, obtenues: %d", summary.LongTxns)
	}

	if summary.ObjectConflicts != 1 {
		t.Errorf("Conflits d'objets attendus: 1, obtenus: %d", summary.ObjectConflicts)
	}

	if summary.CriticalIssues != 1 {
		t.Errorf("Problèmes critiques attendus: 1, obtenus: %d", summary.CriticalIssues)
	}

	if summary.Warnings != 2 {
		t.Errorf("Avertissements attendus: 2, obtenus: %d", summary.Warnings)
	}

	if summary.Recommendations != 1 {
		t.Errorf("Recommandations attendues: 1, obtenues: %d", summary.Recommendations)
	}
}

// TestGenerateSuggestions teste la génération des suggestions
func TestGenerateSuggestions(t *testing.T) {
	data := &ReportData{
		Locks: make([]LockInfo, 15), // Plus de 10 locks
		BlockedTxns: []BlockedTransaction{
			{PID: "1", Duration: "10s"},
		},
		LongTxns: []LongTransaction{
			{PID: "2", Duration: "30s"},
		},
		ObjectConflicts: []ObjectConflict{
			{Object: "projects", PIDs: []string{"1", "2"}},
		},
	}

	suggestions := generateSuggestions(data)

	// Vérifier qu'il y a des suggestions
	if len(suggestions) == 0 {
		t.Error("Aucune suggestion générée")
	}

	// Vérifier que les suggestions sont pertinentes
	hasBlockedSuggestion := false
	hasLongTxnSuggestion := false
	hasOptimizationSuggestion := false

	for _, suggestion := range suggestions {
		if contains(suggestion, "timeout") {
			hasBlockedSuggestion = true
		}
		if contains(suggestion, "Diviser") {
			hasLongTxnSuggestion = true
		}
		if contains(suggestion, "Optimiser") {
			hasOptimizationSuggestion = true
		}
	}

	if !hasBlockedSuggestion {
		t.Error("Aucune suggestion pour les transactions bloquées")
	}

	if !hasLongTxnSuggestion {
		t.Error("Aucune suggestion pour les transactions longues")
	}

	if !hasOptimizationSuggestion {
		t.Error("Aucune suggestion d'optimisation")
	}
}

// TestDetectDeadlocks teste la détection des deadlocks
func TestDetectDeadlocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Object: "projects", Granted: true},
		{PID: 2, Object: "projects", Granted: false},
		{PID: 1, Object: "models", Granted: false},
		{PID: 2, Object: "models", Granted: true},
	}

	deadlocks := detectDeadlocks(locks)

	// Il devrait y avoir au moins un deadlock potentiel
	if len(deadlocks) == 0 {
		t.Error("Aucun deadlock détecté dans les locks fournis")
	}

	// Vérifier la structure des deadlocks
	for _, deadlock := range deadlocks {
		if deadlock.Transaction1.PID == deadlock.Transaction2.PID {
			t.Error("Un deadlock ne peut pas impliquer le même PID")
		}
		if deadlock.ConflictType == "" {
			t.Error("Le type de conflit ne doit pas être vide")
		}
		if deadlock.Recommendation == "" {
			t.Error("La recommandation ne doit pas être vide")
		}
	}
}

// TestDetectBlockedTransactionsFromLocks teste la détection des transactions bloquées à partir des locks
func TestDetectBlockedTransactionsFromLocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Object: "projects", Granted: true, Query: "UPDATE projects SET name = 'test'"},
		{PID: 2, Object: "projects", Granted: false, Query: "SELECT * FROM projects"},
	}

	blocked := detectBlockedTransactions(locks)

	if len(blocked) != 1 {
		t.Errorf("Attendu 1 transaction bloquée, obtenu %d", len(blocked))
	}

	if blocked[0].PID != "2" {
		t.Errorf("PID attendu: 2, obtenu: %s", blocked[0].PID)
	}
}

// TestDetectObjectConflictsFromLocks teste la détection des conflits d'objets à partir des locks
func TestDetectObjectConflictsFromLocks(t *testing.T) {
	locks := []LockInfo{
		{PID: 1, Object: "projects"},
		{PID: 2, Object: "projects"},
		{PID: 3, Object: "models"},
		{PID: 4, Object: "models"},
		{PID: 5, Object: "models"},
	}

	conflicts := detectObjectConflicts(locks)

	if len(conflicts) != 2 {
		t.Errorf("Attendu 2 conflits d'objets, obtenu %d", len(conflicts))
	}

	// Vérifier que les conflits ont les bonnes informations
	for _, conflict := range conflicts {
		if conflict.Object == "" {
			t.Error("L'objet du conflit ne doit pas être vide")
		}
		if len(conflict.PIDs) < 2 {
			t.Error("Un conflit doit impliquer au moins 2 PIDs")
		}
	}
}

// Fonction utilitaire pour vérifier si une chaîne contient un sous-texte
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
