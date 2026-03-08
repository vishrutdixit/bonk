package tui

import (
	"testing"

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
