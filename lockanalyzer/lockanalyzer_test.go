package lockanalyzer

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Modèles pour les tests (copiés depuis cmd/example/main.go)
type Project struct {
	ID         string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name       string
	ModifiedAt time.Time `bun:",default:now()"`
}

type Model struct {
	ID        string   `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProjectID string   `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Project   *Project `bun:"rel:belongs-to,join:project_id=id"`
	State     string
}

type File struct {
	ID        string   `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProjectID string   `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Project   *Project `bun:"rel:belongs-to,join:project_id=id"`
	Content   string
}

type Block struct {
	ID      string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ModelID string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Model   *Model `bun:"rel:belongs-to,join:model_id=id"`
	Type    string `bun:",notnull"` // 'GENERATED', 'STANDARD', 'SUBSYSTEM', etc.
	Name    string
}

type Parameter struct {
	ID      string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	BlockID string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Block   *Block `bun:"rel:belongs-to,join:block_id=id"`
	Key     string `bun:",column:key"`
	FileID  string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	File    *File  `bun:"rel:belongs-to,join:file_id=id"`
}

// TestDB contient une base de données de test avec fixtures
type TestDB struct {
	DB *bun.DB
}

// setupTestDB configure une base de données de test avec fixtures
func setupTestDB(t *testing.T, fixtureFile string) *TestDB {
	dsn := "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable"
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Erreur de connexion à la base de données: %v", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	// Enregistrer les modèles
	db.RegisterModel((*Project)(nil), (*Model)(nil), (*File)(nil), (*Block)(nil), (*Parameter)(nil))

	// Charger les fixtures
	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	if err := fixture.Load(context.Background(), os.DirFS("../testdata"), fixtureFile); err != nil {
		t.Fatalf("Erreur lors du chargement des fixtures: %v", err)
	}

	// Créer les contraintes et triggers
	setupTestConstraints(t, db)

	return &TestDB{DB: db}
}

// setupTestConstraints configure les contraintes et triggers pour les tests
func setupTestConstraints(t *testing.T, db *bun.DB) {
	ctx := context.Background()

	// Contraintes FK
	constraints := []string{
		"ALTER TABLE models ADD CONSTRAINT fk_models_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE",
		"ALTER TABLE files ADD CONSTRAINT fk_files_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE",
		"ALTER TABLE blocks ADD CONSTRAINT fk_blocks_model_id FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE",
		"ALTER TABLE parameters ADD CONSTRAINT fk_parameters_block_id FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE",
		"ALTER TABLE parameters ADD CONSTRAINT fk_parameters_file_id FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE",
	}

	for _, constraint := range constraints {
		if _, err := db.ExecContext(ctx, constraint); err != nil {
			t.Logf("Contrainte déjà existante ou erreur: %v", err)
		}
	}

	// Index
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_models_project_id ON models(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id)",
	}

	for _, index := range indexes {
		if _, err := db.ExecContext(ctx, index); err != nil {
			t.Logf("Index déjà existant ou erreur: %v", err)
		}
	}

	// Trigger function
	triggerFunction := `
		CREATE OR REPLACE FUNCTION update_project_timestamp() RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			IF TG_TABLE_NAME = 'models' THEN
				IF (TG_OP = 'DELETE') THEN
					UPDATE projects AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.project_id;
				ELSE
					UPDATE projects AS p SET modified_at = current_timestamp FROM new_table AS n WHERE p.id = n.project_id;
				END IF;
			END IF;
			
			IF TG_TABLE_NAME = 'files' THEN
				IF (TG_OP = 'DELETE') THEN
					UPDATE projects AS p SET modified_at = current_timestamp 
					FROM old_table AS o 
					JOIN parameters param ON param.file_id = o.id
					JOIN blocks b ON b.id = param.block_id
					WHERE p.id = o.project_id AND b.type != 'GENERATED';
				ELSE
					UPDATE projects AS p SET modified_at = current_timestamp 
					FROM new_table AS n 
					JOIN parameters param ON param.file_id = n.id
					JOIN blocks b ON b.id = param.block_id
					WHERE p.id = n.project_id AND b.type != 'GENERATED';
				END IF;
			END IF;
			
			RETURN NULL;
		END;
		$$;
	`
	if _, err := db.ExecContext(ctx, triggerFunction); err != nil {
		t.Logf("Fonction trigger déjà existante ou erreur: %v", err)
	}

	// Triggers
	triggers := []string{
		"DROP TRIGGER IF EXISTS table_project_timestamp_update ON models",
		"CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION update_project_timestamp()",
		"DROP TRIGGER IF EXISTS table_project_timestamp_update ON files",
		"CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON files REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION update_project_timestamp()",
	}

	for _, trigger := range triggers {
		if _, err := db.ExecContext(ctx, trigger); err != nil {
			t.Logf("Trigger déjà existant ou erreur: %v", err)
		}
	}
}

// cleanupTestDB ferme la connexion à la base de données
func (tdb *TestDB) cleanupTestDB() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// TestGenerateLocksReport teste la génération de rapport sans locks
func TestGenerateLocksReport(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
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
	tdb := setupTestDB(t, "fixture.yml")
	defer tdb.cleanupTestDB()

	blocked := DetectBlockedTransactions(tdb.DB)

	// Sans transactions actives, il ne devrait pas y avoir de transactions bloquées
	if len(blocked) > 0 {
		t.Logf("Transactions bloquées détectées: %v", blocked)
	}
}

// TestGetLocks teste la récupération des locks
func TestGetLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks: %v", err)
	}

	// Vérifier que la fonction ne retourne pas d'erreur
	if locks == nil {
		t.Error("La liste des locks ne doit pas être nil")
	}
}

// TestGetRowLocks teste la récupération des locks de lignes
func TestGetRowLocks(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
	defer tdb.cleanupTestDB()

	rowLocks, err := getRowLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks de lignes: %v", err)
	}

	// Vérifier que la fonction ne retourne pas d'erreur
	if rowLocks == nil {
		t.Error("La liste des locks de lignes ne doit pas être nil")
	}
}

// TestDetectLongTransactions teste la détection des transactions longues
func TestDetectLongTransactions(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
	defer tdb.cleanupTestDB()

	longTxns := detectLongTransactions(tdb.DB)

	// Sans transactions actives, il ne devrait pas y avoir de transactions longues
	if len(longTxns) > 0 {
		t.Logf("Transactions longues détectées: %v", longTxns)
	}
}

// TestDetectObjectConflicts teste la détection des conflits d'objets
func TestDetectObjectConflicts(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
	defer tdb.cleanupTestDB()

	locks, err := getLocks(tdb.DB)
	if err != nil {
		t.Fatalf("Erreur lors de la récupération des locks: %v", err)
	}

	conflicts := detectObjectConflicts(locks)

	// Vérifier que la fonction fonctionne correctement
	if conflicts == nil {
		t.Error("La liste des conflits ne doit pas être nil")
	}
}

// TestAnalyzeIndexes teste l'analyse des index
func TestAnalyzeIndexes(t *testing.T) {
	tdb := setupTestDB(t, "fixture.yml")
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
