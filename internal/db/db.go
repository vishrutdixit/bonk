package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  skill_id TEXT NOT NULL,
  started_at TEXT NOT NULL DEFAULT (datetime('now')),
  finished_at TEXT,
  rating INTEGER,
  assessment TEXT
);

CREATE TABLE IF NOT EXISTS exchanges (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  turn INTEGER NOT NULL,
  question TEXT NOT NULL,
  question_type TEXT,
  facet TEXT,
  answer TEXT,
  struggled INTEGER DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY(session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS scheduling (
  skill_id TEXT PRIMARY KEY,
  due_at TEXT NOT NULL DEFAULT (datetime('now')),
  stability REAL NOT NULL DEFAULT 1.0,
  difficulty REAL NOT NULL DEFAULT 5.0,
  lapses INTEGER NOT NULL DEFAULT 0,
  last_rating INTEGER,
  last_reviewed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_scheduling_due ON scheduling(due_at);
CREATE INDEX IF NOT EXISTS idx_exchanges_session ON exchanges(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_skill ON sessions(skill_id);
`

type DB struct {
	conn *sql.DB
}

func dbPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bonk", "data.sqlite")
}

func Open() (*DB, error) {
	path := dbPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Enable WAL and foreign keys
	if _, err := conn.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set pragmas: %w", err)
	}

	// Create tables
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// Session management

func (db *DB) CreateSession(skillID string) (string, error) {
	id := uuid.New().String()
	_, err := db.conn.Exec(
		"INSERT INTO sessions (id, skill_id) VALUES (?, ?)",
		id, skillID,
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return id, nil
}

func (db *DB) FinishSession(sessionID string, rating int, assessment string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Update session
	_, err = tx.Exec(
		"UPDATE sessions SET finished_at = datetime('now'), rating = ?, assessment = ? WHERE id = ?",
		rating, assessment, sessionID,
	)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	// Get skill_id
	var skillID string
	err = tx.QueryRow("SELECT skill_id FROM sessions WHERE id = ?", sessionID).Scan(&skillID)
	if err != nil {
		return fmt.Errorf("get skill_id: %w", err)
	}

	// Get current scheduling data (or defaults for new skill)
	var stability, difficulty float64
	var lapses int
	err = tx.QueryRow(`
		SELECT COALESCE(stability, 1.0), COALESCE(difficulty, 2.5), COALESCE(lapses, 0)
		FROM scheduling WHERE skill_id = ?
	`, skillID).Scan(&stability, &difficulty, &lapses)
	if err == sql.ErrNoRows {
		stability, difficulty, lapses = 1.0, 2.5, 0
	} else if err != nil {
		return fmt.Errorf("get scheduling: %w", err)
	}

	// SM-2 algorithm
	// Map our 1-4 rating to SM-2's 0-5 scale: 1->1, 2->2, 3->4, 4->5
	var q float64
	switch rating {
	case 1:
		q = 1 // Again - complete failure
	case 2:
		q = 2 // Hard - barely passed
	case 3:
		q = 4 // Good - correct with effort
	case 4:
		q = 5 // Easy - perfect recall
	}

	// Update easiness factor (difficulty in our schema, but inverted meaning)
	// EF' = EF + (0.1 - (5-q) * (0.08 + (5-q) * 0.02))
	difficulty = difficulty + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if difficulty < 1.3 {
		difficulty = 1.3 // Minimum EF
	}

	// Calculate interval
	var intervalDays float64
	if rating <= 2 {
		// Lapse - reset stability, count the lapse
		lapses++
		stability = 1.0
		intervalDays = 1
	} else {
		// Success - multiply interval by easiness factor
		if stability < 1 {
			stability = 1
		}
		stability = stability * difficulty
		intervalDays = stability
	}

	// Cap interval at 365 days
	if intervalDays > 365 {
		intervalDays = 365
	}

	// Update scheduling
	_, err = tx.Exec(`
		INSERT INTO scheduling (skill_id, due_at, stability, difficulty, lapses, last_rating, last_reviewed_at)
		VALUES (?, datetime('now', '+' || ? || ' days'), ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(skill_id) DO UPDATE SET
			due_at = datetime('now', '+' || ? || ' days'),
			stability = ?,
			difficulty = ?,
			lapses = ?,
			last_rating = ?,
			last_reviewed_at = datetime('now')
	`, skillID, int(intervalDays), stability, difficulty, lapses, rating,
		int(intervalDays), stability, difficulty, lapses, rating)
	if err != nil {
		return fmt.Errorf("update scheduling: %w", err)
	}

	return tx.Commit()
}

// Exchange management

type Exchange struct {
	Turn     int
	Question string
	Answer   string
	Facet    string
}

type SessionDetail struct {
	ID         string
	SkillID    string
	StartedAt  string
	FinishedAt string
	Rating     int
	Assessment string
	Exchanges  []Exchange
}

func (db *DB) GetLastSession(skillID string) (*SessionDetail, error) {
	var s SessionDetail
	var finishedAt, assessment sql.NullString
	var rating sql.NullInt64

	query := `
		SELECT id, skill_id, started_at, finished_at, rating, assessment
		FROM sessions
		WHERE finished_at IS NOT NULL
	`
	args := []interface{}{}
	if skillID != "" {
		query += " AND skill_id = ?"
		args = append(args, skillID)
	}
	query += " ORDER BY finished_at DESC LIMIT 1"

	err := db.conn.QueryRow(query, args...).Scan(
		&s.ID, &s.SkillID, &s.StartedAt, &finishedAt, &rating, &assessment,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if finishedAt.Valid {
		s.FinishedAt = finishedAt.String
	}
	if rating.Valid {
		s.Rating = int(rating.Int64)
	}
	if assessment.Valid {
		s.Assessment = assessment.String
	}

	// Get exchanges
	rows, err := db.conn.Query(`
		SELECT turn, question, answer, facet
		FROM exchanges
		WHERE session_id = ?
		ORDER BY turn ASC
	`, s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e Exchange
		var facet sql.NullString
		if err := rows.Scan(&e.Turn, &e.Question, &e.Answer, &facet); err != nil {
			return nil, err
		}
		if facet.Valid {
			e.Facet = facet.String
		}
		s.Exchanges = append(s.Exchanges, e)
	}

	return &s, rows.Err()
}

func (db *DB) SaveExchange(sessionID string, turn int, question, questionType, facet, answer string, struggled bool) error {
	id := uuid.New().String()
	struggledInt := 0
	if struggled {
		struggledInt = 1
	}

	_, err := db.conn.Exec(
		"INSERT INTO exchanges (id, session_id, turn, question, question_type, facet, answer, struggled) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id, sessionID, turn, question, questionType, facet, answer, struggledInt,
	)
	if err != nil {
		return fmt.Errorf("save exchange: %w", err)
	}
	return nil
}

// Stats

type SkillStats struct {
	SkillID   string
	Count     int
	AvgRating float64
}

type FacetStats struct {
	Facet     string
	Total     int
	Struggled int
}

func (db *DB) GetTotalSessions() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sessions WHERE finished_at IS NOT NULL").Scan(&count)
	return count, err
}

func (db *DB) GetSkillStats(limit int) ([]SkillStats, error) {
	rows, err := db.conn.Query(`
		SELECT skill_id, COUNT(*) as count, AVG(rating) as avg_rating
		FROM sessions
		WHERE finished_at IS NOT NULL
		GROUP BY skill_id
		ORDER BY count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []SkillStats
	for rows.Next() {
		var s SkillStats
		var avgRating sql.NullFloat64
		if err := rows.Scan(&s.SkillID, &s.Count, &avgRating); err != nil {
			return nil, err
		}
		if avgRating.Valid {
			s.AvgRating = avgRating.Float64
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (db *DB) GetWeakFacets(limit int) ([]FacetStats, error) {
	rows, err := db.conn.Query(`
		SELECT facet, COUNT(*) as total, SUM(struggled) as struggled
		FROM exchanges
		WHERE facet IS NOT NULL AND facet != ''
		GROUP BY facet
		HAVING total >= 3
		ORDER BY CAST(struggled AS FLOAT) / total DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []FacetStats
	for rows.Next() {
		var s FacetStats
		if err := rows.Scan(&s.Facet, &s.Total, &s.Struggled); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// Get average rating for a skill (0 if no data)
func (db *DB) GetSkillAvgRating(skillID string) (float64, int, error) {
	var avg sql.NullFloat64
	var count int
	err := db.conn.QueryRow(`
		SELECT AVG(rating), COUNT(*)
		FROM sessions
		WHERE skill_id = ? AND finished_at IS NOT NULL AND rating IS NOT NULL
	`, skillID).Scan(&avg, &count)
	if err != nil {
		return 0, 0, err
	}
	if avg.Valid {
		return avg.Float64, count, nil
	}
	return 0, 0, nil
}

// Get overall average rating across all skills
func (db *DB) GetOverallAvgRating() (float64, int, error) {
	var avg sql.NullFloat64
	var count int
	err := db.conn.QueryRow(`
		SELECT AVG(rating), COUNT(*)
		FROM sessions
		WHERE finished_at IS NOT NULL AND rating IS NOT NULL
	`).Scan(&avg, &count)
	if err != nil {
		return 0, 0, err
	}
	if avg.Valid {
		return avg.Float64, count, nil
	}
	return 0, 0, nil
}

// Scheduling queries

type SchedulingInfo struct {
	SkillID    string
	DueAt      string
	Stability  float64
	Difficulty float64
	Lapses     int
}

// GetDueSkills returns skills that are due for review (due_at <= now)
func (db *DB) GetDueSkills() ([]SchedulingInfo, error) {
	rows, err := db.conn.Query(`
		SELECT skill_id, due_at, stability, difficulty, lapses
		FROM scheduling
		WHERE due_at <= datetime('now')
		ORDER BY due_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []SchedulingInfo
	for rows.Next() {
		var s SchedulingInfo
		if err := rows.Scan(&s.SkillID, &s.DueAt, &s.Stability, &s.Difficulty, &s.Lapses); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

// GetNewSkills returns skill IDs that have never been reviewed
func (db *DB) GetNewSkills(allSkillIDs []string) []string {
	reviewed := make(map[string]bool)
	rows, _ := db.conn.Query("SELECT skill_id FROM scheduling")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			rows.Scan(&id)
			reviewed[id] = true
		}
	}

	var newSkills []string
	for _, id := range allSkillIDs {
		if !reviewed[id] {
			newSkills = append(newSkills, id)
		}
	}
	return newSkills
}

// GetDueCount returns count of skills due for review
func (db *DB) GetDueCount() (int, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM scheduling WHERE due_at <= datetime('now')
	`).Scan(&count)
	return count, err
}

// History context for LLM

func (db *DB) GetHistoryContext(skillID string, limit int) (string, error) {
	rows, err := db.conn.Query(`
		SELECT e.facet, e.struggled
		FROM exchanges e
		JOIN sessions s ON s.id = e.session_id
		WHERE s.skill_id = ? AND e.facet IS NOT NULL AND e.facet != ''
		ORDER BY e.created_at DESC
		LIMIT ?
	`, skillID, limit*3)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type facetStat struct {
		total     int
		struggled int
	}
	facetStats := make(map[string]*facetStat)
	var lines []string

	for rows.Next() {
		var facet string
		var struggled int
		if err := rows.Scan(&facet, &struggled); err != nil {
			return "", err
		}

		if _, ok := facetStats[facet]; !ok {
			facetStats[facet] = &facetStat{}
		}
		facetStats[facet].total++
		if struggled == 1 {
			facetStats[facet].struggled++
			lines = append(lines, fmt.Sprintf("- Asked about %s: struggled", facet))
		} else {
			lines = append(lines, fmt.Sprintf("- Asked about %s: got it", facet))
		}
	}

	if len(lines) == 0 {
		return "", nil
	}

	// Find weak facets
	var weak []string
	for f, s := range facetStats {
		if float64(s.struggled)/float64(s.total) > 0.3 {
			weak = append(weak, f)
		}
	}

	result := "Recent questions:\n"
	if len(lines) > 10 {
		lines = lines[:10]
	}
	for _, l := range lines {
		result += l + "\n"
	}

	if len(weak) > 0 {
		result += fmt.Sprintf("\nWeak areas (prioritize these): %s", weak)
	}

	return result, nil
}

// Streak calculates consecutive days with at least one completed session
func (db *DB) GetStreak() (current int, longest int, err error) {
	rows, err := db.conn.Query(`
		SELECT DISTINCT date(finished_at) as d
		FROM sessions
		WHERE finished_at IS NOT NULL
		ORDER BY d DESC
	`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return 0, 0, err
		}
		dates = append(dates, d)
	}

	if len(dates) == 0 {
		return 0, 0, nil
	}

	// Calculate current streak (must include today or yesterday)
	current = 1
	for i := 1; i < len(dates); i++ {
		if isConsecutiveDay(dates[i-1], dates[i]) {
			current++
		} else {
			break
		}
	}

	// Check if streak is still active (last drill was today or yesterday)
	if !isRecentDay(dates[0]) {
		current = 0
	}

	// Calculate longest streak
	longest = 1
	streak := 1
	for i := 1; i < len(dates); i++ {
		if isConsecutiveDay(dates[i-1], dates[i]) {
			streak++
			if streak > longest {
				longest = streak
			}
		} else {
			streak = 1
		}
	}

	return current, longest, nil
}

func parseDate(s string) (time.Time, error) {
	// SQLite date format: YYYY-MM-DD
	return time.Parse("2006-01-02", s)
}

func isConsecutiveDay(newer, older string) bool {
	n, err1 := parseDate(newer)
	o, err2 := parseDate(older)
	if err1 != nil || err2 != nil {
		return false
	}
	diff := n.Sub(o).Hours() / 24
	return diff == 1
}

func isRecentDay(date string) bool {
	d, err := parseDate(date)
	if err != nil {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	diff := today.Sub(d).Hours() / 24
	return diff <= 1
}

// DomainStats holds aggregate stats for a domain
type DomainStats struct {
	Domain       string
	TotalSkills  int
	Practiced    int
	AvgRating    float64
	SessionCount int
}

// GetDomainStats returns stats grouped by domain
func (db *DB) GetDomainStats(skillsByDomain map[string][]string) ([]DomainStats, error) {
	stats := make(map[string]*DomainStats)

	// Initialize with total counts
	for domain, skills := range skillsByDomain {
		stats[domain] = &DomainStats{
			Domain:      domain,
			TotalSkills: len(skills),
		}
	}

	// Get practiced skills and ratings per domain
	rows, err := db.conn.Query(`
		SELECT skill_id, COUNT(*) as count, AVG(rating) as avg_rating
		FROM sessions
		WHERE finished_at IS NOT NULL
		GROUP BY skill_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map skill_id to domain (caller provides this mapping)
	skillToDomain := make(map[string]string)
	for domain, skills := range skillsByDomain {
		for _, s := range skills {
			skillToDomain[s] = domain
		}
	}

	for rows.Next() {
		var skillID string
		var count int
		var avgRating sql.NullFloat64
		if err := rows.Scan(&skillID, &count, &avgRating); err != nil {
			return nil, err
		}

		domain, ok := skillToDomain[skillID]
		if !ok {
			continue
		}

		ds := stats[domain]
		ds.Practiced++
		ds.SessionCount += count
		if avgRating.Valid {
			// Weighted average
			ds.AvgRating = (ds.AvgRating*float64(ds.SessionCount-count) + avgRating.Float64*float64(count)) / float64(ds.SessionCount)
		}
	}

	// Convert to slice
	result := make([]DomainStats, 0, len(stats))
	for _, ds := range stats {
		result = append(result, *ds)
	}
	return result, nil
}

// RecentSession holds info about a past session
type RecentSession struct {
	SkillID    string
	Rating     int
	FinishedAt string
}

// GetRecentSessions returns the N most recent completed sessions
func (db *DB) GetRecentSessions(limit int) ([]RecentSession, error) {
	rows, err := db.conn.Query(`
		SELECT skill_id, rating, finished_at
		FROM sessions
		WHERE finished_at IS NOT NULL AND rating IS NOT NULL
		ORDER BY finished_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []RecentSession
	for rows.Next() {
		var s RecentSession
		if err := rows.Scan(&s.SkillID, &s.Rating, &s.FinishedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetDueThisWeek returns count of skills due within 7 days
func (db *DB) GetDueThisWeek() (int, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM scheduling
		WHERE due_at <= datetime('now', '+7 days')
	`).Scan(&count)
	return count, err
}

// GetRecentRatings returns the last N ratings for sparkline display
func (db *DB) GetRecentRatings(limit int) ([]int, error) {
	rows, err := db.conn.Query(`
		SELECT rating FROM sessions
		WHERE finished_at IS NOT NULL AND rating IS NOT NULL
		ORDER BY finished_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []int
	for rows.Next() {
		var r int
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}
	// Reverse to show oldest to newest (left to right)
	for i, j := 0, len(ratings)-1; i < j; i, j = i+1, j-1 {
		ratings[i], ratings[j] = ratings[j], ratings[i]
	}
	return ratings, rows.Err()
}

// GetTodaySessionCount returns sessions completed today
func (db *DB) GetTodaySessionCount() (int, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sessions
		WHERE date(finished_at) = date('now') AND finished_at IS NOT NULL
	`).Scan(&count)
	return count, err
}
