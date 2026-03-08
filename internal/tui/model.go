package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"bonk/internal/db"
	"bonk/internal/llm"
	"bonk/internal/skills"
	"bonk/internal/voice"
)

// Debug logging for input issues
var debugLog *os.File

func init() {
	if os.Getenv("BONK_DEBUG") != "" {
		home, _ := os.UserHomeDir()
		debugLog, _ = os.OpenFile(
			filepath.Join(home, ".bonk", "debug.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
	}
}

func debugf(format string, args ...interface{}) {
	if debugLog != nil {
		fmt.Fprintf(debugLog, format+"\n", args...)
		debugLog.Sync()
	}
}

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

// Phase name mappings for system-design-practical
var phaseNames = map[string]string{
	"requirements": "Requirements",
	"entities":     "Core Entities",
	"api":          "API Design",
	"dataflow":     "Data Flow",
	"highlevel":    "High-Level Design",
	"deepdives":    "Deep Dives",
}

var phaseOrder = map[string]int{
	"requirements": 1,
	"entities":     2,
	"api":          3,
	"dataflow":     4,
	"highlevel":    5,
	"deepdives":    6,
}

type Model struct {
	db           *db.DB
	skill        *skills.Skill
	conversation *llm.Conversation
	sessionID    string

	state             state
	turn              int
	maxTurns          int
	lastResp          *llm.Response
	phase             string // current phase for system-design-practical
	history           []exchange
	textarea          textarea.Model
	viewport          viewport.Model
	spinner           spinner.Model
	width             int
	height            int
	err               error
	quitting          bool
	continueToNext    bool
	showDebug         bool
	historyCtx        string
	difficulty        string
	systemPrompt      string
	llmRating         int // LLM's rating of user performance (1-4, 0 if not provided)
	selectedDomain    string
	allowDomainPicker bool
	voiceEnabled      bool
	recording         bool
	transcribing      bool
	recordingProc     *voice.Recording
	speechProc        *voice.SpeechProcess

	// Welcome screen stats
	totalSessions  int
	currentStreak  int
	dueCount       int
	dueWeekCount   int
	newSkillCount  int
	avgRating      float64
	todayCount     int
	recentRatings  []int
	recentSessions []db.RecentSession
	weakFacets     []db.FacetStats
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

type recordingStartedMsg struct {
	rec *voice.Recording
	err error
}

type transcriptionMsg struct {
	text string
	err  error
}

func NewModel(database *db.DB, skill *skills.Skill, allowDomainPicker bool, voiceEnabled bool) Model {
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
	dueWeekCount, _ := database.GetDueThisWeek()
	newSkillCount := len(database.GetNewSkills(skills.ListIDs()))
	avgRating, _, _ := database.GetOverallAvgRating()
	todayCount, _ := database.GetTodaySessionCount()
	recentRatings, _ := database.GetRecentRatings(10)
	recentSessions, _ := database.GetRecentSessions(5)
	weakFacets, _ := database.GetWeakFacets(2)

	defaultDomain := ""
	if allowDomainPicker && skill != nil {
		defaultDomain = skill.Domain
	}

	return Model{
		db:                database,
		skill:             skill,
		state:             stateWelcome,
		turn:              0,
		maxTurns:          20, // Default; overridden per-domain in startDrill
		showDebug:         false,
		allowDomainPicker: allowDomainPicker,
		voiceEnabled:      voiceEnabled,
		history:           []exchange{},
		textarea:          ta,
		viewport:          vp,
		spinner:           sp,
		totalSessions:     totalSessions,
		currentStreak:     currentStreak,
		dueCount:          dueCount,
		dueWeekCount:      dueWeekCount,
		newSkillCount:     newSkillCount,
		avgRating:         avgRating,
		todayCount:        todayCount,
		recentRatings:     recentRatings,
		recentSessions:    recentSessions,
		weakFacets:        weakFacets,
		selectedDomain:    defaultDomain,
	}
}

func (m Model) Init() tea.Cmd {
	debugf("Init: state=%d allowDomainPicker=%v skill=%s", m.state, m.allowDomainPicker, m.skill.ID)
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

func (m Model) startRecording() tea.Cmd {
	return func() tea.Msg {
		rec, err := voice.StartRecording()
		return recordingStartedMsg{rec: rec, err: err}
	}
}

func (m Model) stopAndTranscribe() tea.Cmd {
	rec := m.recordingProc
	return func() tea.Msg {
		audioPath, err := rec.Stop()
		if err != nil {
			return transcriptionMsg{err: err}
		}
		text, err := voice.Transcribe(audioPath)
		os.Remove(audioPath) // cleanup temp file
		return transcriptionMsg{text: text, err: err}
	}
}

func (m *Model) startDrill() tea.Cmd {
	debugf("startDrill: beginning, skill=%s selectedDomain=%s", m.skill.ID, m.selectedDomain)
	if m.domainPickerEnabled() && m.selectedDomain != "" {
		if s := pickRandomSkillFromDomain(m.selectedDomain); s != nil {
			m.skill = s
			debugf("startDrill: switched to skill=%s", m.skill.ID)
		}
	}

	// Set maxTurns based on domain - practical interviews need more exchanges
	if m.skill.Domain == "system-design-practical" {
		m.maxTurns = 40 // Full interview simulation with 6 phases
	}

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

	m.conversation = llm.NewConversation(m.skill, historyCtx, perf, m.maxTurns)
	m.systemPrompt = m.conversation.SystemPrompt()
	m.state = stateLoading
	m.textarea.Focus()

	debugf("startDrill: complete, launching session+coach commands")
	return tea.Batch(
		m.createSession(),
		m.getCoachResponse(""),
	)
}
