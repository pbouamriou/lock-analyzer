package lockanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// LockInfo contient les informations détaillées d'un lock
type LockInfo struct {
	PID           int
	Mode          string
	Granted       bool
	ObjectType    string
	ObjectName    string
	Page          string
	Tuple         string
	VirtualXID    string
	TransactionID string
	WaitTime      time.Duration
	Query         string
	Type          string
	Object        string
}

// RowLockInfo contient les informations sur les locks de lignes
type RowLockInfo struct {
	PID     int
	Table   string
	Page    string
	Tuple   string
	Mode    string
	Granted bool
}

// BlockedTransaction contient les informations sur une transaction bloquée
type BlockedTransaction struct {
	PID       string
	Duration  string
	Query     string
	WaitEvent string
}

// LongTransaction contient les informations sur une transaction longue
type LongTransaction struct {
	PID      string
	Duration string
	Query    string
}

// ObjectConflict contient les informations sur un conflit d'objet
type ObjectConflict struct {
	Object         string
	PIDs           []string
	Mode           string
	Recommendation string
}

// IndexInfo contient les informations sur un index
type IndexInfo struct {
	Name  string
	Table string
	Size  string
	Usage string
}

// DeadlockInfo contient les informations sur un deadlock potentiel
type DeadlockInfo struct {
	Transaction1   LockInfo
	Transaction2   LockInfo
	ConflictType   string
	Recommendation string
}

// ReportData contient toutes les données pour générer un rapport
type ReportData struct {
	Timestamp       time.Time
	Locks           []LockInfo
	RowLocks        []RowLockInfo
	Deadlocks       []DeadlockInfo
	BlockedTxns     []BlockedTransaction
	LongTxns        []LongTransaction
	ObjectConflicts []ObjectConflict
	IndexAnalysis   []IndexInfo
	Suggestions     []string
	Summary         ReportSummary
}

// ReportSummary contient un résumé des problèmes détectés
type ReportSummary struct {
	TotalLocks      int
	BlockedTxns     int
	LongTxns        int
	Deadlocks       int
	ObjectConflicts int
	CriticalIssues  int
	Warnings        int
	Recommendations int
}

// DetectBlockedTransactions détecte les transactions bloquées en temps réel
func DetectBlockedTransactions(db *bun.DB) []string {
	var blockedQueries []string

	query := `
		SELECT 
			pid,
			now() - query_start AS duration,
			query,
			wait_event_type,
			wait_event
		FROM pg_stat_activity 
		WHERE state = 'active' 
		AND wait_event_type IS NOT NULL
		AND pid != pg_backend_pid()
		ORDER BY duration DESC
	`

	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return []string{fmt.Sprintf("Erreur lors de la détection des transactions bloquées: %v", err)}
	}
	defer rows.Close()

	for rows.Next() {
		var pid int
		var duration, queryText, waitEventType, waitEvent string

		if err := rows.Scan(&pid, &duration, &queryText, &waitEventType, &waitEvent); err != nil {
			continue
		}

		blockedQueries = append(blockedQueries, fmt.Sprintf("PID %d bloqué depuis %s: %s (événement: %s - %s)",
			pid, duration, queryText, waitEventType, waitEvent))
	}

	return blockedQueries
}

// GenerateLocksReport génère un rapport complet des locks et retourne les données
func GenerateLocksReport(db *bun.DB) (*ReportData, error) {
	// Collecter toutes les données
	data := &ReportData{
		Timestamp: time.Now(),
	}

	// Récupérer les locks
	locks, err := getLocks(db)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des locks: %v", err)
	}
	data.Locks = locks

	// Récupérer les locks de lignes
	rowLocks, err := getRowLocks(db)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des locks de lignes: %v", err)
	}
	data.RowLocks = rowLocks

	// Analyser les deadlocks
	deadlocks := detectDeadlocks(locks)
	data.Deadlocks = deadlocks

	// Analyser les transactions bloquées
	blockedTxns := detectBlockedTransactions(locks)
	data.BlockedTxns = blockedTxns

	// Analyser les transactions longues
	longTxns := detectLongTransactions(db)
	data.LongTxns = longTxns

	// Analyser les conflits d'objets
	objectConflicts := detectObjectConflicts(locks)
	data.ObjectConflicts = objectConflicts

	// Analyser les index
	indexAnalysis := analyzeIndexes(db)
	data.IndexAnalysis = indexAnalysis

	// Générer les suggestions
	suggestions := generateSuggestions(data)
	data.Suggestions = suggestions

	// Calculer le résumé
	data.Summary = calculateSummary(data)

	return data, nil
}

// calculateSummary calcule le résumé des problèmes détectés
func calculateSummary(data *ReportData) ReportSummary {
	summary := ReportSummary{
		TotalLocks:      len(data.Locks),
		BlockedTxns:     len(data.BlockedTxns),
		LongTxns:        len(data.LongTxns),
		Deadlocks:       len(data.Deadlocks),
		ObjectConflicts: len(data.ObjectConflicts),
		Recommendations: len(data.Suggestions),
	}

	// Calculer les problèmes critiques
	summary.CriticalIssues = summary.Deadlocks + summary.BlockedTxns

	// Calculer les avertissements
	summary.Warnings = summary.LongTxns + summary.ObjectConflicts

	return summary
}

// generateSuggestions génère des suggestions basées sur l'analyse
func generateSuggestions(data *ReportData) []string {
	var suggestions []string

	// Suggestions basées sur les transactions bloquées
	if len(data.BlockedTxns) > 0 {
		suggestions = append(suggestions, "Considérer l'ajout de timeouts sur les transactions longues")
		suggestions = append(suggestions, "Vérifier l'ordre d'acquisition des locks pour éviter les blocages")
	}

	// Suggestions basées sur les transactions longues
	if len(data.LongTxns) > 0 {
		suggestions = append(suggestions, "Diviser les transactions longues en transactions plus petites")
		suggestions = append(suggestions, "Optimiser les requêtes pour réduire le temps d'exécution")
	}

	// Suggestions basées sur les conflits d'objets
	if len(data.ObjectConflicts) > 0 {
		suggestions = append(suggestions, "Réviser la stratégie d'acquisition des locks")
		suggestions = append(suggestions, "Considérer l'utilisation de niveaux d'isolation plus faibles si approprié")
	}

	// Suggestions basées sur les deadlocks
	if len(data.Deadlocks) > 0 {
		suggestions = append(suggestions, "Implémenter une logique de retry avec backoff exponentiel")
		suggestions = append(suggestions, "Standardiser l'ordre d'accès aux tables pour éviter les deadlocks")
	}

	// Suggestions générales
	if len(data.Locks) > 10 {
		suggestions = append(suggestions, "Considérer l'optimisation des index pour réduire les locks")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Aucun problème critique détecté. Continuer à surveiller les performances")
	}

	return suggestions
}

// getLocks récupère tous les locks actifs
func getLocks(db *bun.DB) ([]LockInfo, error) {
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
				ELSE 'unknown'
			END as object_type,
			CASE 
				WHEN l.relation IS NOT NULL THEN t.relname
				ELSE 'N/A'
			END as object_name,
			l.page,
			l.tuple,
			l.virtualxid,
			l.transactionid
		FROM pg_locks l
		LEFT JOIN pg_class t ON l.relation = t.oid
		WHERE l.pid != pg_backend_pid()
		ORDER BY l.pid, l.mode;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []LockInfo
	for rows.Next() {
		var lock LockInfo
		var page, tuple, virtualxid, transactionid sql.NullString

		err := rows.Scan(&lock.PID, &lock.Mode, &lock.Granted, &lock.ObjectType, &lock.ObjectName,
			&page, &tuple, &virtualxid, &transactionid)
		if err != nil {
			continue
		}

		if page.Valid {
			lock.Page = page.String
		}
		if tuple.Valid {
			lock.Tuple = tuple.String
		}
		if virtualxid.Valid {
			lock.VirtualXID = virtualxid.String
		}
		if transactionid.Valid {
			lock.TransactionID = transactionid.String
		}

		lock.Type = lock.ObjectType
		lock.Object = lock.ObjectName

		locks = append(locks, lock)
	}

	return locks, nil
}

// getRowLocks récupère les locks de lignes
func getRowLocks(db *bun.DB) ([]RowLockInfo, error) {
	query := `
		SELECT 
			l.pid,
			t.relname,
			l.page,
			l.tuple,
			l.mode,
			l.granted
		FROM pg_locks l
		LEFT JOIN pg_class t ON l.relation = t.oid
		WHERE l.pid != pg_backend_pid()
		AND l.page IS NOT NULL
		ORDER BY l.pid;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rowLocks []RowLockInfo
	for rows.Next() {
		var rowLock RowLockInfo
		var table sql.NullString

		err := rows.Scan(&rowLock.PID, &table, &rowLock.Page, &rowLock.Tuple, &rowLock.Mode, &rowLock.Granted)
		if err != nil {
			continue
		}

		if table.Valid {
			rowLock.Table = table.String
		}

		rowLocks = append(rowLocks, rowLock)
	}

	return rowLocks, nil
}

// detectDeadlocks détecte les deadlocks potentiels
func detectDeadlocks(locks []LockInfo) []DeadlockInfo {
	var deadlocks []DeadlockInfo

	// Logique simplifiée pour détecter les deadlocks
	// En pratique, il faudrait une analyse plus complexe
	for i, lock1 := range locks {
		for j, lock2 := range locks {
			if i >= j {
				continue
			}

			if lock1.PID != lock2.PID &&
				lock1.Object == lock2.Object &&
				lock1.Granted != lock2.Granted {

				deadlock := DeadlockInfo{
					Transaction1:   lock1,
					Transaction2:   lock2,
					ConflictType:   "Lock conflict",
					Recommendation: "Review transaction order",
				}
				deadlocks = append(deadlocks, deadlock)
			}
		}
	}

	return deadlocks
}

// detectBlockedTransactions détecte les transactions bloquées
func detectBlockedTransactions(locks []LockInfo) []BlockedTransaction {
	var blocked []BlockedTransaction

	// Logique simplifiée pour détecter les transactions bloquées
	for _, lock := range locks {
		if !lock.Granted {
			blocked = append(blocked, BlockedTransaction{
				PID:       fmt.Sprintf("%d", lock.PID),
				Duration:  "unknown",
				Query:     lock.Query,
				WaitEvent: "lock",
			})
		}
	}

	return blocked
}

// detectLongTransactions détecte les transactions longues
func detectLongTransactions(db *bun.DB) []LongTransaction {
	query := `
		SELECT 
			pid,
			now() - query_start AS duration,
			query
		FROM pg_stat_activity 
		WHERE state = 'active' 
		AND pid != pg_backend_pid()
		AND now() - query_start > interval '5 seconds'
		ORDER BY duration DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var longTxns []LongTransaction
	for rows.Next() {
		var txn LongTransaction
		var duration, query sql.NullString

		err := rows.Scan(&txn.PID, &duration, &query)
		if err != nil {
			continue
		}

		if duration.Valid {
			txn.Duration = duration.String
		}
		if query.Valid {
			txn.Query = query.String
		}

		longTxns = append(longTxns, txn)
	}

	return longTxns
}

// detectObjectConflicts détecte les conflits d'objets
func detectObjectConflicts(locks []LockInfo) []ObjectConflict {
	objectMap := make(map[string][]string)

	for _, lock := range locks {
		if lock.Object != "" {
			objectMap[lock.Object] = append(objectMap[lock.Object], fmt.Sprintf("%d", lock.PID))
		}
	}

	var conflicts []ObjectConflict
	for object, pids := range objectMap {
		if len(pids) > 1 {
			conflicts = append(conflicts, ObjectConflict{
				Object:         object,
				PIDs:           pids,
				Mode:           "multiple",
				Recommendation: "Review access patterns",
			})
		}
	}

	return conflicts
}

// analyzeIndexes analyse les index
func analyzeIndexes(db *bun.DB) []IndexInfo {
	query := `
		SELECT 
			indexname,
			tablename,
			pg_size_pretty(pg_relation_size(indexname::regclass)) as size
		FROM pg_indexes 
		WHERE schemaname = 'public'
		ORDER BY pg_relation_size(indexname::regclass) DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		var size sql.NullString

		err := rows.Scan(&index.Name, &index.Table, &size)
		if err != nil {
			continue
		}

		if size.Valid {
			index.Size = size.String
		}

		indexes = append(indexes, index)
	}

	return indexes
}
