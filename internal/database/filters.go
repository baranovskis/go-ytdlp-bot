package database

import (
	"strings"
	"time"
)

type URLFilter struct {
	ID                 int64
	Hosts              []string
	ExcludeQueryParams bool
	PathRegex          string
	CookiesFile        string
	CreatedAt          time.Time
}

func (db *DB) InsertFilter(hosts []string, excludeQueryParams bool, pathRegex, cookiesFile string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO url_filters (hosts, exclude_query_params, path_regex, cookies_file) VALUES (?, ?, ?, ?)`,
		strings.Join(hosts, "\n"), boolToInt(excludeQueryParams), pathRegex, cookiesFile,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) UpdateFilter(id int64, hosts []string, excludeQueryParams bool, pathRegex, cookiesFile string) error {
	_, err := db.Exec(
		`UPDATE url_filters SET hosts = ?, exclude_query_params = ?, path_regex = ?, cookies_file = ? WHERE id = ?`,
		strings.Join(hosts, "\n"), boolToInt(excludeQueryParams), pathRegex, cookiesFile, id,
	)
	return err
}

func (db *DB) DeleteFilter(id int64) error {
	_, err := db.Exec(`DELETE FROM url_filters WHERE id = ?`, id)
	return err
}

func (db *DB) ListFilters() ([]URLFilter, error) {
	rows, err := db.Query(`SELECT id, hosts, exclude_query_params, path_regex, cookies_file, created_at FROM url_filters ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []URLFilter
	for rows.Next() {
		var f URLFilter
		var hostsStr string
		var excludeQP int
		if err := rows.Scan(&f.ID, &hostsStr, &excludeQP, &f.PathRegex, &f.CookiesFile, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Hosts = splitHosts(hostsStr)
		f.ExcludeQueryParams = excludeQP != 0
		filters = append(filters, f)
	}
	return filters, rows.Err()
}

func (db *DB) FilterCount() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM url_filters`).Scan(&count)
	return count, err
}

// SeedFilters inserts config filters into the DB only if the table is empty.
// If no config filters are provided, seeds with default popular platforms.
func (db *DB) SeedFilters(filters []struct {
	Hosts              []string
	ExcludeQueryParams bool
	PathRegex          string
	CookiesFile        string
}) error {
	count, err := db.FilterCount()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Use config filters if provided
	if len(filters) > 0 {
		for _, f := range filters {
			if _, err := db.InsertFilter(f.Hosts, f.ExcludeQueryParams, f.PathRegex, f.CookiesFile); err != nil {
				return err
			}
		}
		return nil
	}

	// Seed defaults for common video platforms
	defaults := []struct {
		hosts              []string
		excludeQueryParams bool
	}{
		{hosts: []string{"tiktok.com", "www.tiktok.com", "vm.tiktok.com"}, excludeQueryParams: true},
		{hosts: []string{"youtube.com", "www.youtube.com", "youtu.be", "m.youtube.com"}, excludeQueryParams: true},
		{hosts: []string{"instagram.com", "www.instagram.com"}, excludeQueryParams: true},
		{hosts: []string{"twitter.com", "x.com", "www.x.com"}, excludeQueryParams: true},
		{hosts: []string{"reddit.com", "www.reddit.com", "old.reddit.com"}, excludeQueryParams: true},
		{hosts: []string{"facebook.com", "www.facebook.com", "fb.watch"}, excludeQueryParams: true},
	}

	for _, d := range defaults {
		if _, err := db.InsertFilter(d.hosts, d.excludeQueryParams, "", ""); err != nil {
			return err
		}
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func splitHosts(s string) []string {
	if s == "" {
		return nil
	}
	var hosts []string
	for _, h := range strings.Split(s, "\n") {
		h = strings.TrimSpace(h)
		if h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}
