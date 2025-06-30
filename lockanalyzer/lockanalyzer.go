package lockanalyzer

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/uptrace/bun"
)

// ShowLocks affiche tous les locks actifs dans la base de données
func ShowLocks(db *bun.DB) {
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

// ExplainLocks explique en détail chaque type de lock ExclusiveLock
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

// ShowLockedRows affiche les détails des lignes lockées
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

// getTableColumns retourne le nom de la colonne spécifique à la table
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

// getOtherColumnName retourne le nom affiché de la colonne spécifique
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
