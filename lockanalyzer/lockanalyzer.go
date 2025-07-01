package lockanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

// LockAnalysisData contient toutes les données d'analyse des locks
type LockAnalysisData struct {
	Locks           []LockInfo
	RowLocks        []RowLockInfo
	Deadlocks       []DeadlockInfo
	BlockedTxns     []BlockedTransaction
	LongTxns        []LongTransaction
	ObjectConflicts []ObjectConflict
	IndexAnalysis   []IndexInfo
}

// DeadlockInfo contient les informations sur un deadlock potentiel
type DeadlockInfo struct {
	Transaction1   LockInfo
	Transaction2   LockInfo
	ConflictType   string
	Recommendation string
}

// ShowLocks affiche tous les locks actifs dans la base de données
// SUPPRESSION de ShowLocks et des fonctions d'affichage direct console qui ne sont plus utiles

// AnalyzeDeadlocks détecte les deadlocks potentiels
func analyzeDeadlocks(db *bun.DB) {
	query := `
		SELECT 
			l1.pid as pid1,
			l1.mode as mode1,
			l1.granted as granted1,
			t1.relname as table1,
			l2.pid as pid2,
			l2.mode as mode2,
			l2.granted as granted2,
			t2.relname as table2
		FROM pg_locks l1
		JOIN pg_locks l2 ON l1.relation = l2.relation AND l1.pid != l2.pid
		LEFT JOIN pg_class t1 ON l1.relation = t1.oid
		LEFT JOIN pg_class t2 ON l2.relation = t2.oid
		WHERE l1.pid != pg_backend_pid()
		AND l2.pid != pg_backend_pid()
		AND l1.granted = true
		AND l2.granted = false
		AND l1.relation IS NOT NULL
		AND l2.relation IS NOT NULL
		ORDER BY l1.pid, l2.pid;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'analyse des deadlocks: %v", err)
		return
	}
	defer rows.Close()

	deadlockCount := 0
	for rows.Next() {
		var pid1, pid2 int
		var mode1, granted1, table1, mode2, granted2, table2 sql.NullString

		rows.Scan(&pid1, &mode1, &granted1, &table1, &pid2, &mode2, &granted2, &table2)

		deadlockCount++
		fmt.Printf("Deadlock potentiel #%d:\n", deadlockCount)
		fmt.Printf("  PID %d (%s sur %s) bloque PID %d (%s sur %s)\n",
			pid1, mode1.String, table1.String, pid2, mode2.String, table2.String)

		// Analyser le type de conflit
		conflictType := analyzeConflictType(mode1.String, mode2.String)
		fmt.Printf("  Type de conflit: %s\n", conflictType)

		// Suggérer une solution
		suggestion := suggestDeadlockSolution(pid1, pid2, table1.String, table2.String, conflictType)
		fmt.Printf("  Suggestion: %s\n\n", suggestion)
	}

	if deadlockCount == 0 {
		fmt.Println("Aucun deadlock détecté")
	}
}

// AnalyzeBlockedTransactions détecte les transactions bloquées en temps réel
func analyzeBlockedTransactions(db *bun.DB) {
	query := `
		SELECT 
			waiter.pid as waiter_pid,
			waiter.mode as waiter_mode,
			waiter.granted as waiter_granted,
			waiter_table.relname as waiter_table,
			blocker.pid as blocker_pid,
			blocker.mode as blocker_mode,
			blocker.granted as blocker_granted,
			blocker_table.relname as blocker_table,
			EXTRACT(EPOCH FROM (now() - sa.query_start)) as wait_duration
		FROM pg_locks waiter
		JOIN pg_locks blocker ON waiter.relation = blocker.relation 
			AND waiter.pid != blocker.pid
			AND waiter.mode != blocker.mode
		LEFT JOIN pg_class waiter_table ON waiter.relation = waiter_table.oid
		LEFT JOIN pg_class blocker_table ON blocker.relation = blocker_table.oid
		LEFT JOIN pg_stat_activity sa ON waiter.pid = sa.pid
		WHERE waiter.pid != pg_backend_pid()
		AND blocker.pid != pg_backend_pid()
		AND waiter.granted = false
		AND blocker.granted = true
		AND waiter.relation IS NOT NULL
		AND blocker.relation IS NOT NULL
		ORDER BY wait_duration DESC;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'analyse des transactions bloquées: %v", err)
		return
	}
	defer rows.Close()

	blockedCount := 0
	for rows.Next() {
		var waiterPID, blockerPID int
		var waiterMode, waiterGranted, waiterTable, blockerMode, blockerGranted, blockerTable sql.NullString
		var waitDuration sql.NullFloat64

		rows.Scan(&waiterPID, &waiterMode, &waiterGranted, &waiterTable, &blockerPID, &blockerMode, &blockerGranted, &blockerTable, &waitDuration)

		blockedCount++
		fmt.Printf("Transaction bloquée #%d:\n", blockedCount)
		fmt.Printf("  PID %d (%s sur %s) attend depuis %.2f secondes\n",
			waiterPID, waiterMode.String, waiterTable.String, waitDuration.Float64)
		fmt.Printf("  Bloqué par PID %d (%s sur %s)\n",
			blockerPID, blockerMode.String, blockerTable.String)

		// Analyser le type de conflit
		conflictType := analyzeConflictType(waiterMode.String, blockerMode.String)
		fmt.Printf("  Type de conflit: %s\n", conflictType)

		// Suggérer une solution
		suggestion := suggestBlockedTransactionSolution(waiterPID, blockerPID, waiterTable.String, blockerTable.String, conflictType)
		fmt.Printf("  Suggestion: %s\n\n", suggestion)
	}

	if blockedCount == 0 {
		fmt.Println("Aucune transaction bloquée détectée")
	}
}

// SuggestCorrections propose des corrections pour les locks problématiques
func suggestCorrections(db *bun.DB) {
	// Analyser les transactions longues
	analyzeLongTransactions(db)

	// Analyser les locks sur les mêmes objets
	analyzeObjectConflicts(db)

	// Suggérer des optimisations d'index
	suggestIndexOptimizations(db)
}

// AnalyzeLongTransactions identifie les transactions longues
func analyzeLongTransactions(db *bun.DB) {
	query := `
		SELECT 
			pid,
			now() - query_start as duration,
			query
		FROM pg_stat_activity 
		WHERE state = 'active'
		AND pid != pg_backend_pid()
		AND now() - query_start > interval '5 seconds'
		ORDER BY duration DESC;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'analyse des transactions longues: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Transactions longues détectées:")
	longTransactionCount := 0
	for rows.Next() {
		var pid int
		var duration, query sql.NullString
		rows.Scan(&pid, &duration, &query)

		longTransactionCount++
		fmt.Printf("  PID %d: %s\n", pid, duration.String)
		fmt.Printf("    Query: %s\n", truncateQuery(query.String, 100))
		fmt.Printf("    Suggestion: Vérifier si la transaction peut être divisée\n\n")
	}

	if longTransactionCount == 0 {
		fmt.Println("Aucune transaction longue détectée")
	}
}

// AnalyzeObjectConflicts analyse les conflits sur les mêmes objets
func analyzeObjectConflicts(db *bun.DB) {
	query := `
		SELECT 
			t.relname as table_name,
			l.mode,
			count(*) as lock_count,
			array_agg(l.pid) as pids
		FROM pg_locks l
		JOIN pg_class t ON l.relation = t.oid
		WHERE l.pid != pg_backend_pid()
		AND l.relation IS NOT NULL
		GROUP BY t.relname, l.mode
		HAVING count(*) > 1
		ORDER BY lock_count DESC;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'analyse des conflits d'objets: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Conflits d'objets détectés:")
	conflictCount := 0
	for rows.Next() {
		var tableName, mode sql.NullString
		var lockCount int
		var pids sql.NullString
		rows.Scan(&tableName, &mode, &lockCount, &pids)

		conflictCount++
		fmt.Printf("  Table %s: %d locks en mode %s (PIDs: %s)\n",
			tableName.String, lockCount, mode.String, pids.String)
		fmt.Printf("    Suggestion: Vérifier l'ordre des opérations\n\n")
	}

	if conflictCount == 0 {
		fmt.Println("Aucun conflit d'objet détecté")
	}
}

// SuggestIndexOptimizations suggère des optimisations d'index
func suggestIndexOptimizations(db *bun.DB) {
	query := `
		SELECT 
			t.relname as table_name,
			i.relname as index_name,
			pg_size_pretty(pg_relation_size(i.oid)) as index_size
		FROM pg_index x
		JOIN pg_class t ON x.indrelid = t.oid
		JOIN pg_class i ON x.indexrelid = i.oid
		WHERE t.relname IN ('models', 'files', 'projects')
		ORDER BY pg_relation_size(i.oid) DESC;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Erreur lors de l'analyse des index: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Analyse des index:")
	for rows.Next() {
		var tableName, indexName, indexSize sql.NullString
		rows.Scan(&tableName, &indexName, &indexSize)

		fmt.Printf("  %s.%s: %s\n", tableName.String, indexName.String, indexSize.String)
	}
	fmt.Println("  Suggestion: Vérifier que les index sont utilisés efficacement")
}

// AnalyzeConflictType détermine le type de conflit entre deux modes de lock
func analyzeConflictType(mode1, mode2 string) string {
	if mode1 == "ExclusiveLock" && mode2 == "ExclusiveLock" {
		return "Conflit exclusif-exclusif"
	}
	if (mode1 == "RowExclusiveLock" && mode2 == "ExclusiveLock") ||
		(mode1 == "ExclusiveLock" && mode2 == "RowExclusiveLock") {
		return "Conflit exclusif-row_exclusive"
	}
	if mode1 == "ShareLock" && mode2 == "ExclusiveLock" {
		return "Conflit partage-exclusif"
	}
	return "Conflit de modes"
}

// SuggestDeadlockSolution suggère une solution pour un deadlock
func suggestDeadlockSolution(pid1, pid2 int, table1, table2, conflictType string) string {
	if table1 == table2 {
		return fmt.Sprintf("Standardiser l'ordre des opérations sur %s", table1)
	}
	if conflictType == "Conflit exclusif-exclusif" {
		return "Utiliser des transactions plus courtes ou des locks explicites"
	}
	return "Revoir l'architecture des transactions pour éviter les dépendances circulaires"
}

// SuggestBlockedTransactionSolution suggère une solution pour une transaction bloquée
func suggestBlockedTransactionSolution(waiterPID, blockerPID int, waiterTable, blockerTable, conflictType string) string {
	if waiterTable == blockerTable {
		return fmt.Sprintf("Standardiser l'ordre des opérations sur %s", waiterTable)
	}
	if conflictType == "Conflit exclusif-exclusif" {
		return "Utiliser des transactions plus courtes ou des locks explicites"
	}
	if conflictType == "Conflit exclusif-row_exclusive" {
		return "Vérifier si les opérations peuvent être réorganisées"
	}
	return "Revoir l'architecture des transactions pour éviter les dépendances circulaires"
}

// TruncateQuery tronque une requête pour l'affichage
func truncateQuery(query string, maxLength int) string {
	if len(query) <= maxLength {
		return query
	}
	return query[:maxLength] + "..."
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

// SuggestSolutions suggère des solutions adaptées aux problèmes détectés
func SuggestSolutions(data *LockAnalysisData) []string {
	var suggestions []string

	// Suggestions basées sur les transactions longues
	if len(data.LongTxns) > 0 {
		suggestions = append(suggestions, "Diviser les transactions longues en transactions plus petites")
		suggestions = append(suggestions, "Utiliser des niveaux d'isolation plus faibles si possible")
		suggestions = append(suggestions, "Implémenter des timeouts appropriés pour éviter les blocages prolongés")
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
