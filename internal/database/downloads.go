package database

import "time"

type Download struct {
	ID               int64
	URL              string
	TelegramUserID   int64
	TelegramUsername string
	ChatID           int64
	Status           string
	Filename         string
	ErrorMessage     string
	CreatedAt        time.Time
}

type DownloadFilter struct {
	Status string
	UserID int64
	Limit  int
	Offset int
}

func (db *DB) InsertDownload(url string, telegramUserID int64, telegramUsername string, chatID int64, status, filename, errorMessage string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO downloads (url, telegram_user_id, telegram_username, chat_id, status, filename, error_message) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		url, telegramUserID, telegramUsername, chatID, status, filename, errorMessage,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) UpdateDownloadStatus(id int64, status, filename, errorMessage string) error {
	_, err := db.Exec(
		`UPDATE downloads SET status = ?, filename = ?, error_message = ? WHERE id = ?`,
		status, filename, errorMessage, id,
	)
	return err
}

func (db *DB) ListDownloads(f DownloadFilter) ([]Download, int, error) {
	if f.Limit == 0 {
		f.Limit = 50
	}

	countQuery := "SELECT COUNT(*) FROM downloads WHERE 1=1"
	query := "SELECT id, url, telegram_user_id, telegram_username, chat_id, status, filename, error_message, created_at FROM downloads WHERE 1=1"
	var args []any

	if f.Status != "" {
		countQuery += " AND status = ?"
		query += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.UserID != 0 {
		countQuery += " AND telegram_user_id = ?"
		query += " AND telegram_user_id = ?"
		args = append(args, f.UserID)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		if err := rows.Scan(&d.ID, &d.URL, &d.TelegramUserID, &d.TelegramUsername, &d.ChatID, &d.Status, &d.Filename, &d.ErrorMessage, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		downloads = append(downloads, d)
	}

	return downloads, total, rows.Err()
}

type DownloadStats struct {
	Total     int
	Succeeded int
	Failed    int
}

func (db *DB) GetDownloadStats() (DownloadStats, error) {
	var s DownloadStats
	err := db.QueryRow(`SELECT
		COUNT(*),
		COUNT(CASE WHEN status = 'success' THEN 1 END),
		COUNT(CASE WHEN status = 'failed' THEN 1 END)
		FROM downloads`).Scan(&s.Total, &s.Succeeded, &s.Failed)
	return s, err
}
