package database

import "time"

type LogEntry struct {
	ID        int64
	Level     string
	Message   string
	Fields    string
	CreatedAt time.Time
}

type LogFilter struct {
	Level  string
	Search string
	Limit  int
	Offset int
}

func (db *DB) InsertLog(level, message, fields string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO logs (level, message, fields) VALUES (?, ?, ?)`,
		level, message, fields,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) ListLogs(f LogFilter) ([]LogEntry, int, error) {
	if f.Limit == 0 {
		f.Limit = 100
	}

	countQuery := "SELECT COUNT(*) FROM logs WHERE 1=1"
	query := "SELECT id, level, message, fields, created_at FROM logs WHERE 1=1"
	var args []any

	if f.Level != "" {
		countQuery += " AND level = ?"
		query += " AND level = ?"
		args = append(args, f.Level)
	}
	if f.Search != "" {
		countQuery += " AND (message LIKE ? OR fields LIKE ?)"
		query += " AND (message LIKE ? OR fields LIKE ?)"
		s := "%" + f.Search + "%"
		args = append(args, s, s)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.Level, &l.Message, &l.Fields, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, total, rows.Err()
}

func (db *DB) ListLogsSince(id int64) ([]LogEntry, error) {
	rows, err := db.Query(
		"SELECT id, level, message, fields, created_at FROM logs WHERE id > ? ORDER BY id ASC",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.Level, &l.Message, &l.Fields, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}
