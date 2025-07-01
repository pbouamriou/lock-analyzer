package lockanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/uptrace/bun"
)

// LockInfo contains detailed information about a lock
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

// RowLockInfo contains information about row locks
type RowLockInfo struct {
	PID     int
	Table   string
	Page    string
	Tuple   string
	Mode    string
	Granted bool
}

// BlockedTransaction contains information about a blocked transaction
type BlockedTransaction struct {
	PID       string
	Duration  string
	Query     string
	WaitEvent string
}

// LongTransaction contains information about a long transaction
type LongTransaction struct {
	PID      string
	Duration string
	Query    string
}

// ObjectConflict contains information about an object conflict
type ObjectConflict struct {
	Object         string
	PIDs           []string
	Mode           string
	Recommendation string
}

// IndexInfo contains information about an index
type IndexInfo struct {
	Name  string
	Table string
	Size  string
	Usage string
}

// DeadlockInfo contains information about a potential deadlock
type DeadlockInfo struct {
	Transaction1   LockInfo
	Transaction2   LockInfo
	ConflictType   string
	Recommendation string
}

// ReportData contains all data needed to generate a report
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

// ReportSummary contains a summary of detected issues
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

// LockReportFormatter defines the interface for report formatters
type LockReportFormatter interface {
	Format(data *ReportData, output io.Writer) error
	GetFileExtension() string
}

// DetectBlockedTransactions detects blocked transactions in real-time
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
		return []string{fmt.Sprintf("Error detecting blocked transactions: %v", err)}
	}
	defer rows.Close()

	for rows.Next() {
		var pid int
		var duration, queryText, waitEventType, waitEvent string

		if err := rows.Scan(&pid, &duration, &queryText, &waitEventType, &waitEvent); err != nil {
			continue
		}

		blockedQueries = append(blockedQueries, fmt.Sprintf("PID %d blocked for %s: %s (event: %s - %s)",
			pid, duration, queryText, waitEventType, waitEvent))
	}

	return blockedQueries
}

// GenerateLocksReport generates a complete locks report and returns the data
func GenerateLocksReport(db *bun.DB) (*ReportData, error) {
	// Collect all data
	data := &ReportData{
		Timestamp: time.Now(),
	}

	// Retrieve locks
	locks, err := getLocks(db)
	if err != nil {
		return nil, fmt.Errorf("error retrieving locks: %v", err)
	}
	data.Locks = locks

	// Retrieve row locks
	rowLocks, err := getRowLocks(db)
	if err != nil {
		return nil, fmt.Errorf("error retrieving row locks: %v", err)
	}
	data.RowLocks = rowLocks

	// Analyze deadlocks
	deadlocks := detectDeadlocks(locks)
	data.Deadlocks = deadlocks

	// Analyze blocked transactions
	blockedTxns := detectBlockedTransactions(locks)
	data.BlockedTxns = blockedTxns

	// Analyze long transactions
	longTxns := detectLongTransactions(db)
	data.LongTxns = longTxns

	// Analyze object conflicts
	objectConflicts := detectObjectConflicts(locks)
	data.ObjectConflicts = objectConflicts

	// Analyze indexes
	indexAnalysis := analyzeIndexes(db)
	data.IndexAnalysis = indexAnalysis

	// Generate suggestions
	suggestions := generateSuggestions(data)
	data.Suggestions = suggestions

	// Calculate summary
	data.Summary = calculateSummary(data)

	return data, nil
}

// calculateSummary calculates the summary of detected issues
func calculateSummary(data *ReportData) ReportSummary {
	summary := ReportSummary{
		TotalLocks:      len(data.Locks),
		BlockedTxns:     len(data.BlockedTxns),
		LongTxns:        len(data.LongTxns),
		Deadlocks:       len(data.Deadlocks),
		ObjectConflicts: len(data.ObjectConflicts),
		Recommendations: len(data.Suggestions),
	}

	// Calculate critical issues
	summary.CriticalIssues = summary.Deadlocks + summary.BlockedTxns

	// Calculate warnings
	summary.Warnings = summary.LongTxns + summary.ObjectConflicts

	return summary
}

// generateSuggestions generates suggestions based on analysis
func generateSuggestions(data *ReportData) []string {
	var suggestions []string

	// Suggestions based on blocked transactions
	if len(data.BlockedTxns) > 0 {
		suggestions = append(suggestions, "Consider adding timeouts on long transactions")
		suggestions = append(suggestions, "Check lock acquisition order to avoid deadlocks")
	}

	// Suggestions based on long transactions
	if len(data.LongTxns) > 0 {
		suggestions = append(suggestions, "Split long transactions into smaller ones")
		suggestions = append(suggestions, "Optimize queries to reduce execution time")
	}

	// Suggestions based on object conflicts
	if len(data.ObjectConflicts) > 0 {
		suggestions = append(suggestions, "Review lock acquisition strategy")
		suggestions = append(suggestions, "Consider using lower isolation levels if appropriate")
	}

	// Suggestions based on deadlocks
	if len(data.Deadlocks) > 0 {
		suggestions = append(suggestions, "Implement retry logic with exponential backoff")
		suggestions = append(suggestions, "Standardize table access order to avoid deadlocks")
	}

	// General suggestions
	if len(data.Locks) > 10 {
		suggestions = append(suggestions, "Consider reviewing transaction patterns")
		suggestions = append(suggestions, "Monitor lock wait times regularly")
	}

	return suggestions
}

// getLocks retrieves all active locks
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

// getRowLocks retrieves row locks
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

// detectDeadlocks detects potential deadlocks
func detectDeadlocks(locks []LockInfo) []DeadlockInfo {
	var deadlocks []DeadlockInfo

	// Simplified logic to detect deadlocks
	// In practice, a more complex analysis would be needed
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

// detectBlockedTransactions detects blocked transactions
func detectBlockedTransactions(locks []LockInfo) []BlockedTransaction {
	var blocked []BlockedTransaction

	// Simplified logic to detect blocked transactions
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

// detectLongTransactions detects long transactions
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

// detectObjectConflicts detects object conflicts
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

// analyzeIndexes analyzes indexes
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
