package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	"concurrent-db/formatters"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Project struct {
	ID         string `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name       string
	ModifiedAt time.Time `bun:",default:now()"`
}

// Custom type for context keys
type contextKey string

const (
	txIDKey contextKey = "tx_id"
)

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

// Available isolation levels:
// sql.LevelReadUncommitted  - Can read uncommitted data (dirty reads)
// sql.LevelReadCommitted    - Only reads committed data (PostgreSQL default)
// sql.LevelRepeatableRead   - Guarantees repeatable reads
// sql.LevelSerializable     - Most strict, avoids all anomalies
var isolationLevel = sql.LevelReadCommitted

func main() {
	dsn := "postgres://philippebouamriou@localhost:5432/testdb?sslmode=disable"
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	db := bun.NewDB(sqldb, pgdialect.New())
	ctx := context.Background()

	db.RegisterModel((*Project)(nil), (*Model)(nil), (*File)(nil), (*Block)(nil), (*Parameter)(nil))

	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	if err := fixture.Load(ctx, os.DirFS("testdata"), "fixture_example.yml"); err != nil {
		log.Fatalf("Error loading fixtures: %v", err)
	}

	// Create foreign key constraints manually
	if _, err := db.ExecContext(ctx, "ALTER TABLE models ADD CONSTRAINT fk_models_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE"); err != nil {
		log.Printf("Warning: FK constraint models.project_id already exists or error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE files ADD CONSTRAINT fk_files_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE"); err != nil {
		log.Printf("Warning: FK constraint files.project_id already exists or error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE blocks ADD CONSTRAINT fk_blocks_model_id FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Warning: FK constraint blocks.model_id already exists or error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE parameters ADD CONSTRAINT fk_parameters_block_id FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Warning: FK constraint parameters.block_id already exists or error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE parameters ADD CONSTRAINT fk_parameters_file_id FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Warning: FK constraint parameters.file_id already exists or error: %v", err)
	}

	// Create indexes on project_id to reproduce the lock problem
	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_models_project_id ON models(project_id)"); err != nil {
		log.Printf("Warning: index idx_models_project_id already exists or error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id)"); err != nil {
		log.Printf("Warning: index idx_files_project_id already exists or error: %v", err)
	}

	// Create simple trigger function to update project timestamp
	createTriggerFunction := `
		CREATE OR REPLACE FUNCTION update_project_timestamp() RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			-- For models, always update timestamp
			IF TG_TABLE_NAME = 'models' THEN
				IF (TG_OP = 'DELETE') THEN
					UPDATE projects AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.project_id;
				ELSE
					UPDATE projects AS p SET modified_at = current_timestamp FROM new_table AS n WHERE p.id = n.project_id;
				END IF;
			END IF;
			
			-- For files, update timestamp (simplified)
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
	if _, err := db.ExecContext(ctx, createTriggerFunction); err != nil {
		log.Printf("Warning: trigger function already exists or error: %v", err)
	}

	// Create triggers on models and files
	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS table_project_timestamp_update ON models"); err != nil {
		log.Printf("Warning: error deleting models trigger: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON models FOR EACH ROW EXECUTE FUNCTION update_project_timestamp()"); err != nil {
		log.Printf("Warning: error creating models trigger: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS table_project_timestamp_update ON files"); err != nil {
		log.Printf("Warning: error deleting files trigger: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON files FOR EACH ROW EXECUTE FUNCTION update_project_timestamp()"); err != nil {
		log.Printf("Warning: error creating files trigger: %v", err)
	}

	fmt.Println("Database initialized successfully (Bun fixtures)")

	// Retrieve example data for tests
	var projects []Project
	if err := db.NewSelect().Model(&projects).Scan(ctx); err != nil {
		log.Fatalf("Error retrieving projects: %v", err)
	}

	if len(projects) == 0 {
		log.Fatal("No projects found in database")
	}

	var models []Model
	if err := db.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(ctx); err != nil {
		log.Fatalf("Error retrieving models: %v", err)
	}

	var files []File
	if err := db.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(ctx); err != nil {
		log.Fatalf("Error retrieving files: %v", err)
	}

	if len(models) == 0 || len(files) == 0 {
		log.Fatal("No models or files found in database")
	}

	// Test transactions
	fmt.Println("Starting transaction test")

	// With different timeouts and identification values
	ctx1, cancel1 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel1()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel2()

	// Test with shared indexes
	testWithSharedIndexes(db, ctx1, ctx2)
}

func testWithSharedIndexes(db *bun.DB, ctx1, ctx2 context.Context) {
	// Retrieve data
	var projects []Project
	db.NewSelect().Model(&projects).Scan(ctx1)
	var models []Model
	db.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(ctx1)
	var files []File
	db.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(ctx1)

	// Retrieve blocks and parameters to understand associations
	var blocks []Block
	db.NewSelect().Model(&blocks).Where("model_id = ?", models[0].ID).Scan(ctx1)
	var parameters []Parameter
	db.NewSelect().Model(&parameters).Scan(ctx1)

	model := models[0]

	// Find file linked to a STANDARD block (must update timestamp)
	var userFile File
	var generatedFile File

	for _, param := range parameters {
		for _, block := range blocks {
			if param.BlockID == block.ID {
				switch block.Type {
				case "STANDARD":
					userFile = files[0] // file1
				case "GENERATED":
					generatedFile = files[1] // file2
				}
			}
		}
	}

	fmt.Printf("Test with project_id: %s\n", projects[0].ID)
	fmt.Printf("Model ID: %s\n", model.ID)
	fmt.Printf("User File ID: %s (linked to STANDARD block)\n", userFile.ID)
	fmt.Printf("Generated File ID: %s (linked to GENERATED block)\n", generatedFile.ID)

	// Test 1: T1 modifies models, T2 modifies STANDARD file (should cause lock)
	fmt.Println("\n=== Test 1: T1 (models) vs T2 (STANDARD file) ===")
	testConcurrentTransactions(db, ctx1, ctx2, model, userFile, "STANDARD")

	// Test 2: T1 modifies models, T2 modifies GENERATED file (should not cause lock)
	fmt.Println("\n=== Test 2: T1 (models) vs T2 (GENERATED file) ===")
	testConcurrentTransactions(db, ctx1, ctx2, model, generatedFile, "GENERATED")
}

func testConcurrentTransactions(db *bun.DB, ctx1, ctx2 context.Context, model Model, file File, fileType string) {
	// Transaction T1: UPDATE models (long transaction)
	tx1, err := db.BeginTx(ctx1, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  false,
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx1.NewUpdate().Model(&Model{ID: model.ID, State: "updated state"}).Column("state").WherePK().Exec(ctx1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction T1 (UPDATE models) started - Trigger will update projects.modified_at\n")

	// Display locks before T2
	fmt.Println("\n--- Locks state before T2 ---")
	markdownFormatter := formatters.NewMarkdownFormatter("")
	if err := formatters.GenerateAndDisplayReport(db, markdownFormatter); err != nil {
		log.Printf("Error displaying report: %v", err)
	}

	// Transaction T2: UPDATE files (short transaction)
	tx2, err := db.BeginTx(ctx2, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Attempting T2 (UPDATE files %s) - Trigger will ", fileType)
	if fileType == "STANDARD" {
		fmt.Print("also")
	} else {
		fmt.Print("NOT")
	}
	fmt.Print(" update projects.modified_at...\n")

	// Start real-time lock analysis in a goroutine
	stopAnalysis := make(chan bool)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("\n--- Real-time locks analysis ---")
				// Display report in markdown format
				markdownFormatter := formatters.NewMarkdownFormatter("")
				if err := formatters.GenerateAndDisplayReport(db, markdownFormatter); err != nil {
					fmt.Printf("Error displaying report: %v\n", err)
				}
			case <-stopAnalysis:
				return
			}
		}
	}()

	// Execute T2 and see if it's blocked
	_, err = tx2.NewUpdate().Model(&File{ID: file.ID, Content: "updated content"}).Column("content").WherePK().Exec(ctx2)

	// Stop real-time analysis
	stopAnalysis <- true

	if err != nil {
		fmt.Printf("T2 blocked (expected lock on projects): %v\n", err)
	} else {
		fmt.Printf("Transaction T2 (UPDATE files %s) succeeded - No lock on projects\n", fileType)
	}

	// Display locks after T2
	fmt.Println("\n--- Locks state after T2 ---")
	postMarkdownFormatter := formatters.NewMarkdownFormatter("")
	if err := formatters.GenerateAndDisplayReport(db, postMarkdownFormatter); err != nil {
		log.Printf("Error displaying report: %v", err)
	}

	fmt.Println("Pause for 5 seconds")
	time.Sleep(5 * time.Second)

	// Validate transactions
	if err := tx1.Commit(); err != nil {
		log.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test %s completed\n", fileType)

	// Generate complete analysis report after test
	fmt.Println("\n--- Generating analysis report ---")

	// Generate final reports
	finalTextFormatter := formatters.NewTextFormatter("")
	finalJSONFormatter := formatters.NewJSONFormatter("")
	finalMarkdownFormatter := formatters.NewMarkdownFormatter("")

	// Generate text report
	if err := formatters.GenerateAndWriteReport(db, finalTextFormatter, fmt.Sprintf("lock_report_%s.txt", fileType)); err != nil {
		log.Printf("Error generating text report: %v", err)
	} else {
		fmt.Printf("Text report generated: lock_report_%s.txt\n", fileType)
	}

	// Generate JSON report
	if err := formatters.GenerateAndWriteReport(db, finalJSONFormatter, fmt.Sprintf("lock_report_%s.json", fileType)); err != nil {
		log.Printf("Error generating JSON report: %v", err)
	} else {
		fmt.Printf("JSON report generated: lock_report_%s.json\n", fileType)
	}

	// Generate Markdown report
	if err := formatters.GenerateAndWriteReport(db, finalMarkdownFormatter, fmt.Sprintf("lock_report_%s.md", fileType)); err != nil {
		log.Printf("Error generating Markdown report: %v", err)
	} else {
		fmt.Printf("Markdown report generated: lock_report_%s.md\n", fileType)
	}
}
