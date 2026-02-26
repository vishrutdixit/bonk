package skills

import (
	"strings"
	"testing"
)

func TestGetGuide(t *testing.T) {
	// Test existing guide
	guide := GetGuide("ci-cd-systems")
	if guide == "" {
		t.Error("expected ci-cd-systems guide to exist")
	}
	if !strings.Contains(guide, "CI/CD") {
		t.Error("guide should contain 'CI/CD'")
	}

	// Test non-existent guide returns empty
	empty := GetGuide("nonexistent-skill")
	if empty != "" {
		t.Error("expected empty string for non-existent guide")
	}
}
