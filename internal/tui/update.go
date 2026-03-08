package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"bonk/internal/voice"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		debugf("KeyMsg: Type=%d String=%q Runes=%v Alt=%v state=%d",
			msg.Type, msg.String(), msg.Runes, msg.Alt, m.state)
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		debugf("WindowSizeMsg: %dx%d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		m.syncLayout()

	case sessionCreatedMsg:
		debugf("sessionCreatedMsg: sessionID=%s err=%v", msg.sessionID, msg.err)
		if msg.err != nil {
			debugf("QUIT: session creation failed: %v", msg.err)
			m.err = msg.err
			return m, tea.Quit
		}
		m.sessionID = msg.sessionID

	case coachResponseMsg:
		debugf("coachResponseMsg: err=%v", msg.err)
		if msg.err != nil {
			debugf("QUIT: coach response failed: %v", msg.err)
			m.err = msg.err
			return m, tea.Quit
		}
		m.lastResp = msg.resp
		m.turn++

		if msg.resp.Phase != "" {
			m.phase = msg.resp.Phase
		}

		if msg.resp.IsFinal || m.turn > m.maxTurns {
			m.state = stateRating
			m.llmRating = msg.resp.LLMRating
		} else {
			m.state = stateDrilling
			if m.voiceEnabled {
				m.speechProc = voice.Speak(msg.resp.Text)
			}
		}

	case recordingStartedMsg:
		if msg.err != nil {
			// Recording failed to start - stay in drilling state
			// TODO: show error to user
			return m, nil
		}
		m.recording = true
		m.recordingProc = msg.rec

	case transcriptionMsg:
		m.recording = false
		m.transcribing = false
		m.recordingProc = nil
		if msg.err == nil && msg.text != "" {
			m.textarea.SetValue(msg.text)
		}
		// TODO: show error if transcription failed

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateWelcome:
		return m.handleWelcomeKey(msg)
	case stateDrilling:
		return m.handleDrillingKey(msg)
	case stateRating:
		return m.handleRatingKey(msg)
	case stateLoading:
		return m.handleLoadingKey(msg)
	default:
		return m, nil
	}
}

func (m Model) maybeToggleDebug(msg tea.KeyMsg) (Model, bool) {
	if msg.Type == tea.KeyTab || msg.Type == tea.KeyShiftTab {
		m.showDebug = !m.showDebug
		m.syncLayout()
		return m, true
	}
	return m, false
}

func (m Model) handleWelcomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.domainPickerEnabled() {
		if msg.Type == tea.KeyUp || msg.String() == "k" {
			m.selectedDomain = cycleDomainSelection(m.selectedDomain, -1)
			return m, nil
		}
		if msg.Type == tea.KeyDown || msg.String() == "j" {
			m.selectedDomain = cycleDomainSelection(m.selectedDomain, 1)
			return m, nil
		}

		switch msg.String() {
		case "1":
			m.selectedDomain = "data-structures"
			return m, nil
		case "2":
			m.selectedDomain = "algorithm-patterns"
			return m, nil
		case "3":
			m.selectedDomain = "system-design"
			return m, nil
		case "4":
			m.selectedDomain = "system-design-practical"
			return m, nil
		case "5":
			m.selectedDomain = "leetcode-patterns"
			return m, nil
		}
	}

	switch msg.String() {
	case "enter", "s", " ":
		return m, m.startDrill()
	case "q":
		debugf("QUIT: matched 'q' key")
		m.quitting = true
		return m, tea.Quit
	}
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		debugf("QUIT: matched Esc/CtrlC, Type=%d", msg.Type)
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleDrillingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if updated, toggled := m.maybeToggleDebug(msg); toggled {
		return updated, nil
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		m.textarea.Reset()
		return m, nil
	case tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyEnter:
		if msg.Alt {
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
		answer := strings.TrimSpace(m.textarea.Value())
		if answer == "" {
			return m, nil
		}
		if m.speechProc != nil {
			m.speechProc.Stop()
			m.speechProc = nil
		}

		if m.lastResp != nil {
			m.db.SaveExchange(
				m.sessionID,
				m.turn,
				m.lastResp.Text,
				m.lastResp.QuestionType,
				m.lastResp.Facet,
				answer,
				m.lastResp.Struggled,
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
		if msg.String() == "q" && strings.TrimSpace(m.textarea.Value()) == "" {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "s" && m.voiceEnabled && m.speechProc != nil {
			m.speechProc.Stop()
			m.speechProc = nil
			return m, nil
		}
		if msg.String() == " " && m.voiceEnabled && strings.TrimSpace(m.textarea.Value()) == "" {
			if m.recording {
				m.recording = false
				m.transcribing = true
				return m, m.stopAndTranscribe()
			}
			if m.speechProc != nil {
				m.speechProc.Stop()
				m.speechProc = nil
			}
			return m, m.startRecording()
		}
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
}

func (m Model) handleRatingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if updated, toggled := m.maybeToggleDebug(msg); toggled {
		return updated, nil
	}

	switch msg.String() {
	case "1", "2", "3", "4":
		userRating := int(msg.String()[0] - '0')
		assessment := ""
		if m.lastResp != nil {
			assessment = m.lastResp.Assessment
		}
		finalRating := userRating
		if m.llmRating > 0 {
			finalRating = (userRating + m.llmRating + 1) / 2 // +1 for rounding
		}
		m.db.FinishSession(m.sessionID, finalRating, assessment)
		m.continueToNext = true
		return m, tea.Quit
	case "c":
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
	return m, nil
}

func (m Model) handleLoadingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if updated, toggled := m.maybeToggleDebug(msg); toggled {
		return updated, nil
	}
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC || msg.String() == "q" {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}
