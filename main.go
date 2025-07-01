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

// Type personnalisé pour les clés de contexte
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

// Niveaux d'isolation disponibles :
// sql.LevelReadUncommitted  - Peut lire les données non validées (lectures sales)
// sql.LevelReadCommitted    - Ne lit que les données validées (défaut PostgreSQL)
// sql.LevelRepeatableRead   - Garantit des lectures répétables
// sql.LevelSerializable     - Le plus strict, évite toutes les anomalies
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
	if err := fixture.Load(ctx, os.DirFS("testdata"), "fixture.yml"); err != nil {
		log.Fatalf("Erreur lors du chargement des fixtures : %v", err)
	}

	// Créer les contraintes de clé étrangère manuellement
	if _, err := db.ExecContext(ctx, "ALTER TABLE models ADD CONSTRAINT fk_models_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE"); err != nil {
		log.Printf("Attention: contrainte FK models.project_id déjà existante ou erreur: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE files ADD CONSTRAINT fk_files_project_id FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE"); err != nil {
		log.Printf("Attention: contrainte FK files.project_id déjà existante ou erreur: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE blocks ADD CONSTRAINT fk_blocks_model_id FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Attention: contrainte FK blocks.model_id déjà existante ou erreur: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE parameters ADD CONSTRAINT fk_parameters_block_id FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Attention: contrainte FK parameters.block_id déjà existante ou erreur: %v", err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE parameters ADD CONSTRAINT fk_parameters_file_id FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE"); err != nil {
		log.Printf("Attention: contrainte FK parameters.file_id déjà existante ou erreur: %v", err)
	}

	// Créer des index sur project_id pour reproduire le problème de lock
	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_models_project_id ON models(project_id)"); err != nil {
		log.Printf("Attention: index idx_models_project_id déjà existant ou erreur: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id)"); err != nil {
		log.Printf("Attention: index idx_files_project_id déjà existant ou erreur: %v", err)
	}

	// Créer la fonction trigger pour mettre à jour le timestamp du projet
	createTriggerFunction := `
		CREATE OR REPLACE FUNCTION update_project_timestamp() RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			-- Pour les models, toujours mettre à jour le timestamp
			IF TG_TABLE_NAME = 'models' THEN
				IF (TG_OP = 'DELETE') THEN
					UPDATE projects AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.project_id;
				ELSE
					UPDATE projects AS p SET modified_at = current_timestamp FROM new_table AS n WHERE p.id = n.project_id;
				END IF;
			END IF;
			
			-- Pour les files, vérifier si le fichier est lié à des paramètres de blocs non-GENERATED
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
		log.Printf("Attention: fonction trigger déjà existante ou erreur: %v", err)
	}

	// Créer les triggers sur models et files
	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS table_project_timestamp_update ON models"); err != nil {
		log.Printf("Attention: erreur lors de la suppression du trigger models: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON models REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION update_project_timestamp()"); err != nil {
		log.Printf("Attention: erreur lors de la création du trigger models: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS table_project_timestamp_update ON files"); err != nil {
		log.Printf("Attention: erreur lors de la suppression du trigger files: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TRIGGER table_project_timestamp_update AFTER UPDATE ON files REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE FUNCTION update_project_timestamp()"); err != nil {
		log.Printf("Attention: erreur lors de la création du trigger files: %v", err)
	}

	fmt.Println("Base de données initialisée avec succès (fixtures Bun)")

	// Récupérer les données d'exemple pour les tests
	var projects []Project
	if err := db.NewSelect().Model(&projects).Scan(ctx); err != nil {
		log.Fatalf("Erreur lors de la récupération des projets: %v", err)
	}

	if len(projects) == 0 {
		log.Fatal("Aucun projet trouvé dans la base de données")
	}

	var models []Model
	if err := db.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(ctx); err != nil {
		log.Fatalf("Erreur lors de la récupération des modèles: %v", err)
	}

	var files []File
	if err := db.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(ctx); err != nil {
		log.Fatalf("Erreur lors de la récupération des fichiers: %v", err)
	}

	if len(models) == 0 || len(files) == 0 {
		log.Fatal("Aucun modèle ou fichier trouvé dans la base de données")
	}

	// Test des transactions
	fmt.Println("Démarrage du test de transaction")

	// Avec des timeouts différents et des valeurs d'identification
	ctx1, cancel1 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel1()
	ctx1 = context.WithValue(ctx1, txIDKey, "T1")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel2()
	ctx2 = context.WithValue(ctx2, txIDKey, "T2")

	// Test avec index partagés - devrait causer un lock
	fmt.Println("\n--- Test avec triggers de timestamp sur projects ---")
	testWithSharedIndexes(db, ctx1, ctx2)
}

func testWithSharedIndexes(db *bun.DB, ctx1, ctx2 context.Context) {
	// Récupérer les données
	var projects []Project
	db.NewSelect().Model(&projects).Scan(ctx1)
	var models []Model
	db.NewSelect().Model(&models).Where("project_id = ?", projects[0].ID).Scan(ctx1)
	var files []File
	db.NewSelect().Model(&files).Where("project_id = ?", projects[0].ID).Scan(ctx1)

	// Récupérer les blocs et paramètres pour comprendre les associations
	var blocks []Block
	db.NewSelect().Model(&blocks).Where("model_id = ?", models[0].ID).Scan(ctx1)
	var parameters []Parameter
	db.NewSelect().Model(&parameters).Scan(ctx1)

	model := models[0]

	// Trouver le fichier lié à un bloc STANDARD (doit mettre à jour le timestamp)
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

	fmt.Printf("Test avec project_id: %s\n", projects[0].ID)
	fmt.Printf("Model ID: %s\n", model.ID)
	fmt.Printf("User File ID: %s (lié à bloc STANDARD)\n", userFile.ID)
	fmt.Printf("Generated File ID: %s (lié à bloc GENERATED)\n", generatedFile.ID)

	// Test 1: T1 modifie models, T2 modifie fichier STANDARD (devrait causer un lock)
	fmt.Println("\n=== Test 1: T1 (models) vs T2 (fichier STANDARD) ===")
	testConcurrentTransactions(db, ctx1, ctx2, model, userFile, "STANDARD")

	// Test 2: T1 modifie models, T2 modifie fichier GENERATED (ne devrait pas causer de lock)
	fmt.Println("\n=== Test 2: T1 (models) vs T2 (fichier GENERATED) ===")
	testConcurrentTransactions(db, ctx1, ctx2, model, generatedFile, "GENERATED")
}

func testConcurrentTransactions(db *bun.DB, ctx1, ctx2 context.Context, model Model, file File, fileType string) {
	// Transaction T1: UPDATE models (longue transaction)
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
	fmt.Printf("Transaction T1 (UPDATE models) démarrée - Trigger va mettre à jour projects.modified_at\n")

	// Afficher les locks avant T2
	fmt.Println("\n--- État des locks avant T2 ---")
	markdownFormatter := formatters.NewMarkdownFormatter()
	if err := formatters.GenerateAndDisplayReport(db, markdownFormatter); err != nil {
		log.Printf("Erreur lors de l'affichage du rapport: %v", err)
	}

	// Transaction T2: UPDATE files (courte transaction)
	tx2, err := db.BeginTx(ctx2, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tentative de T2 (UPDATE files %s) - Trigger va ", fileType)
	if fileType == "STANDARD" {
		fmt.Print("aussi")
	} else {
		fmt.Print("NE PAS")
	}
	fmt.Print(" mettre à jour projects.modified_at...\n")

	// Démarrer l'analyse des locks en temps réel dans une goroutine
	stopAnalysis := make(chan bool)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Println("\n--- Analyse des locks en temps réel ---")
				// Afficher le rapport en format markdown
				markdownFormatter := formatters.NewMarkdownFormatter()
				if err := formatters.GenerateAndDisplayReport(db, markdownFormatter); err != nil {
					fmt.Printf("Erreur lors de l'affichage du rapport: %v\n", err)
				}
			case <-stopAnalysis:
				return
			}
		}
	}()

	// Exécuter T2 et voir si elle est bloquée
	_, err = tx2.NewUpdate().Model(&File{ID: file.ID, Content: "updated content"}).Column("content").WherePK().Exec(ctx2)

	// Arrêter l'analyse en temps réel
	stopAnalysis <- true

	if err != nil {
		fmt.Printf("T2 bloquée (lock sur projects attendu): %v\n", err)
	} else {
		fmt.Printf("Transaction T2 (UPDATE files %s) réussie - Pas de lock sur projects\n", fileType)
	}

	// Afficher les locks après T2
	fmt.Println("\n--- État des locks après T2 ---")
	postMarkdownFormatter := formatters.NewMarkdownFormatter()
	if err := formatters.GenerateAndDisplayReport(db, postMarkdownFormatter); err != nil {
		log.Printf("Erreur lors de l'affichage du rapport: %v", err)
	}

	fmt.Println("Pause de 5 secondes")
	time.Sleep(5 * time.Second)

	// Valider les transactions
	if err := tx1.Commit(); err != nil {
		log.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test %s terminé\n", fileType)

	// Générer un rapport complet après le test
	fmt.Println("\n--- Génération du rapport d'analyse ---")

	// Générer les rapports finaux
	finalTextFormatter := formatters.NewTextFormatter()
	finalJSONFormatter := formatters.NewJSONFormatter()
	finalMarkdownFormatter := formatters.NewMarkdownFormatter()

	// Générer le rapport texte
	if err := formatters.GenerateAndWriteReport(db, finalTextFormatter, fmt.Sprintf("lock_report_%s.txt", fileType)); err != nil {
		log.Printf("Erreur lors de la génération du rapport texte: %v", err)
	} else {
		fmt.Printf("Rapport texte généré: lock_report_%s.txt\n", fileType)
	}

	// Générer le rapport JSON
	if err := formatters.GenerateAndWriteReport(db, finalJSONFormatter, fmt.Sprintf("lock_report_%s.json", fileType)); err != nil {
		log.Printf("Erreur lors de la génération du rapport JSON: %v", err)
	} else {
		fmt.Printf("Rapport JSON généré: lock_report_%s.json\n", fileType)
	}

	// Générer le rapport Markdown
	if err := formatters.GenerateAndWriteReport(db, finalMarkdownFormatter, fmt.Sprintf("lock_report_%s.md", fileType)); err != nil {
		log.Printf("Erreur lors de la génération du rapport Markdown: %v", err)
	} else {
		fmt.Printf("Rapport Markdown généré: lock_report_%s.md\n", fileType)
	}
}
