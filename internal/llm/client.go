package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"bonk/internal/skills"
)

// These can be set at build time via ldflags:
//
//	go build -ldflags "-X bonk/internal/llm.embeddedAPIKey=sk-ant-..."
var (
	embeddedAPIKey = ""
	embeddedModel  = ""
)

var (
	apiKey = getAPIKey()
	model  = getModel()
)

func getAPIKey() string {
	if embeddedAPIKey != "" {
		return embeddedAPIKey
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}

func getModel() string {
	if embeddedModel != "" {
		return embeddedModel
	}
	return getEnvOrDefault("BONK_MODEL", "claude-sonnet-4-20250514")
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type Response struct {
	Text         string
	Facet        string
	QuestionType string // "conceptual" or "problem"
	IsFinal      bool
	Assessment   string
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type apiResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type PerformanceContext struct {
	SkillAvgRating   float64
	SkillSessions    int
	OverallAvgRating float64
	OverallSessions  int
}

func DifficultyLevel(perf *PerformanceContext) string {
	if perf == nil || perf.OverallSessions < 3 {
		return "medium"
	}
	if perf.OverallAvgRating >= 3.5 {
		return "hard"
	}
	if perf.OverallAvgRating >= 2.5 {
		return "medium"
	}
	return "easy"
}

func BuildSystemPrompt(skill *skills.Skill, historyContext string, perf *PerformanceContext) string {
	// Use different prompt for LC domain (problem-solving focused)
	if skill.Domain == "leetcode-patterns" {
		return buildLCPrompt(skill, historyContext, perf)
	}

	facets := strings.Join(skill.Facets, "\n- ")
	problems := strings.Join(skill.ExampleProblems, "\n- ")

	historySection := ""
	if historyContext != "" {
		historySection = fmt.Sprintf(`
## Recent History for This Skill
%s

Use this to avoid repeating questions and to target weak areas.
`, historyContext)
	}

	difficultySection := ""
	if perf != nil && perf.OverallSessions >= 3 {
		level := DifficultyLevel(perf)
		instruction := ""

		if level == "hard" {
			instruction = `The user is performing very well. Challenge them with:
- Edge cases and corner cases
- Subtle variations that trip people up
- "What if..." scenarios that test deeper understanding
- Ask them to compare/contrast with similar techniques
- Time/space optimization questions`
		} else if level == "medium" {
			instruction = "The user is performing adequately. Use standard difficulty questions."
		} else {
			instruction = `The user is struggling. Help them build confidence:
- Start with more direct questions
- Give clearer hints when stuck
- Focus on core concepts before edge cases
- Be encouraging`
		}

		difficultySection = fmt.Sprintf(`
## Difficulty Adjustment
User performance: %.1f avg rating across %d sessions (level: %s)
%s
`, perf.OverallAvgRating, perf.OverallSessions, level, instruction)
	}

	return fmt.Sprintf(`You are a Socratic coding coach. Your job is to drill the user on a specific skill until they demonstrate solid understanding.

## Skill being drilled
Name: %s
Domain: %s
Description: %s

## Facets to probe (different angles of understanding)
- %s

## Example problems that use this skill
- %s
%s%s
## Structure

**Opening:** Ask ONE of these (randomly vary across sessions):
- Conceptual: Ask them to explain the concept, when to use it, or how it works.
- Problem-based: Give them a specific problem and ask them to walk through their approach.

**Middle exchanges:** Based on their answers:
- Probe gaps or push deeper on things they mentioned
- Hit different angles (complexity, edge cases, trade-offs)
- Keep each exchange focused - don't ask multiple questions at once

**Ending:** When you feel they've demonstrated understanding (or clearly need to review):
- Give the correct answer/explanation for anything they got wrong
- Give a 2-3 sentence assessment summarizing what they nailed and what to review
- Mark this as the final exchange

You decide when to end based on their responses. Typically 3-6 exchanges, but go longer if the conversation is productive or they're working through something.

## Output Format
At the END of each response, add a metadata line in this exact format:
[meta: facet=<facet_name>, type=<conceptual|problem>, final=<true|false>]

Where:
- facet: which facet you're testing (use short names like "mechanics", "complexity", "application", etc.)
- type: whether this is a conceptual or problem-based question
- final: true only on exchange 4 when you give the assessment

## Rules
- Be concise - short questions, short feedback
- Push back on vague answers, but don't lecture
- If they're stuck, give a tiny hint, not the full answer
- When ending, ALWAYS include the correct answer/explanation before the assessment
- YOU decide when to end (set final=true) - typically 3-6 exchanges, but be flexible
- Always include the [meta: ...] line at the end of your response

Start with your first question now.
`, skill.Name, skill.Domain, skill.Description, facets, problems, historySection, difficultySection)
}

func buildLCPrompt(skill *skills.Skill, historyContext string, perf *PerformanceContext) string {
	facets := strings.Join(skill.Facets, "\n- ")
	problems := strings.Join(skill.ExampleProblems, "\n- ")

	historySection := ""
	if historyContext != "" {
		historySection = fmt.Sprintf(`
## Recent History
%s

Use this to vary the problems you present and focus on areas they struggled with.
`, historyContext)
	}

	difficultySection := ""
	if perf != nil && perf.OverallSessions >= 3 {
		level := DifficultyLevel(perf)
		if level == "hard" {
			difficultySection = `
## Difficulty: Hard
User is performing well. Challenge them with:
- Harder variations or follow-up constraints
- "What if the input was unsorted?" or "What if we need O(1) space?"
- Ask them to compare this pattern to similar ones`
		} else if level == "easy" {
			difficultySection = `
## Difficulty: Easy
User is struggling. Help them:
- Start with simpler versions of the problem
- Give more guiding questions
- Focus on recognizing the pattern before optimization`
		}
	}

	return fmt.Sprintf(`You are a mock interview coach. Your job is to drill the user on recognizing and applying a specific LeetCode problem pattern.

## Pattern being drilled
Name: %s
Description: %s

## Key facets to probe
- %s

## Example problems using this pattern
- %s
%s%s
## Coaching Approach

**Opening:** Present a problem that uses this pattern. You can:
- Use one of the example problems
- Create a variation with different constraints
- Describe a real-world scenario that maps to this pattern

Ask: "How would you approach this?"

**Middle exchanges:**
- If they identify the pattern, probe WHY this pattern works
- If they're stuck, ask guiding questions about the key insight
- Push on complexity: "What's the time/space complexity? Can we do better?"
- Explore edge cases: "What if the input is empty? What about duplicates?"

**Ending:** When they've demonstrated understanding (or clearly need review):
- Confirm the correct approach if they got it
- Explain what they missed if they struggled
- Give a 2-3 sentence assessment
- Mark as final

## Important
- Do NOT write code. Focus on strategy and reasoning.
- Keep exchanges focused - one question at a time
- You decide when to end (typically 3-6 exchanges)

## Output Format
At the END of each response, add:
[meta: facet=<facet>, type=problem, final=<true|false>]

Start by presenting a problem now.
`, skill.Name, skill.Description, facets, problems, historySection, difficultySection)
}

var metaRegex = regexp.MustCompile(`\[meta:\s*facet=([^,]+),\s*type=([^,]+),\s*final=([^\]]+)\]`)

func parseResponse(text string) *Response {
	resp := &Response{Text: text}

	match := metaRegex.FindStringSubmatch(text)
	if match != nil {
		resp.Facet = strings.TrimSpace(match[1])
		resp.QuestionType = strings.ToLower(strings.TrimSpace(match[2]))
		resp.IsFinal = strings.ToLower(strings.TrimSpace(match[3])) == "true"

		// Remove meta line from display text
		resp.Text = strings.TrimSpace(metaRegex.ReplaceAllString(text, ""))

		if resp.IsFinal {
			resp.Assessment = resp.Text
		}
	}

	return resp
}

type Conversation struct {
	systemPrompt string
	messages     []message
}

func (c *Conversation) SystemPrompt() string {
	return c.systemPrompt
}

func NewConversation(skill *skills.Skill, historyContext string, perf *PerformanceContext) *Conversation {
	return &Conversation{
		systemPrompt: BuildSystemPrompt(skill, historyContext, perf),
		messages:     []message{{Role: "user", Content: "Start the drill."}},
	}
}

func (c *Conversation) Send(userMessage string) (*Response, error) {
	if userMessage != "" {
		c.messages = append(c.messages, message{Role: "user", Content: userMessage})
	}

	resp, err := callAPI(c.systemPrompt, c.messages)
	if err != nil {
		return nil, err
	}

	c.messages = append(c.messages, message{Role: "assistant", Content: resp.Text})
	return resp, nil
}

func callAPI(systemPrompt string, messages []message) (*Response, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	reqBody := apiRequest{
		Model:     model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages:  messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	return parseResponse(apiResp.Content[0].Text), nil
}
