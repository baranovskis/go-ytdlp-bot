package database

import (
	"database/sql"
	"fmt"
)

var migrations = []string{
	// Migration 1: Initial schema
	`CREATE TABLE IF NOT EXISTS downloads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		telegram_user_id INTEGER NOT NULL,
		telegram_username TEXT NOT NULL DEFAULT '',
		chat_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		filename TEXT NOT NULL DEFAULT '',
		error_message TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_downloads_created_at ON downloads(created_at);
	CREATE INDEX IF NOT EXISTS idx_downloads_status ON downloads(status);
	CREATE INDEX IF NOT EXISTS idx_downloads_telegram_user_id ON downloads(telegram_user_id);

	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		fields TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);

	CREATE TABLE IF NOT EXISTS allowed_groups (
		chat_id INTEGER PRIMARY KEY,
		title TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		added_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS allowed_users (
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL DEFAULT '',
		added_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		expires_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);`,

	// Migration 2: URL filters table
	`CREATE TABLE IF NOT EXISTS url_filters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hosts TEXT NOT NULL,
		exclude_query_params INTEGER NOT NULL DEFAULT 0,
		path_regex TEXT NOT NULL DEFAULT '',
		cookies_file TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);`,

	// Migration 3: Add status column to allowed_users (pending/approved/rejected)
	`ALTER TABLE allowed_users ADD COLUMN status TEXT NOT NULL DEFAULT 'approved';`,
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	var current int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("get current migration version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", i+1, err)
		}

		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %d: %w", i+1, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", i+1); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", i+1, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", i+1, err)
		}
	}

	return nil
}
