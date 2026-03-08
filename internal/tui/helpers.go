package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"bonk/internal/skills"
)

func domainShort(domain string) string {
	switch domain {
	case "data-structures":
		return "ds"
	case "algorithm-patterns":
		return "algo"
	case "system-design":
		return "sys"
	case "system-design-practical":
		return "sysp"
	case "leetcode-patterns":
		return "lc"
	default:
		return ""
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) ShouldContinue() bool {
	return m.continueToNext
}

func (m Model) Err() error {
	return m.err
}

func (m Model) SelectedDomain() string {
	return m.selectedDomain
}

func (m Model) Skill() *skills.Skill {
	return m.skill
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		width = 60
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if len(line) <= width {
			result.WriteString(line + "\n")
			continue
		}
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if len(currentLine)+len(word)+1 > width {
				result.WriteString(currentLine + "\n")
				currentLine = word
			} else if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
		if currentLine != "" {
			result.WriteString(currentLine + "\n")
		}
	}
	return strings.TrimSuffix(result.String(), "\n")
}

func formatRating(r float64) string {
	if r == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", r)
}

func pickRandomSkillFromDomain(domain string) *skills.Skill {
	candidates := skills.ListByDomain(domain)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[rand.Intn(len(candidates))]
}

func (m Model) domainPickerEnabled() bool {
	return m.allowDomainPicker
}

func (m Model) effectiveDomain() string {
	if m.selectedDomain != "" {
		return m.selectedDomain
	}
	if m.skill == nil {
		return ""
	}
	return m.skill.Domain
}

func (m Model) mainPanelWidth() int {
	mainWidth := m.width - m.sidebarWidth() - 3
	if mainWidth < 40 {
		mainWidth = 40
	}
	return mainWidth
}

func (m Model) mainContentWidth() int {
	mainWidth := m.mainPanelWidth() - 1
	if mainWidth < 40 {
		mainWidth = 40
	}
	return mainWidth
}

func (m Model) sidebarWidth() int {
	if m.showDebug {
		return 52
	}
	return 16
}

func (m *Model) syncLayout() {
	contentWidth := m.mainContentWidth()
	m.textarea.SetWidth(max(20, contentWidth-2))
	m.viewport.Width = max(20, contentWidth)
	m.viewport.Height = max(5, m.height-15)
}

func cycleDomainSelection(current string, delta int) string {
	options := []string{
		"data-structures",
		"algorithm-patterns",
		"system-design",
		"system-design-practical",
		"leetcode-patterns",
	}

	idx := 0
	for i, option := range options {
		if option == current {
			idx = i
			break
		}
	}

	next := (idx + delta + len(options)) % len(options)
	return options[next]
}

func ratingGlyph(rating int) string {
	color := "241"
	switch rating {
	case 1:
		color = "210"
	case 2:
		color = "214"
	case 3:
		color = "114"
	case 4:
		color = "212"
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render("●")
}

func formatDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func truncateASCII(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func relativeTime(sqliteDateTime string) string {
	t, err := time.Parse("2006-01-02 15:04:05", sqliteDateTime)
	if err != nil {
		return sqliteDateTime
	}
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	return fmt.Sprintf("%dd ago", days)
}
