package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"bonk/internal/skills"
)

func mustSkill(t *testing.T, id string) *skills.Skill {
	t.Helper()
	s := skills.Get(id)
	if s == nil {
		t.Fatalf("skills.Get(%q) = nil", id)
	}
	return s
}

func TestDomainShort(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"data-structures", "ds"},
		{"algorithm-patterns", "algo"},
		{"system-design", "sys"},
		{"system-design-practical", "sysp"},
		{"leetcode-patterns", "lc"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		if got := domainShort(tt.in); got != tt.want {
			t.Fatalf("domainShort(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestWordWrap(t *testing.T) {
	in := "alpha beta gamma delta"
	got := wordWrap(in, 10)
	want := "alpha beta\ngamma\ndelta"
	if got != want {
		t.Fatalf("wordWrap() = %q, want %q", got, want)
	}

	// Width guard path.
	got = wordWrap("a b c", 0)
	if !strings.Contains(got, "a b c") {
		t.Fatalf("wordWrap width fallback unexpected: %q", got)
	}
}

func TestTruncateASCII(t *testing.T) {
	if got := truncateASCII("hello", 10); got != "hello" {
		t.Fatalf("truncateASCII no-op = %q, want hello", got)
	}
	if got := truncateASCII("helloworld", 5); got != "he..." {
		t.Fatalf("truncateASCII ellipsis = %q, want he...", got)
	}
	if got := truncateASCII("hello", 2); got != "he" {
		t.Fatalf("truncateASCII hard cut = %q, want he", got)
	}
}

func TestFormatHelpers(t *testing.T) {
	if got := formatRating(0); got != "—" {
		t.Fatalf("formatRating(0) = %q, want —", got)
	}
	if got := formatRating(3.25); got != "3.2" {
		t.Fatalf("formatRating(3.25) = %q, want 3.2", got)
	}
	if got := formatDate("2026-03-07 10:11:12"); got != "2026-03-07" {
		t.Fatalf("formatDate() = %q, want 2026-03-07", got)
	}
	if got := formatDate("short"); got != "short" {
		t.Fatalf("formatDate short = %q, want short", got)
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now().UTC()

	justNow := now.Add(-20 * time.Second).Format("2006-01-02 15:04:05")
	if got := relativeTime(justNow); got != "just now" {
		t.Fatalf("relativeTime just now = %q, want just now", got)
	}

	mins := now.Add(-3 * time.Minute).Format("2006-01-02 15:04:05")
	if got := relativeTime(mins); got != "3m ago" {
		t.Fatalf("relativeTime minutes = %q, want 3m ago", got)
	}

	hours := now.Add(-2 * time.Hour).Format("2006-01-02 15:04:05")
	if got := relativeTime(hours); got != "2h ago" {
		t.Fatalf("relativeTime hours = %q, want 2h ago", got)
	}

	day := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	if got := relativeTime(day); got != "1d ago" {
		t.Fatalf("relativeTime day = %q, want 1d ago", got)
	}

	multi := now.Add(-72 * time.Hour).Format("2006-01-02 15:04:05")
	if got := relativeTime(multi); got != "3d ago" {
		t.Fatalf("relativeTime days = %q, want 3d ago", got)
	}

	if got := relativeTime("not-a-date"); got != "not-a-date" {
		t.Fatalf("relativeTime invalid = %q, want input passthrough", got)
	}
}

func TestCycleDomainSelection(t *testing.T) {
	if got := cycleDomainSelection("data-structures", 1); got != "algorithm-patterns" {
		t.Fatalf("cycle forward = %q, want algorithm-patterns", got)
	}
	if got := cycleDomainSelection("data-structures", -1); got != "leetcode-patterns" {
		t.Fatalf("cycle backward wrap = %q, want leetcode-patterns", got)
	}
	if got := cycleDomainSelection("unknown", 1); got != "algorithm-patterns" {
		t.Fatalf("cycle unknown default = %q, want algorithm-patterns", got)
	}
}

func TestModelDomainAndLayoutHelpers(t *testing.T) {
	m := Model{
		skill:             mustSkill(t, "hash-maps"),
		allowDomainPicker: true,
		selectedDomain:    "leetcode-patterns",
		width:             50,
		height:            20,
	}
	if !m.domainPickerEnabled() {
		t.Fatal("domainPickerEnabled = false, want true")
	}
	if got := m.effectiveDomain(); got != "leetcode-patterns" {
		t.Fatalf("effectiveDomain with selected = %q, want leetcode-patterns", got)
	}

	// Width clamping paths.
	if got := m.mainPanelWidth(); got != 40 {
		t.Fatalf("mainPanelWidth = %d, want 40 min clamp", got)
	}
	if got := m.mainContentWidth(); got != 40 {
		t.Fatalf("mainContentWidth = %d, want 40 min clamp", got)
	}
}

func TestRenderHeaderTurnAndPhase(t *testing.T) {
	base := Model{
		skill: mustSkill(t, "hash-maps"),
		turn:  3,
	}
	h := base.renderHeader()
	if !strings.Contains(h, "bonk") || !strings.Contains(h, "ds") || !strings.Contains(h, "turn 3") {
		t.Fatalf("renderHeader regular missing expected pieces: %q", h)
	}

	sysp := Model{
		skill: mustSkill(t, "design-twitter"),
		phase: "api",
		turn:  10, // should be ignored in phase mode
	}
	ph := sysp.renderHeader()
	if !strings.Contains(ph, "sysp") || !strings.Contains(ph, "API Design (3/6)") {
		t.Fatalf("renderHeader sysp missing phase info: %q", ph)
	}
	if strings.Contains(ph, "turn 10") {
		t.Fatalf("renderHeader sysp should not include turn marker: %q", ph)
	}
}

func TestRenderSparklineAndRatingLabelAndGlyph(t *testing.T) {
	s := renderSparkline([]int{1, 2, 3, 4, 0, 9})
	// ANSI codes are present; assert glyphs exist in order.
	for _, block := range []string{"▁", "▃", "▅", "▇"} {
		if !strings.Contains(s, block) {
			t.Fatalf("renderSparkline missing %q in %q", block, s)
		}
	}

	if got := llmRatingLabel(0); got != "" {
		t.Fatalf("llmRatingLabel(0) = %q, want empty", got)
	}
	for i := 1; i <= 4; i++ {
		if got := llmRatingLabel(i); got == "" {
			t.Fatalf("llmRatingLabel(%d) = empty, want label", i)
		}
		if g := ratingGlyph(i); !strings.Contains(g, "●") {
			t.Fatalf("ratingGlyph(%d) missing dot: %q", i, g)
		}
	}
}

func TestMinMax(t *testing.T) {
	if got := min(1, 2); got != 1 {
		t.Fatalf("min = %d, want 1", got)
	}
	if got := max(1, 2); got != 2 {
		t.Fatalf("max = %d, want 2", got)
	}
}

func TestShouldContinueErrSelectedDomainSkill(t *testing.T) {
	skill := mustSkill(t, "hash-maps")
	testErr := fmt.Errorf("boom")
	m := Model{
		continueToNext: true,
		err:            testErr,
		selectedDomain: "data-structures",
		skill:          skill,
	}
	if !m.ShouldContinue() {
		t.Fatal("ShouldContinue = false, want true")
	}
	if m.Err() == nil || m.Err().Error() != "boom" {
		t.Fatalf("Err() = %v, want boom", m.Err())
	}
	if m.SelectedDomain() != "data-structures" {
		t.Fatalf("SelectedDomain() = %q, want data-structures", m.SelectedDomain())
	}
	if m.Skill() != skill {
		t.Fatalf("Skill() pointer mismatch")
	}
}
