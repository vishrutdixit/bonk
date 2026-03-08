package tui

import (
	"fmt"
	"testing"
)

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

	if got := m.mainPanelWidth(); got != 40 {
		t.Fatalf("mainPanelWidth = %d, want 40 min clamp", got)
	}
	if got := m.mainContentWidth(); got != 40 {
		t.Fatalf("mainContentWidth = %d, want 40 min clamp", got)
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
