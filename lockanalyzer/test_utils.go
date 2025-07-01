package lockanalyzer

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Shared test models for all tests in the lockanalyzer package
// These models are used for Bun fixtures and integration tests

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

// TestDB contains a test database with fixtures
type TestDB struct {
	DB *bun.DB
}

// setupTestDB configures a test database with fixtures
func setupTestDB(t *testing.T, fixtureFile string) *TestDB {
	dsn := "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable"
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Database connection error: %v", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	// Register models with Bun (as recommended in the official documentation)
	db.RegisterModel((*Project)(nil), (*Model)(nil), (*File)(nil), (*Block)(nil), (*Parameter)(nil))

	// Load fixtures
	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	if err := fixture.Load(context.Background(), os.DirFS("../testdata"), fixtureFile); err != nil {
		t.Fatalf("Error loading fixtures: %v", err)
	}

	// Create constraints and triggers
	setupTestConstraints(t, db)

	return &TestDB{DB: db}
}

// setupTestConstraints configures constraints and triggers for tests
func setupTestConstraints(t *testing.T, db *bun.DB) {
	ctx := context.Background()

	// FK constraints
	constraints := []string{
		"ALTER TABLE models ADD CONSTRAINT fk_models_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE",
		"ALTER TABLE files ADD CONSTRAINT fk_files_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE",
		"ALTER TABLE blocks ADD CONSTRAINT fk_blocks_model_id FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE",
		"ALTER TABLE parameters ADD CONSTRAINT fk_parameters_block_id FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE",
		"ALTER TABLE parameters ADD CONSTRAINT fk_parameters_file_id FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE",
	}

	for _, constraint := range constraints {
		if _, err := db.ExecContext(ctx, constraint); err != nil {
			t.Logf("Constraint already exists or error: %v", err)
		}
	}

	// Indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_models_project_id ON models(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id)",
	}

	for _, index := range indexes {
		if _, err := db.ExecContext(ctx, index); err != nil {
			t.Logf("Index already exists or error: %v", err)
		}
	}

	// Trigger function
	triggerFunction := `
		CREATE OR REPLACE FUNCTION update_project_timestamp() RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			IF TG_TABLE_NAME = 'models' THEN
				UPDATE projects SET modified_at = current_timestamp WHERE id = NEW.project_id;
			END IF;
			
			IF TG_TABLE_NAME = 'files' THEN
				UPDATE projects SET modified_at = current_timestamp WHERE id = NEW.project_id;
			END IF;
			
			RETURN NEW;
		END;
		$$;
	`
	if _, err := db.ExecContext(ctx, triggerFunction); err != nil {
		t.Logf("Trigger function already exists or error: %v", err)
	}

	// Triggers
	triggers := []string{
		"DROP TRIGGER IF EXISTS trigger_update_project_timestamp_models ON models",
		"CREATE TRIGGER trigger_update_project_timestamp_models AFTER INSERT OR UPDATE OR DELETE ON models FOR EACH ROW EXECUTE FUNCTION update_project_timestamp()",
		"DROP TRIGGER IF EXISTS trigger_update_project_timestamp_files ON files",
		"CREATE TRIGGER trigger_update_project_timestamp_files AFTER INSERT OR UPDATE OR DELETE ON files FOR EACH ROW EXECUTE FUNCTION update_project_timestamp()",
	}

	for _, trigger := range triggers {
		if _, err := db.ExecContext(ctx, trigger); err != nil {
			t.Logf("Trigger already exists or error: %v", err)
		}
	}
}

// cleanupTestDB closes the database connection
func (tdb *TestDB) cleanupTestDB() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}
