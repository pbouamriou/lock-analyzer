package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

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
			IF (TG_OP = 'DELETE') THEN
				UPDATE projects AS p SET modified_at = current_timestamp FROM old_table AS o WHERE p.id = o.project_id;
			ELSE
				UPDATE projects AS p SET modified_at = current_timestamp FROM new_table AS n WHERE p.id = n.project_id;
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

	model := models[0]
	file := files[0]

	fmt.Printf("Test avec project_id: %s\n", projects[0].ID)
	fmt.Printf("Model ID: %s, File ID: %s\n", model.ID, file.ID)

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
	fmt.Println("Transaction T1 (UPDATE models) démarrée - Trigger va mettre à jour projects.modified_at")

	// Afficher les locks avant T2
	showLocks(db)

	// Transaction T2: UPDATE files (courte transaction) - devrait être bloquée
	tx2, err := db.BeginTx(ctx2, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Tentative de T2 (UPDATE files) - Trigger va aussi mettre à jour projects.modified_at...")
	_, err = tx2.NewUpdate().Model(&File{ID: file.ID, Content: "updated content"}).Column("content").WherePK().Exec(ctx2)
	if err != nil {
		fmt.Printf("T2 bloquée (lock sur projects attendu): %v\n", err)
	} else {
		fmt.Println("Transaction T2 (UPDATE files) réussie - Pas de lock sur projects")
	}

	// Afficher les locks après T2
	showLocks(db)

	fmt.Println("Pause de 10 secondes")
	time.Sleep(10 * time.Second)

	// Valider les transactions
	if err := tx1.Commit(); err != nil {
		log.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Test terminé")
}

func showLocks(db *bun.DB) {
	query := `
		SELECT 
			l.pid,
			l.mode,
			l.granted,
			CASE 
				WHEN l.relation IS NOT NULL THEN t.relname
				WHEN l.database IS NOT NULL THEN 'database'
				WHEN l.page IS NOT NULL THEN 'page'
				WHEN l.tuple IS NOT NULL THEN 'tuple'
				WHEN l.virtualxid IS NOT NULL THEN 'virtualxid'
				WHEN l.transactionid IS NOT NULL THEN 'transactionid'
				WHEN l.classid IS NOT NULL AND l.objid IS NOT NULL THEN 
					CASE l.classid::regclass::text
						WHEN 'pg_class' THEN 'table/index'
						WHEN 'pg_namespace' THEN 'schema'
						WHEN 'pg_database' THEN 'database'
						WHEN 'pg_tablespace' THEN 'tablespace'
						WHEN 'pg_foreign_data_wrapper' THEN 'fdw'
						WHEN 'pg_foreign_server' THEN 'foreign_server'
						WHEN 'pg_policy' THEN 'policy'
						WHEN 'pg_publication' THEN 'publication'
						WHEN 'pg_subscription' THEN 'subscription'
						ELSE 'other_object'
					END
				ELSE 'unknown'
			END as object_type,
			CASE 
				WHEN l.relation IS NOT NULL THEN t.relname
				WHEN l.classid IS NOT NULL AND l.objid IS NOT NULL THEN 
					CASE l.classid::regclass::text
						WHEN 'pg_class' THEN (SELECT relname FROM pg_class WHERE oid = l.objid)
						WHEN 'pg_namespace' THEN (SELECT nspname FROM pg_namespace WHERE oid = l.objid)
						WHEN 'pg_database' THEN (SELECT datname FROM pg_database WHERE oid = l.objid)
						ELSE l.objid::text
					END
				ELSE 'N/A'
			END as object_name,
			l.page,
			l.tuple,
			l.virtualxid,
			l.transactionid,
			l.classid,
			l.objid,
			l.objsubid,
			l.virtualtransaction
		FROM pg_locks l
		LEFT JOIN pg_class t ON l.relation = t.oid
		WHERE l.pid != pg_backend_pid()
		ORDER BY l.pid, l.mode;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de la requête des locks: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Locks actifs ===")
	for rows.Next() {
		var pid int
		var mode, granted, objectType, objectName, page, tuple, virtualxid, transactionid, classid, objid, objsubid, virtualtransaction sql.NullString

		rows.Scan(&pid, &mode, &granted, &objectType, &objectName, &page, &tuple, &virtualxid, &transactionid, &classid, &objid, &objsubid, &virtualtransaction)

		// Construire une description plus détaillée
		description := fmt.Sprintf("PID: %d, Mode: %s, Granted: %s", pid, mode.String, granted.String)

		if objectType.String != "" {
			description += fmt.Sprintf(", Type: %s", objectType.String)
		}

		if objectName.String != "" && objectName.String != "N/A" {
			description += fmt.Sprintf(", Object: %s", objectName.String)
		}

		if page.String != "" {
			description += fmt.Sprintf(", Page: %s", page.String)
		}

		if tuple.String != "" {
			description += fmt.Sprintf(", Tuple: %s", tuple.String)
		}

		if virtualxid.String != "" {
			description += fmt.Sprintf(", VirtualXID: %s", virtualxid.String)
		}

		if transactionid.String != "" {
			description += fmt.Sprintf(", TransactionID: %s", transactionid.String)
		}

		fmt.Println(description)
	}

	// Afficher les détails des lignes lockées si il y en a
	fmt.Println("\n=== Détails des lignes lockées ===")
	showLockedRows(db)

	// Afficher l'explication des locks
	fmt.Println("\n=== Explication des locks ===")
	explainLocks(db)
}

func explainLocks(db *bun.DB) {
	query := `
		SELECT 
			l.pid,
			l.mode,
			l.granted,
			CASE 
				WHEN l.virtualxid IS NOT NULL THEN 'virtualxid'
				WHEN l.transactionid IS NOT NULL THEN 'transactionid'
				WHEN l.relation IS NOT NULL THEN 'relation'
				WHEN l.page IS NOT NULL THEN 'page'
				WHEN l.tuple IS NOT NULL THEN 'tuple'
				ELSE 'other'
			END as lock_category,
			l.virtualxid,
			l.transactionid,
			t.relname as table_name
		FROM pg_locks l
		LEFT JOIN pg_class t ON l.relation = t.oid
		WHERE l.pid != pg_backend_pid()
		AND l.mode = 'ExclusiveLock'
		ORDER BY l.pid, lock_category;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'explication des locks: %v", err)
		return
	}
	defer rows.Close()

	currentPID := 0
	for rows.Next() {
		var pid int
		var mode, granted, lockCategory, virtualxid, transactionid, tableName sql.NullString

		rows.Scan(&pid, &mode, &granted, &lockCategory, &virtualxid, &transactionid, &tableName)

		if pid != currentPID {
			fmt.Printf("\n--- PID %d ---\n", pid)
			currentPID = pid
		}

		switch lockCategory.String {
		case "virtualxid":
			fmt.Printf("  ExclusiveLock sur VirtualXID (%s):\n", virtualxid.String)
			fmt.Printf("    → Chaque transaction a un identifiant virtuel unique\n")
			fmt.Printf("    → Ce lock empêche d'autres transactions d'utiliser le même VirtualXID\n")
			fmt.Printf("    → Normal et attendu pour toute transaction\n\n")

		case "transactionid":
			fmt.Printf("  ExclusiveLock sur TransactionID (%s):\n", transactionid.String)
			fmt.Printf("    → Identifiant unique de la transaction dans le système\n")
			fmt.Printf("    → Utilisé pour la gestion MVCC (Multi-Version Concurrency Control)\n")
			fmt.Printf("    → Empêche les conflits de numérotation des transactions\n")
			fmt.Printf("    → Normal et attendu pour toute transaction\n\n")

		case "relation":
			fmt.Printf("  ExclusiveLock sur Relation (%s):\n", tableName.String)
			fmt.Printf("    → Lock sur la structure de la table/index\n")
			fmt.Printf("    → Empêche les opérations DDL (ALTER, DROP, etc.)\n")
			fmt.Printf("    → Différent du RowExclusiveLock (qui permet les modifications de données)\n")
			fmt.Printf("    → Peut indiquer une opération de maintenance\n\n")

		case "page":
			fmt.Printf("  ExclusiveLock sur Page:\n")
			fmt.Printf("    → Lock sur une page spécifique du stockage\n")
			fmt.Printf("    → Utilisé lors de modifications structurelles de la page\n")
			fmt.Printf("    → Peut indiquer une opération de maintenance ou de réorganisation\n\n")

		case "tuple":
			fmt.Printf("  ExclusiveLock sur Tuple:\n")
			fmt.Printf("    → Lock sur une ligne spécifique\n")
			fmt.Printf("    → Indique un conflit d'accès concurrent à la même ligne\n")
			fmt.Printf("    → Peut causer des deadlocks\n\n")

		default:
			fmt.Printf("  ExclusiveLock sur autre objet:\n")
			fmt.Printf("    → Lock sur un objet système non standard\n")
			fmt.Printf("    → Peut être lié à des opérations de maintenance\n\n")
		}
	}
}

func showLockedRows(db *bun.DB) {
	// Exécuter la requête pour chaque table
	tables := []string{"models", "files"}
	for _, table := range tables {
		// Récupérer les locks pour cette table
		lockQuery := fmt.Sprintf(`
			SELECT DISTINCT
				l.pid,
				l.page,
				l.tuple
			FROM pg_locks l
			JOIN pg_class t ON l.relation = t.oid
			WHERE l.pid != pg_backend_pid()
			AND l.page IS NOT NULL 
			AND l.tuple IS NOT NULL
			AND t.relname = '%s'
		`, table)

		rows, err := db.Query(lockQuery)
		if err != nil {
			log.Printf("Erreur lors de la requête des locks pour %s: %v", table, err)
			continue
		}

		hasLocks := false
		for rows.Next() {
			var pid int
			var page, tuple sql.NullString
			rows.Scan(&pid, &page, &tuple)

			if !hasLocks {
				fmt.Printf("\nLocks sur la table %s:\n", table)
				hasLocks = true
			}

			fmt.Printf("  PID: %d, Page: %s, Tuple: %s", pid, page.String, tuple.String)

			// Essayer de récupérer les données de la ligne si possible
			if page.String != "" && tuple.String != "" {
				rowQuery := fmt.Sprintf(`
					SELECT id, project_id, %s 
					FROM %s 
					WHERE ctid = ('(%s,%s)'::tid)
				`, getTableColumns(table), table, page.String, tuple.String)

				var id, projectID, otherField sql.NullString
				err := db.QueryRow(rowQuery).Scan(&id, &projectID, &otherField)
				if err == nil {
					fmt.Printf(" -> ID: %s, ProjectID: %s", id.String, projectID.String)
					if otherField.String != "" {
						fmt.Printf(", %s: %s", getOtherColumnName(table), otherField.String)
					}
				}
			}
			fmt.Println()
		}
		rows.Close()

		if !hasLocks {
			fmt.Printf("\nAucun lock de ligne sur la table %s\n", table)
		}
	}
}

func getTableColumns(table string) string {
	switch table {
	case "models":
		return "state"
	case "files":
		return "content"
	default:
		return "id"
	}
}

func getOtherColumnName(table string) string {
	switch table {
	case "models":
		return "State"
	case "files":
		return "Content"
	default:
		return "ID"
	}
}
