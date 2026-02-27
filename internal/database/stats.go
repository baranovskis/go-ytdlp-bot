package database

type DailyCount struct {
	Date  string
	Count int
}

type TopDomain struct {
	Domain string
	Count  int
}

type Stats struct {
	TotalDownloads int
	Succeeded      int
	Failed         int
	ActiveUsers    int
	DailyCounts    []DailyCount
	TopDomains     []TopDomain
}

func (db *DB) GetStats() (Stats, error) {
	var s Stats

	err := db.QueryRow(`SELECT
		COUNT(*),
		COUNT(CASE WHEN status = 'success' THEN 1 END),
		COUNT(CASE WHEN status = 'failed' THEN 1 END)
		FROM downloads`).Scan(&s.TotalDownloads, &s.Succeeded, &s.Failed)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`SELECT COUNT(DISTINCT telegram_user_id) FROM downloads`).Scan(&s.ActiveUsers)
	if err != nil {
		return s, err
	}

	rows, err := db.Query(`SELECT date(created_at) as day, COUNT(*) as cnt
		FROM downloads
		WHERE created_at >= datetime('now', '-30 days')
		GROUP BY day ORDER BY day DESC`)
	if err != nil {
		return s, err
	}
	defer rows.Close()

	for rows.Next() {
		var dc DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return s, err
		}
		s.DailyCounts = append(s.DailyCounts, dc)
	}
	if err := rows.Err(); err != nil {
		return s, err
	}

	rows2, err := db.Query(`SELECT
		CASE
			WHEN instr(replace(replace(url, 'https://', ''), 'http://', ''), '/') > 0
			THEN substr(replace(replace(url, 'https://', ''), 'http://', ''), 1, instr(replace(replace(url, 'https://', ''), 'http://', ''), '/') - 1)
			ELSE replace(replace(url, 'https://', ''), 'http://', '')
		END as domain,
		COUNT(*) as cnt
		FROM downloads
		GROUP BY domain ORDER BY cnt DESC LIMIT 10`)
	if err != nil {
		return s, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var td TopDomain
		if err := rows2.Scan(&td.Domain, &td.Count); err != nil {
			return s, err
		}
		s.TopDomains = append(s.TopDomains, td)
	}

	return s, rows2.Err()
}
