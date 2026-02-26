package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"bonk/internal/db"
	"bonk/internal/llm"
	"bonk/internal/skills"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	domainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	coachStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			PaddingLeft(2)

	coachLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	userLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(2)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	ratingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	skillRevealStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	ratingOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	ratingKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))
)

type state int

const (
	stateWelcome state = iota
	stateDrilling
	stateRating
	stateLoading
)

type Model struct {
	db           *db.DB
	skill        *skills.Skill
	conversation *llm.Conversation
	sessionID    string

	state          state
	turn           int
	maxTurns       int
	lastResp       *llm.Response
	history        []exchange
	textarea       textarea.Model
	viewport       viewport.Model
	spinner        spinner.Model
	width          int
	height         int
	err            error
	quitting       bool
	continueToNext bool
	devMode        bool
	showDebug      bool
	historyCtx     string
	difficulty     string
	systemPrompt   string
	llmRating      int // LLM's rating of user performance (1-4, 0 if not provided)

	// Welcome screen stats
	totalSessions int
	currentStreak int
	dueCount      int
	todayCount    int
	recentRatings []int
}

type exchange struct {
	question string
	answer   string
}

// Messages
type coachResponseMsg struct {
	resp *llm.Response
	err  error
}

type sessionCreatedMsg struct {
	sessionID string
	err       error
}

func NewModel(database *db.DB, skill *skills.Skill, devMode bool) Model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.CharLimit = 2000
	ta.SetWidth(60)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.Prompt = "  "

	vp := viewport.New(60, 10)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = loadingStyle

	// Fetch welcome stats
	totalSessions, _ := database.GetTotalSessions()
	currentStreak, _, _ := database.GetStreak()
	dueCount, _ := database.GetDueCount()
	todayCount, _ := database.GetTodaySessionCount()
	recentRatings, _ := database.GetRecentRatings(10)

	return Model{
		db:            database,
		skill:         skill,
		state:         stateWelcome,
		turn:          0,
		maxTurns:      10,
		devMode:       devMode,
		history:       []exchange{},
		textarea:      ta,
		viewport:      vp,
		spinner:       sp,
		totalSessions: totalSessions,
		currentStreak: currentStreak,
		dueCount:      dueCount,
		todayCount:    todayCount,
		recentRatings: recentRatings,
	}
}

func (m Model) Init() tea.Cmd {
	// Just start the spinner - session starts when user presses enter
	return m.spinner.Tick
}

func (m Model) createSession() tea.Cmd {
	return func() tea.Msg {
		id, err := m.db.CreateSession(m.skill.ID)
		return sessionCreatedMsg{sessionID: id, err: err}
	}
}

func (m Model) getCoachResponse(userMsg string) tea.Cmd {
	conv := m.conversation
	return func() tea.Msg {
		resp, err := conv.Send(userMsg)
		return coachResponseMsg{resp: resp, err: err}
	}
}

func (m *Model) startDrill() tea.Cmd {
	// Initialize conversation
	historyCtx, _ := m.db.GetHistoryContext(m.skill.ID, 5)

	var perf *llm.PerformanceContext
	skillAvg, skillCount, _ := m.db.GetSkillAvgRating(m.skill.ID)
	overallAvg, overallCount, _ := m.db.GetOverallAvgRating()
	if overallCount > 0 {
		perf = &llm.PerformanceContext{
			SkillAvgRating:   skillAvg,
			SkillSessions:    skillCount,
			OverallAvgRating: overallAvg,
			OverallSessions:  overallCount,
		}
	}
	m.historyCtx = historyCtx
	m.difficulty = llm.DifficultyLevel(perf)

	m.conversation = llm.NewConversation(m.skill, historyCtx, perf)
	m.systemPrompt = m.conversation.SystemPrompt()
	m.state = stateLoading
	m.textarea.Focus()

	return tea.Batch(
		m.createSession(),
		m.getCoachResponse(""),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateWelcome:
			switch msg.String() {
			case "enter", "s", " ":
				return m, m.startDrill()
			case "q":
				m.quitting = true
				return m, tea.Quit
			}
			if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
				m.quitting = true
				return m, tea.Quit
			}

		case stateDrilling:
			if m.devMode && msg.String() == "S" {
				m.showDebug = !m.showDebug
				return m, nil
			}
			switch msg.Type {
			case tea.KeyCtrlC:
				// Clear buffer
				m.textarea.Reset()
				return m, nil
			case tea.KeyEsc:
				m.quitting = true
				return m, tea.Quit
			case tea.KeyEnter:
				if msg.Alt {
					// Alt+Enter for newline
					var cmd tea.Cmd
					m.textarea, cmd = m.textarea.Update(msg)
					return m, cmd
				}
				// Submit answer
				answer := strings.TrimSpace(m.textarea.Value())
				if answer == "" {
					return m, nil
				}

				// Save exchange
				if m.lastResp != nil {
					m.db.SaveExchange(
						m.sessionID,
						m.turn,
						m.lastResp.Text,
						m.lastResp.QuestionType,
						m.lastResp.Facet,
						answer,
						false,
					)
				}

				m.history = append(m.history, exchange{
					question: m.lastResp.Text,
					answer:   answer,
				})
				m.textarea.Reset()
				m.state = stateLoading
				m.turn++

				return m, m.getCoachResponse(answer)
			default:
				// q quits if buffer is empty
				if msg.String() == "q" && strings.TrimSpace(m.textarea.Value()) == "" {
					m.quitting = true
					return m, tea.Quit
				}
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}

		case stateRating:
			if m.devMode && msg.String() == "S" {
				m.showDebug = !m.showDebug
				return m, nil
			}
			switch msg.String() {
			case "1", "2", "3", "4":
				userRating := int(msg.String()[0] - '0')
				assessment := ""
				if m.lastResp != nil {
					assessment = m.lastResp.Assessment
				}
				// Compute combined rating (average of user + LLM, rounded)
				finalRating := userRating
				if m.llmRating > 0 {
					finalRating = (userRating + m.llmRating + 1) / 2 // +1 for rounding
				}
				m.db.FinishSession(m.sessionID, finalRating, assessment)
				m.continueToNext = true
				return m, tea.Quit
			case "c":
				// Continue exploring - go back to drilling state
				m.state = stateDrilling
				m.textarea.Focus()
				return m, nil
			case "q", "esc":
				m.quitting = true
				return m, tea.Quit
			}
			if msg.Type == tea.KeyEsc {
				m.quitting = true
				return m, tea.Quit
			}

		case stateLoading:
			if m.devMode && msg.String() == "S" {
				m.showDebug = !m.showDebug
				return m, nil
			}
			if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC || msg.String() == "q" {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 15

	case sessionCreatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.sessionID = msg.sessionID

	case coachResponseMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.lastResp = msg.resp
		m.turn++

		if msg.resp.IsFinal || m.turn > m.maxTurns {
			m.state = stateRating
			m.llmRating = msg.resp.LLMRating
		} else {
			m.state = stateDrilling
		}

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}

	switch m.state {
	case stateWelcome:
		return m.renderWelcome()
	default:
		return m.renderWithSidebar()
	}
}

func (m Model) renderWithSidebar() string {
	mainContent := m.renderMainContent()
	sidebar := m.renderSidebar()

	// Calculate widths
	sidebarWidth := m.sidebarWidth()
	mainWidth := m.width - sidebarWidth - 3 // 3 for separator
	if mainWidth < 40 {
		mainWidth = 40
	}

	// Style the panels
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

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	mainWidth := m.width - m.sidebarWidth() - 4
	if mainWidth < 40 {
		mainWidth = 40
	}

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

		b.WriteString(userLabelStyle.Render("You") + "\n")
		b.WriteString(m.textarea.View() + "\n\n")
		help := "enter submit • ctrl+c clear • esc quit"
		if m.devMode {
			help += " • S debug"
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

		// Show LLM's rating if available
		if m.llmRating > 0 {
			llmLabel := llmRatingLabel(m.llmRating)
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Coach thinks: ") + llmLabel + "\n\n")
		}

		b.WriteString(ratingStyle.Render("How did that go?") + "\n\n")
		b.WriteString("  " + ratingKeyStyle.Render("[1]") + ratingOptionStyle.Render(" Again  "))
		b.WriteString(ratingKeyStyle.Render("[2]") + ratingOptionStyle.Render(" Hard  "))
		b.WriteString(ratingKeyStyle.Render("[3]") + ratingOptionStyle.Render(" Good  "))
		b.WriteString(ratingKeyStyle.Render("[4]") + ratingOptionStyle.Render(" Easy") + "\n\n")
		help := "1-4 rate • c continue • q quit"
		if m.devMode {
			help += " • S debug"
		}
		b.WriteString(helpStyle.Render(help))
	}

	return b.String()
}

func (m Model) renderSidebar() string {
	if m.devMode && m.showDebug {
		return m.renderDebugSidebar()
	}

	var b strings.Builder

	// Session info
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

	// Sparkline of recent ratings
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

	b.WriteString(labelStyle.Render("dev mode") + "\n")
	b.WriteString(valueStyle.Render("on") + "\n\n")

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

func (m Model) sidebarWidth() int {
	if m.devMode && m.showDebug {
		return 52
	}
	return 16
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
	// Trim extra newlines glamour adds
	return strings.TrimSpace(out)
}

func renderSparkline(ratings []int) string {
	// Use block characters to show rating levels
	// Rating 1-4 maps to different heights
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
		// Color based on rating
		var color string
		switch r {
		case 1:
			color = "210" // red
		case 2:
			color = "214" // orange
		case 3:
			color = "114" // green
		case 4:
			color = "212" // pink/good
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
		1: {"Again", "210"}, // red
		2: {"Hard", "214"},  // orange
		3: {"Good", "114"},  // green
		4: {"Easy", "212"},  // pink
	}
	if l, ok := labels[rating]; ok {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(l.color)).Render(l.text)
	}
	return ""
}

func (m Model) renderWelcome() string {
	var b strings.Builder

	// ASCII logo
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

	// Tagline
	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	b.WriteString(tagline.Render("  spaced repetition for technical skills"))
	b.WriteString("\n\n")

	// Stats box
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	b.WriteString(statsStyle.Render(fmt.Sprintf("  sessions: %d", m.totalSessions)))
	if m.currentStreak > 0 {
		streakStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		b.WriteString(streakStyle.Render(fmt.Sprintf("   streak: %d days", m.currentStreak)))
	}
	if m.dueCount > 0 {
		dueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
		b.WriteString(dueStyle.Render(fmt.Sprintf("   due: %d", m.dueCount)))
	}
	b.WriteString("\n\n")

	// Domain hint
	domainHint := domainShort(m.skill.Domain)
	if domainHint != "" {
		b.WriteString(domainStyle.Render(fmt.Sprintf("  next up: %s", domainHint)))
		b.WriteString("\n\n")
	}

	// Start prompt
	startStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	b.WriteString(startStyle.Render("  [enter]") + helpStyle.Render(" start drill"))
	b.WriteString("   ")
	b.WriteString(helpStyle.Render("q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderHeader() string {
	// Turn indicator (simplified, no more dots)
	turnInfo := fmt.Sprintf("turn %d", m.turn)
	if m.turn == 0 {
		turnInfo = ""
	}

	// Domain hint
	domainHint := domainShort(m.skill.Domain)

	// Compose header
	header := titleStyle.Render("bonk")
	if domainHint != "" {
		header += "  " + domainStyle.Render(domainHint)
	}
	if turnInfo != "" {
		header += "  " + helpStyle.Render(turnInfo)
	}

	return header
}

func domainShort(domain string) string {
	switch domain {
	case "data-structures":
		return "ds"
	case "algorithm-patterns":
		return "algo"
	case "system-design":
		return "sys"
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

func (m Model) ShouldContinue() bool {
	return m.continueToNext
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
