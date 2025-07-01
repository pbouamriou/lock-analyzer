package lockanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Modèles de test partagés pour tous les tests du package lockanalyzer
// Ces modèles sont utilisés pour les fixtures Bun et les tests d'intégration

type Project struct {
	bun.BaseModel `bun:"table:projects"`
	ID            string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name          string
	ModifiedAt    time.Time `bun:",default:now()"`
}

type Model struct {
	bun.BaseModel `bun:"table:models"`
	ID            string   `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProjectID     string   `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Project       *Project `bun:"rel:belongs-to,join:project_id=id"`
	State         string
}

type File struct {
	bun.BaseModel `bun:"table:files"`
	ID            string   `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProjectID     string   `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Project       *Project `bun:"rel:belongs-to,join:project_id=id"`
	Content       string
}

type Block struct {
	bun.BaseModel `bun:"table:blocks"`
	ID            string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ModelID       string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Model         *Model `bun:"rel:belongs-to,join:model_id=id"`
	Type          string `bun:",notnull"`
	Name          string
}

type Parameter struct {
	bun.BaseModel `bun:"table:parameters"`
	ID            string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	BlockID       string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	Block         *Block `bun:"rel:belongs-to,join:block_id=id"`
	Key           string `bun:",column:key"`
	FileID        string `bun:",type:uuid,notnull,on_delete:CASCADE"`
	File          *File  `bun:"rel:belongs-to,join:file_id=id"`
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

	// Enregistrer les modèles avec Bun (comme recommandé dans la documentation officielle)
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

// createTables crée les tables nécessaires pour les tests
func createTables(t *testing.T, db *bun.DB) error {
	ctx := context.Background()

	// Créer les tables dans l'ordre (dépendances)
	tables := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			modified_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS models (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
			state TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
			content TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS blocks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			name TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS parameters (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			block_id UUID NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
			key TEXT,
			file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE
		)`,
	}

	for _, tableSQL := range tables {
		if _, err := db.ExecContext(ctx, tableSQL); err != nil {
			return fmt.Errorf("erreur lors de la création de la table: %v", err)
		}
	}

	return nil
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
