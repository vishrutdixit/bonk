package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"bonk/internal/skills"
)

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}

	switch m.state {
	case stateWelcome:
		return m.renderWelcomeCentered()
	default:
		return m.renderWithSidebar()
	}
}

func (m Model) renderWelcomeCentered() string {
	content := m.renderWelcome()
	if m.width <= 0 || m.height <= 0 {
		return content
	}
	blockWidth := lipgloss.Width(content)
	if blockWidth <= 0 {
		return content
	}
	// Keep internal text left-aligned, but move the entire welcome panel as one unit.
	block := lipgloss.NewStyle().Width(blockWidth).Align(lipgloss.Left).Render(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		block,
	)
}

func (m Model) renderWithSidebar() string {
	mainContent := m.renderMainContent()
	sidebar := m.renderSidebar()

	sidebarWidth := m.sidebarWidth()
	mainWidth := m.mainPanelWidth()

	mainStyle := lipgloss.NewStyle().Width(mainWidth)
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "│"}).
		BorderForeground(lipgloss.Color("238")).
		PaddingLeft(1)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainStyle.Render(mainContent),
		sidebarStyle.Render(sidebar),
	)
}

func (m Model) renderMainContent() string {
	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	mainWidth := m.mainContentWidth()

	switch m.state {
	case stateLoading:
		for _, ex := range m.history {
			b.WriteString(coachLabelStyle.Render("Coach") + "\n")
			b.WriteString(renderMarkdown(ex.question, mainWidth-4) + "\n")
			b.WriteString(userLabelStyle.Render("You") + "\n")
			b.WriteString(userStyle.Render(wordWrap(ex.answer, mainWidth-4)) + "\n\n")
		}
		b.WriteString("\n")
		b.WriteString(m.spinner.View() + " " + loadingStyle.Render("Thinking..."))

	case stateDrilling:
		for _, ex := range m.history {
			b.WriteString(coachLabelStyle.Render("Coach") + "\n")
			b.WriteString(renderMarkdown(ex.question, mainWidth-4) + "\n")
			b.WriteString(userLabelStyle.Render("You") + "\n")
			b.WriteString(userStyle.Render(wordWrap(ex.answer, mainWidth-4)) + "\n\n")
		}

		if m.lastResp != nil {
			b.WriteString(coachLabelStyle.Render("Coach") + "\n")
			b.WriteString(renderMarkdown(m.lastResp.Text, mainWidth-4) + "\n")
		}

		b.WriteString(userLabelStyle.Render("You"))
		if m.recording {
			recStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
			b.WriteString("  " + recStyle.Render("● REC"))
		} else if m.transcribing {
			transcribeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
			b.WriteString("  " + transcribeStyle.Render("transcribing..."))
		}
		b.WriteString("\n")
		b.WriteString(m.textarea.View() + "\n\n")
		var help string
		if m.recording {
			help = "space stop recording • esc quit"
		} else if m.transcribing {
			help = "transcribing audio..."
		} else if m.voiceEnabled {
			if m.speechProc != nil {
				help = "s skip • space record • enter submit • esc quit"
			} else {
				help = "space record • enter submit • ctrl+c clear • esc quit • tab sidebar"
			}
		} else {
			help = "enter submit • ctrl+c clear • esc quit • tab sidebar"
		}
		b.WriteString(helpStyle.Render(help))

	case stateRating:
		if m.lastResp != nil {
			b.WriteString(coachLabelStyle.Render("Coach") + "\n")
			b.WriteString(renderMarkdown(m.lastResp.Text, mainWidth-4) + "\n")
		}

		b.WriteString(dividerStyle.Render(strings.Repeat("─", min(50, mainWidth-4))) + "\n\n")
		b.WriteString(skillRevealStyle.Render(m.skill.Name) + "  ")
		b.WriteString(domainStyle.Render(m.skill.Domain) + "\n\n")

		if m.llmRating > 0 {
			llmLabel := llmRatingLabel(m.llmRating)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Coach thinks: ") + llmLabel + "\n\n")
		}

		b.WriteString(ratingStyle.Render("How did that go?") + "\n\n")
		b.WriteString("  " + ratingKeyStyle.Render("[1]") + ratingOptionStyle.Render(" Again  "))
		b.WriteString(ratingKeyStyle.Render("[2]") + ratingOptionStyle.Render(" Hard  "))
		b.WriteString(ratingKeyStyle.Render("[3]") + ratingOptionStyle.Render(" Good  "))
		b.WriteString(ratingKeyStyle.Render("[4]") + ratingOptionStyle.Render(" Easy") + "\n\n")
		help := "1-4 rate • c continue • q quit • tab sidebar"
		b.WriteString(helpStyle.Render(help))
	}

	return b.String()
}

func (m Model) renderSidebar() string {
	if m.showDebug {
		return m.renderDebugSidebar()
	}

	var b strings.Builder

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)

	b.WriteString(labelStyle.Render("turn") + "\n")
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.turn)) + "\n\n")

	b.WriteString(labelStyle.Render("today") + "\n")
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.todayCount)) + "\n\n")

	if m.currentStreak > 0 {
		b.WriteString(labelStyle.Render("streak") + "\n")
		streakStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		b.WriteString(streakStyle.Render(fmt.Sprintf("%dd", m.currentStreak)) + "\n\n")
	}

	if len(m.recentRatings) > 0 {
		b.WriteString(labelStyle.Render("recent") + "\n")
		b.WriteString(renderSparkline(m.recentRatings) + "\n")
	}

	return b.String()
}

func (m Model) renderDebugSidebar() string {
	var b strings.Builder
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	b.WriteString(labelStyle.Render("skill id") + "\n")
	b.WriteString(valueStyle.Render(m.skill.ID) + "\n\n")

	b.WriteString(labelStyle.Render("difficulty") + "\n")
	b.WriteString(valueStyle.Render(m.difficulty) + "\n\n")

	b.WriteString(labelStyle.Render("facets") + "\n")
	b.WriteString(valueStyle.Render(wordWrap(strings.Join(m.skill.Facets, ", "), m.sidebarWidth()-4)) + "\n\n")

	b.WriteString(labelStyle.Render("history ctx") + "\n")
	history := m.historyCtx
	if strings.TrimSpace(history) == "" {
		history = "(none)"
	}
	b.WriteString(valueStyle.Render(wordWrap(history, m.sidebarWidth()-4)) + "\n\n")

	b.WriteString(labelStyle.Render("system prompt") + "\n")
	promptPreview := m.systemPrompt
	if len(promptPreview) > 700 {
		promptPreview = promptPreview[:700] + "\n... (truncated)"
	}
	b.WriteString(valueStyle.Render(wordWrap(promptPreview, m.sidebarWidth()-4)) + "\n")

	return b.String()
}

func (m Model) renderWelcome() string {
	var b strings.Builder

	logo := `
  ██████╗  ██████╗ ███╗   ██╗██╗  ██╗
  ██╔══██╗██╔═══██╗████╗  ██║██║ ██╔╝
  ██████╔╝██║   ██║██╔██╗ ██║█████╔╝
  ██╔══██╗██║   ██║██║╚██╗██║██╔═██╗
  ██████╔╝╚██████╔╝██║ ╚████║██║  ██╗
  ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝`

	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n\n")

	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	b.WriteString(tagline.Render("  spaced repetition for technical skills"))
	b.WriteString("\n\n")

	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	b.WriteString(statsStyle.Render(fmt.Sprintf("  sessions: %-5d avg: %-4s streak: %d days", m.totalSessions, formatRating(m.avgRating), m.currentStreak)))
	b.WriteString("\n")
	b.WriteString(statsStyle.Render(fmt.Sprintf("  due now: %-5d due week: %-5d new: %d", m.dueCount, m.dueWeekCount, m.newSkillCount)))
	b.WriteString("\n")
	if len(m.recentRatings) > 0 {
		b.WriteString(statsStyle.Render("  recent: "))
		b.WriteString(renderSparkline(m.recentRatings))
		b.WriteString("\n")
	}
	if len(m.recentRatings) > 0 || len(m.recentSessions) > 0 {
		b.WriteString(helpStyle.Render("  legend: "))
		b.WriteString(ratingGlyph(1) + " again  ")
		b.WriteString(ratingGlyph(2) + " hard  ")
		b.WriteString(ratingGlyph(3) + " good  ")
		b.WriteString(ratingGlyph(4) + " easy")
		b.WriteString("\n")
	}
	if len(m.recentSessions) > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  last drilled: %s", relativeTime(m.recentSessions[0].FinishedAt))))
		b.WriteString("\n")
	}
	if len(m.weakFacets) > 0 {
		b.WriteString(helpStyle.Render("  weak facets: "))
		for i, facet := range m.weakFacets {
			if i > 0 {
				b.WriteString(helpStyle.Render("  •  "))
			}
			b.WriteString(strings.ToLower(facet.Facet))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.domainPickerEnabled() {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("  choose a domain"))
		b.WriteString("\n\n")
		options := []struct {
			key    string
			id     string
			label  string
			sample string
		}{
			{"1", "data-structures", "Data Structures", "ds"},
			{"2", "algorithm-patterns", "Algorithm Patterns", "algo"},
			{"3", "system-design", "System Design", "sys"},
			{"4", "system-design-practical", "System Design Practical", "sysp"},
			{"5", "leetcode-patterns", "LeetCode Patterns", "lc"},
		}
		for _, opt := range options {
			prefix := "  "
			if m.selectedDomain == opt.id {
				prefix = "→ "
			}
			b.WriteString(fmt.Sprintf("%s[%s] %-25s (%s)\n", prefix, opt.key, opt.label, opt.sample))
		}
		b.WriteString("\n")
	} else {
		domainHint := domainShort(m.effectiveDomain())
		if domainHint != "" {
			b.WriteString(domainStyle.Render(fmt.Sprintf("  next up: %s", domainHint)))
			b.WriteString("\n\n")
		}
	}

	if len(m.recentSessions) > 0 {
		b.WriteString(helpStyle.Render("  recent drills"))
		b.WriteString("\n")
		for _, s := range m.recentSessions {
			skillName := s.SkillID
			if skill := skills.Get(s.SkillID); skill != nil {
				skillName = skill.Name
			}
			b.WriteString(fmt.Sprintf("  %s  %-28s %s\n", ratingGlyph(s.Rating), truncateASCII(skillName, 28), formatDate(s.FinishedAt)))
		}
		b.WriteString("\n")
	}

	startStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	b.WriteString(startStyle.Render("  [enter]") + helpStyle.Render(" start drill"))
	b.WriteString("   ")
	b.WriteString(helpStyle.Render("q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderHeader() string {
	domainHint := domainShort(m.skill.Domain)

	header := titleStyle.Render("bonk")
	if domainHint != "" {
		header += "  " + domainStyle.Render(domainHint)
	}

	if m.skill.Domain == "system-design-practical" && m.phase != "" {
		phaseName := phaseNames[m.phase]
		phaseNum := phaseOrder[m.phase]
		if phaseName != "" && phaseNum > 0 {
			phaseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			header += "  " + phaseStyle.Render(fmt.Sprintf("%s (%d/6)", phaseName, phaseNum))
		}
	} else if m.turn > 0 {
		header += "  " + helpStyle.Render(fmt.Sprintf("turn %d", m.turn))
	}

	return header
}

func renderMarkdown(text string, width int) string {
	if width <= 0 {
		width = 60
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width),
	)
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimSpace(out)
}

func renderSparkline(ratings []int) string {
	blocks := []string{"▁", "▃", "▅", "▇"}
	var result string
	for _, r := range ratings {
		idx := r - 1
		if idx < 0 {
			idx = 0
		}
		if idx > 3 {
			idx = 3
		}
		var color string
		switch r {
		case 1:
			color = "210"
		case 2:
			color = "214"
		case 3:
			color = "114"
		case 4:
			color = "212"
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		result += style.Render(blocks[idx])
	}
	return result
}

func llmRatingLabel(rating int) string {
	labels := map[int]struct {
		text  string
		color string
	}{
		1: {"Again", "210"},
		2: {"Hard", "214"},
		3: {"Good", "114"},
		4: {"Easy", "212"},
	}
	if l, ok := labels[rating]; ok {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(l.color)).Render(l.text)
	}
	return ""
}
