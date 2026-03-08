package db

import (
	"math"
	"strings"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	return database
}

func createFinishedSession(t *testing.T, database *DB, skillID string, rating int, finishedAt string) string {
	t.Helper()

	sessionID, err := database.CreateSession(skillID)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if err := database.SaveExchange(sessionID, 1, "Q1", "opening", "facet-a", "A1", false); err != nil {
		t.Fatalf("SaveExchange() turn 1 error = %v", err)
	}
	if err := database.SaveExchange(sessionID, 2, "Q2", "followup", "facet-b", "A2", true); err != nil {
		t.Fatalf("SaveExchange() turn 2 error = %v", err)
	}
	if err := database.FinishSession(sessionID, rating, "solid"); err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	if _, err := database.conn.Exec("UPDATE sessions SET started_at = ?, finished_at = ? WHERE id = ?", finishedAt, finishedAt, sessionID); err != nil {
		t.Fatalf("updating session timestamps error = %v", err)
	}

	return sessionID
}

func createFinishedSessionAtOffset(t *testing.T, database *DB, skillID string, rating int, offset string) string {
	t.Helper()

	sessionID, err := database.CreateSession(skillID)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if err := database.FinishSession(sessionID, rating, "ok"); err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	if _, err := database.conn.Exec(`
		UPDATE sessions
		SET started_at = datetime('now', ?), finished_at = datetime('now', ?)
		WHERE id = ?
	`, offset, offset, sessionID); err != nil {
		t.Fatalf("setting session offset error = %v", err)
	}
	return sessionID
}

func getScheduling(t *testing.T, database *DB, skillID string) (stability, difficulty float64, lapses, lastRating, intervalDays int) {
	t.Helper()

	err := database.conn.QueryRow(`
		SELECT stability, difficulty, lapses, last_rating,
		       CAST((julianday(due_at) - julianday(last_reviewed_at)) AS INTEGER)
		FROM scheduling
		WHERE skill_id = ?
	`, skillID).Scan(&stability, &difficulty, &lapses, &lastRating, &intervalDays)
	if err != nil {
		t.Fatalf("query scheduling error = %v", err)
	}
	return stability, difficulty, lapses, lastRating, intervalDays
}

func TestFinishSessionSM2FlowAndCap(t *testing.T) {
	database := openTestDB(t)

	sessionID, err := database.CreateSession("hash-maps")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if err := database.FinishSession(sessionID, 4, "great"); err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	stability, difficulty, lapses, lastRating, intervalDays := getScheduling(t, database, "hash-maps")
	if math.Abs(stability-2.6) > 0.001 {
		t.Fatalf("stability = %.3f, want 2.6", stability)
	}
	if math.Abs(difficulty-2.6) > 0.001 {
		t.Fatalf("difficulty = %.3f, want 2.6", difficulty)
	}
	if lapses != 0 {
		t.Fatalf("lapses = %d, want 0", lapses)
	}
	if lastRating != 4 {
		t.Fatalf("lastRating = %d, want 4", lastRating)
	}
	if intervalDays != 2 {
		t.Fatalf("intervalDays = %d, want 2", intervalDays)
	}

	sessionID2, err := database.CreateSession("hash-maps")
	if err != nil {
		t.Fatalf("CreateSession() second error = %v", err)
	}
	if err := database.FinishSession(sessionID2, 2, "rough"); err != nil {
		t.Fatalf("FinishSession() second error = %v", err)
	}

	stability, difficulty, lapses, lastRating, intervalDays = getScheduling(t, database, "hash-maps")
	if math.Abs(stability-1.0) > 0.001 {
		t.Fatalf("stability after lapse = %.3f, want 1.0", stability)
	}
	if math.Abs(difficulty-2.28) > 0.01 {
		t.Fatalf("difficulty after rating 2 = %.3f, want about 2.28", difficulty)
	}
	if lapses != 1 {
		t.Fatalf("lapses after rating 2 = %d, want 1", lapses)
	}
	if lastRating != 2 {
		t.Fatalf("lastRating after rating 2 = %d, want 2", lastRating)
	}
	if intervalDays != 1 {
		t.Fatalf("intervalDays after lapse = %d, want 1", intervalDays)
	}

	if _, err := database.conn.Exec(`
		INSERT INTO scheduling (skill_id, due_at, stability, difficulty, lapses)
		VALUES ('arrays', datetime('now'), 500.0, 3.0, 0)
		ON CONFLICT(skill_id) DO UPDATE SET stability = 500.0, difficulty = 3.0, lapses = 0, due_at = datetime('now')
	`); err != nil {
		t.Fatalf("seeding scheduling row error = %v", err)
	}

	sessionID3, err := database.CreateSession("arrays")
	if err != nil {
		t.Fatalf("CreateSession() third error = %v", err)
	}
	if err := database.FinishSession(sessionID3, 4, "easy"); err != nil {
		t.Fatalf("FinishSession() third error = %v", err)
	}

	_, _, _, _, intervalDays = getScheduling(t, database, "arrays")
	if intervalDays != 365 {
		t.Fatalf("intervalDays cap = %d, want 365", intervalDays)
	}
}

func TestGetRecentSessionsReturnsIDsInFinishedOrder(t *testing.T) {
	database := openTestDB(t)

	olderID := createFinishedSession(t, database, "hash-maps", 2, "2026-03-01 10:00:00")
	newerID := createFinishedSession(t, database, "heaps", 4, "2026-03-02 10:00:00")

	sessions, err := database.GetRecentSessions(10)
	if err != nil {
		t.Fatalf("GetRecentSessions() error = %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("GetRecentSessions() len = %d, want 2", len(sessions))
	}
	if sessions[0].ID != newerID {
		t.Fatalf("sessions[0].ID = %q, want %q", sessions[0].ID, newerID)
	}
	if sessions[0].SkillID != "heaps" {
		t.Fatalf("sessions[0].SkillID = %q, want %q", sessions[0].SkillID, "heaps")
	}
	if sessions[1].ID != olderID {
		t.Fatalf("sessions[1].ID = %q, want %q", sessions[1].ID, olderID)
	}
}

func TestGetSessionByIDIncludesTranscript(t *testing.T) {
	database := openTestDB(t)

	sessionID := createFinishedSession(t, database, "hash-maps", 3, "2026-03-03 09:00:00")

	session, err := database.GetSessionByID(sessionID)
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}
	if session == nil {
		t.Fatal("GetSessionByID() = nil, want session")
	}
	if session.ID != sessionID {
		t.Fatalf("session.ID = %q, want %q", session.ID, sessionID)
	}
	if session.SkillID != "hash-maps" {
		t.Fatalf("session.SkillID = %q, want %q", session.SkillID, "hash-maps")
	}
	if session.Rating != 3 {
		t.Fatalf("session.Rating = %d, want 3", session.Rating)
	}
	if len(session.Exchanges) != 2 {
		t.Fatalf("len(session.Exchanges) = %d, want 2", len(session.Exchanges))
	}
	if session.Exchanges[0].Turn != 1 || session.Exchanges[0].Question != "Q1" || session.Exchanges[0].Answer != "A1" {
		t.Fatalf("first exchange = %#v, want turn/question/answer 1/Q1/A1", session.Exchanges[0])
	}
	if session.Exchanges[1].Turn != 2 || session.Exchanges[1].Facet != "facet-b" {
		t.Fatalf("second exchange = %#v, want turn/facet 2/facet-b", session.Exchanges[1])
	}
	if !session.Exchanges[1].Struggled {
		t.Fatalf("session.Exchanges[1].Struggled = %v, want true", session.Exchanges[1].Struggled)
	}
}

func TestGetSessionByIDReturnsNilWhenMissing(t *testing.T) {
	database := openTestDB(t)

	session, err := database.GetSessionByID("missing")
	if err != nil {
		t.Fatalf("GetSessionByID() error = %v", err)
	}
	if session != nil {
		t.Fatalf("GetSessionByID() = %#v, want nil", session)
	}
}

func TestGetLastSessionSkillFilter(t *testing.T) {
	database := openTestDB(t)

	createFinishedSession(t, database, "hash-maps", 2, "2026-03-01 10:00:00")
	createFinishedSession(t, database, "heaps", 4, "2026-03-02 10:00:00")
	createFinishedSession(t, database, "hash-maps", 3, "2026-03-03 10:00:00")

	lastAny, err := database.GetLastSession("")
	if err != nil {
		t.Fatalf("GetLastSession(empty) error = %v", err)
	}
	if lastAny == nil || lastAny.SkillID != "hash-maps" || lastAny.Rating != 3 {
		t.Fatalf("lastAny = %#v, want latest hash-maps session", lastAny)
	}

	lastHeaps, err := database.GetLastSession("heaps")
	if err != nil {
		t.Fatalf("GetLastSession(heaps) error = %v", err)
	}
	if lastHeaps == nil || lastHeaps.SkillID != "heaps" || lastHeaps.Rating != 4 {
		t.Fatalf("lastHeaps = %#v, want heaps rating 4", lastHeaps)
	}
}

func TestDueAndNewSkillQueries(t *testing.T) {
	database := openTestDB(t)

	if _, err := database.conn.Exec(`
		INSERT INTO scheduling (skill_id, due_at, stability, difficulty, lapses)
		VALUES
			('due-older', datetime('now', '-2 days'), 1.0, 2.5, 0),
			('due-newer', datetime('now', '-1 day'), 1.0, 2.5, 0),
			('future', datetime('now', '+2 days'), 1.0, 2.5, 0)
	`); err != nil {
		t.Fatalf("seed scheduling error = %v", err)
	}

	due, err := database.GetDueSkills()
	if err != nil {
		t.Fatalf("GetDueSkills() error = %v", err)
	}
	if len(due) != 2 {
		t.Fatalf("len(GetDueSkills()) = %d, want 2", len(due))
	}
	if due[0].SkillID != "due-older" || due[1].SkillID != "due-newer" {
		t.Fatalf("due ordering = %#v, want due-older then due-newer", due)
	}

	dueCount, err := database.GetDueCount()
	if err != nil {
		t.Fatalf("GetDueCount() error = %v", err)
	}
	if dueCount != 2 {
		t.Fatalf("GetDueCount() = %d, want 2", dueCount)
	}

	newSkills := database.GetNewSkills([]string{"due-older", "future", "new-a", "new-b"})
	if strings.Join(newSkills, ",") != "new-a,new-b" {
		t.Fatalf("GetNewSkills() = %v, want [new-a new-b]", newSkills)
	}

	dueThisWeek, err := database.GetDueThisWeek()
	if err != nil {
		t.Fatalf("GetDueThisWeek() error = %v", err)
	}
	if dueThisWeek != 3 {
		t.Fatalf("GetDueThisWeek() = %d, want 3", dueThisWeek)
	}
}

func TestGetHistoryContextAndWeakFacets(t *testing.T) {
	database := openTestDB(t)

	sessionID, err := database.CreateSession("graphs")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	entries := []struct {
		facet     string
		struggled bool
	}{
		{"bfs", true},
		{"bfs", true},
		{"bfs", false},
		{"dfs", false},
		{"dfs", false},
		{"dfs", false},
	}
	for i, e := range entries {
		if err := database.SaveExchange(sessionID, i+1, "Q", "followup", e.facet, "A", e.struggled); err != nil {
			t.Fatalf("SaveExchange() error = %v", err)
		}
	}
	if err := database.FinishSession(sessionID, 2, "ok"); err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	ctx, err := database.GetHistoryContext("graphs", 5)
	if err != nil {
		t.Fatalf("GetHistoryContext() error = %v", err)
	}
	if !strings.Contains(ctx, "Recent questions:") {
		t.Fatalf("history context missing header: %q", ctx)
	}
	if !strings.Contains(ctx, "Asked about bfs: struggled") {
		t.Fatalf("history context missing struggled bfs line: %q", ctx)
	}
	if !strings.Contains(ctx, "Weak areas (prioritize these): [bfs]") {
		t.Fatalf("history context missing weak area bfs: %q", ctx)
	}

	facets, err := database.GetWeakFacets(5)
	if err != nil {
		t.Fatalf("GetWeakFacets() error = %v", err)
	}
	if len(facets) != 2 {
		t.Fatalf("len(GetWeakFacets()) = %d, want 2", len(facets))
	}
	if facets[0].Facet != "bfs" || facets[0].Struggled != 2 || facets[0].Total != 3 {
		t.Fatalf("facets[0] = %#v, want bfs with 2/3 struggled", facets[0])
	}
}

func TestStreakAndRecentRatingsAndTodayCount(t *testing.T) {
	database := openTestDB(t)

	createFinishedSessionAtOffset(t, database, "a", 1, "-2 day")
	createFinishedSessionAtOffset(t, database, "b", 2, "-1 day")
	createFinishedSessionAtOffset(t, database, "c", 4, "-0 day")
	createFinishedSessionAtOffset(t, database, "d", 3, "-5 day")
	createFinishedSessionAtOffset(t, database, "e", 3, "-6 day")

	current, longest, err := database.GetStreak()
	if err != nil {
		t.Fatalf("GetStreak() error = %v", err)
	}
	if current != 3 {
		t.Fatalf("current streak = %d, want 3", current)
	}
	if longest != 3 {
		t.Fatalf("longest streak = %d, want 3", longest)
	}

	ratings, err := database.GetRecentRatings(3)
	if err != nil {
		t.Fatalf("GetRecentRatings() error = %v", err)
	}
	if len(ratings) != 3 {
		t.Fatalf("len(GetRecentRatings()) = %d, want 3", len(ratings))
	}
	if ratings[0] != 1 || ratings[1] != 2 || ratings[2] != 4 {
		t.Fatalf("ratings = %v, want [1 2 4] oldest->newest", ratings)
	}

	todayCount, err := database.GetTodaySessionCount()
	if err != nil {
		t.Fatalf("GetTodaySessionCount() error = %v", err)
	}
	if todayCount != 1 {
		t.Fatalf("GetTodaySessionCount() = %d, want 1", todayCount)
	}
}

func TestAveragesAndDomainStats(t *testing.T) {
	database := openTestDB(t)

	createFinishedSession(t, database, "hash-maps", 4, "2026-03-01 10:00:00")
	createFinishedSession(t, database, "hash-maps", 2, "2026-03-02 10:00:00")
	createFinishedSession(t, database, "heaps", 3, "2026-03-03 10:00:00")

	avg, count, err := database.GetSkillAvgRating("hash-maps")
	if err != nil {
		t.Fatalf("GetSkillAvgRating() error = %v", err)
	}
	if count != 2 || math.Abs(avg-3.0) > 0.001 {
		t.Fatalf("GetSkillAvgRating(hash-maps) = (%.3f,%d), want (3.0,2)", avg, count)
	}

	overall, totalCount, err := database.GetOverallAvgRating()
	if err != nil {
		t.Fatalf("GetOverallAvgRating() error = %v", err)
	}
	if totalCount != 3 || math.Abs(overall-3.0) > 0.001 {
		t.Fatalf("GetOverallAvgRating() = (%.3f,%d), want (3.0,3)", overall, totalCount)
	}

	totalSessions, err := database.GetTotalSessions()
	if err != nil {
		t.Fatalf("GetTotalSessions() error = %v", err)
	}
	if totalSessions != 3 {
		t.Fatalf("GetTotalSessions() = %d, want 3", totalSessions)
	}

	skillStats, err := database.GetSkillStats(10)
	if err != nil {
		t.Fatalf("GetSkillStats() error = %v", err)
	}
	if len(skillStats) != 2 {
		t.Fatalf("len(GetSkillStats()) = %d, want 2", len(skillStats))
	}
	if skillStats[0].SkillID != "hash-maps" || skillStats[0].Count != 2 {
		t.Fatalf("skillStats[0] = %#v, want hash-maps count 2", skillStats[0])
	}

	domainStats, err := database.GetDomainStats(map[string][]string{
		"ds":   {"hash-maps", "heaps", "queues"},
		"algo": {"binary-search"},
	})
	if err != nil {
		t.Fatalf("GetDomainStats() error = %v", err)
	}
	byDomain := map[string]DomainStats{}
	for _, ds := range domainStats {
		byDomain[ds.Domain] = ds
	}
	if byDomain["ds"].TotalSkills != 3 || byDomain["ds"].Practiced != 2 || byDomain["ds"].SessionCount != 3 {
		t.Fatalf("ds stats = %#v", byDomain["ds"])
	}
	if math.Abs(byDomain["ds"].AvgRating-3.0) > 0.001 {
		t.Fatalf("ds avg = %.3f, want 3.0", byDomain["ds"].AvgRating)
	}
	if byDomain["algo"].TotalSkills != 1 || byDomain["algo"].Practiced != 0 {
		t.Fatalf("algo stats = %#v", byDomain["algo"])
	}
}

func TestDateHelpers(t *testing.T) {
	if !isConsecutiveDay("2026-03-02", "2026-03-01") {
		t.Fatal("isConsecutiveDay should return true for adjacent dates")
	}
	if isConsecutiveDay("2026-03-03", "2026-03-01") {
		t.Fatal("isConsecutiveDay should return false for non-adjacent dates")
	}
	if isConsecutiveDay("bad", "2026-03-01") {
		t.Fatal("isConsecutiveDay should return false on parse errors")
	}
	if !isRecentDay("2099-01-01") {
		t.Fatal("isRecentDay currently treats future dates as recent; expected true")
	}
}
