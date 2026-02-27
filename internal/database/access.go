package database

import "time"

type AllowedGroup struct {
	ChatID  int64
	Title   string
	Status  string
	AddedAt time.Time
}

type AllowedUser struct {
	UserID   int64
	Username string
	Status   string
	AddedAt  time.Time
}

// Pending groups

func (db *DB) AddPendingGroup(chatID int64, title string) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO allowed_groups (chat_id, title, status) VALUES (?, ?, 'pending')`,
		chatID, title,
	)
	return err
}

func (db *DB) ApprovePendingGroup(chatID int64) error {
	_, err := db.Exec(`UPDATE allowed_groups SET status = 'approved' WHERE chat_id = ?`, chatID)
	return err
}

func (db *DB) RejectPendingGroup(chatID int64) error {
	_, err := db.Exec(`UPDATE allowed_groups SET status = 'rejected' WHERE chat_id = ?`, chatID)
	return err
}

func (db *DB) ListPendingGroups() ([]AllowedGroup, error) {
	return db.listGroupsByStatus("pending")
}

// Allowed groups

func (db *DB) AddAllowedGroup(chatID int64, title string) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO allowed_groups (chat_id, title, status) VALUES (?, ?, 'approved')`,
		chatID, title,
	)
	return err
}

func (db *DB) RemoveAllowedGroup(chatID int64) error {
	_, err := db.Exec(`DELETE FROM allowed_groups WHERE chat_id = ?`, chatID)
	return err
}

func (db *DB) ListAllowedGroups() ([]AllowedGroup, error) {
	return db.listGroupsByStatus("approved")
}

func (db *DB) IsGroupAllowed(chatID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM allowed_groups WHERE chat_id = ? AND status = 'approved'`, chatID).Scan(&count)
	return count > 0, err
}

func (db *DB) IsGroupPending(chatID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM allowed_groups WHERE chat_id = ? AND status = 'pending'`, chatID).Scan(&count)
	return count > 0, err
}

func (db *DB) listGroupsByStatus(status string) ([]AllowedGroup, error) {
	rows, err := db.Query(`SELECT chat_id, title, status, added_at FROM allowed_groups WHERE status = ? ORDER BY added_at DESC`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []AllowedGroup
	for rows.Next() {
		var g AllowedGroup
		if err := rows.Scan(&g.ChatID, &g.Title, &g.Status, &g.AddedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// Pending users

func (db *DB) RegisterUser(userID int64, username string) error {
	_, err := db.Exec(
		`INSERT INTO allowed_users (user_id, username, status) VALUES (?, ?, 'pending')
		 ON CONFLICT(user_id) DO UPDATE SET username = excluded.username`,
		userID, username,
	)
	return err
}

func (db *DB) ApprovePendingUser(userID int64) error {
	_, err := db.Exec(`UPDATE allowed_users SET status = 'approved' WHERE user_id = ?`, userID)
	return err
}

func (db *DB) RejectPendingUser(userID int64) error {
	_, err := db.Exec(`UPDATE allowed_users SET status = 'rejected' WHERE user_id = ?`, userID)
	return err
}

func (db *DB) ListPendingUsers() ([]AllowedUser, error) {
	return db.listUsersByStatus("pending")
}

func (db *DB) IsUserPending(userID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM allowed_users WHERE user_id = ? AND status = 'pending'`, userID).Scan(&count)
	return count > 0, err
}

// Allowed users

func (db *DB) AddAllowedUser(userID int64, username string) error {
	_, err := db.Exec(
		`INSERT INTO allowed_users (user_id, username, status) VALUES (?, ?, 'approved')
		 ON CONFLICT(user_id) DO UPDATE SET username = excluded.username, status = 'approved'`,
		userID, username,
	)
	return err
}

func (db *DB) RemoveAllowedUser(userID int64) error {
	_, err := db.Exec(`DELETE FROM allowed_users WHERE user_id = ?`, userID)
	return err
}

func (db *DB) ListAllowedUsers() ([]AllowedUser, error) {
	return db.listUsersByStatus("approved")
}

func (db *DB) IsUserAllowed(userID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM allowed_users WHERE user_id = ? AND status = 'approved'`, userID).Scan(&count)
	return count > 0, err
}

func (db *DB) listUsersByStatus(status string) ([]AllowedUser, error) {
	rows, err := db.Query(`SELECT user_id, username, status, added_at FROM allowed_users WHERE status = ? ORDER BY added_at DESC`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AllowedUser
	for rows.Next() {
		var u AllowedUser
		if err := rows.Scan(&u.UserID, &u.Username, &u.Status, &u.AddedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
