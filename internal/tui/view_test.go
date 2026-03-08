package tui

import (
	"strings"
	"testing"
)

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
		turn:  10,
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
