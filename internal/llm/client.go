package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
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
	LLMRating    int    // 1-4 rating from LLM, 0 if not provided
	Phase        string // for system-design-practical: requirements, entities, api, dataflow, highlevel, deepdives
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

	// Use interview-style prompt for system design practical
	if skill.Domain == "system-design-practical" {
		return buildSystemDesignPracticalPrompt(skill, historyContext, perf)
	}

	facets := strings.Join(skill.Facets, "\n- ")
	problems := strings.Join(skill.ExampleProblems, "\n- ")
	guide := skills.GetGuide(skill.ID)

	historySection := ""
	if historyContext != "" {
		historySection = fmt.Sprintf(`
## Recent History for This Skill
%s

Use this to avoid repeating questions and to target weak areas.
`, historyContext)
	}

	guideSection := ""
	if guide != "" {
		guideSection = fmt.Sprintf(`
## Reference Guide
%s

Use this guide to inform your questioning and evaluation. Don't read it verbatim, but ensure you probe the key areas.
`, guide)
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
%s%s%s
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
- Provide constructive feedback (see Feedback Guidelines below)
- Mark this as the final exchange

You decide when to end based on their responses. Typically 3-6 exchanges, but go longer if the conversation is productive or they're working through something.

## Feedback Guidelines (for final assessment)
Be specific and honest - no generic praise or sugarcoating.

**Content feedback:**
- What concepts did they nail? What gaps remain?
- Did they miss edge cases, complexity analysis, or trade-offs?
- Were their explanations accurate or did they have misconceptions?

**Delivery feedback** (observe patterns across the session):
- Clarity: Were explanations structured or did they jump around?
- Confidence: Did they commit to answers or hedge with "I think maybe..."?
- Precision: Did they use filler phrases ("basically", "kind of", "sort of") excessively?
- Completeness: Did they trail off or fully finish their thoughts?

**Format:**
✓ Strengths: [1-2 specific things they did well]
✗ To improve: [1-2 specific areas to work on - be direct, not gentle]

Example good feedback: "You correctly identified the O(n) approach but couldn't articulate WHY it works. You also said 'I think' 4 times - commit to your answers more confidently."

Example bad feedback: "Great job overall! You showed solid understanding." (too vague, too positive)

## Output Format
At the END of each response, add a metadata line in this exact format:
[meta: facet=<facet_name>, type=<conceptual|problem>, final=<true|false>, rating=<1-4>]

Where:
- facet: which facet you're testing (use short names like "mechanics", "complexity", "application", etc.)
- type: whether this is a conceptual or problem-based question
- final: true only when you give the final assessment
- rating: your assessment of their understanding (1=poor, 2=shaky, 3=solid, 4=excellent). Include on EVERY exchange, not just final.

## Rules
- Be concise - short questions, short feedback
- Push back on vague answers, but don't lecture
- If they're stuck, give a tiny hint, not the full answer
- When ending, ALWAYS include the correct answer/explanation before the assessment
- YOU decide when to end (set final=true) - typically 3-6 exchanges, but be flexible
- Always include the [meta: ...] line at the end of your response

## Pacing
- You may receive "[System: Turn X/Y - wrap up soon]" hints - use these to pace yourself
- End earlier if they've demonstrated solid understanding
- Don't drag out the session unnecessarily

Start with your first question now.
`, skill.Name, skill.Domain, skill.Description, facets, problems, historySection, guideSection, difficultySection)
}

func buildLCPrompt(skill *skills.Skill, historyContext string, perf *PerformanceContext) string {
	facets := strings.Join(skill.Facets, "\n- ")
	problems := strings.Join(skill.ExampleProblems, "\n- ")
	guide := skills.GetGuide(skill.ID)

	historySection := ""
	if historyContext != "" {
		historySection = fmt.Sprintf(`
## Recent History
%s

Use this to vary the problems you present and focus on areas they struggled with.
`, historyContext)
	}

	guideSection := ""
	if guide != "" {
		guideSection = fmt.Sprintf(`
## Reference Guide
%s

Use this guide to inform your questioning and evaluation.
`, guide)
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
%s%s%s
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
- Provide constructive feedback (see below)
- Mark as final

## Feedback Guidelines (for final assessment)
Be specific and honest - no generic praise or sugarcoating.

**Content feedback:**
- Did they recognize the pattern quickly or struggle to see it?
- Could they explain WHY the pattern works, not just WHAT it is?
- Did they handle complexity analysis and edge cases?

**Delivery feedback** (observe patterns across the session):
- Clarity: Did they think out loud in a structured way?
- Confidence: Did they commit to answers or constantly second-guess?
- Precision: Excessive "I think", "maybe", "kind of"?
- Problem-solving: Did they get stuck and freeze, or work through uncertainty?

**Format:**
✓ Strengths: [1-2 specific things they did well]
✗ To improve: [1-2 specific areas - be direct]

## Important
- Do NOT write code. Focus on strategy and reasoning.
- Keep exchanges focused - one question at a time
- You decide when to end (typically 3-6 exchanges)
- You may receive "[System: Turn X/Y - wrap up soon]" hints - use these to pace yourself

## Output Format
At the END of each response, add:
[meta: facet=<facet>, type=problem, final=<true|false>, rating=<1-4>]

Where rating is your assessment of their understanding (1=poor, 2=shaky, 3=solid, 4=excellent). Include on EVERY exchange.

Start by presenting a problem now.
`, skill.Name, skill.Description, facets, problems, historySection, guideSection, difficultySection)
}

func buildSystemDesignPracticalPrompt(skill *skills.Skill, historyContext string, perf *PerformanceContext) string {
	facets := strings.Join(skill.Facets, "\n- ")
	problems := strings.Join(skill.ExampleProblems, "\n- ")
	guide := skills.GetGuide(skill.ID)

	historySection := ""
	if historyContext != "" {
		historySection = fmt.Sprintf(`
## Recent History for This Skill
%s

Use this to avoid repeating the same design questions.
`, historyContext)
	}

	guideSection := ""
	if guide != "" {
		guideSection = fmt.Sprintf(`
## Reference Guide
%s

Use this guide to inform your questioning and evaluation.
`, guide)
	}

	return fmt.Sprintf(`You are a senior engineer conducting a system design interview. Your job is to guide the candidate through designing a system using the Hello Interview framework.

## System to Design
Name: %s
Description: %s

## Key Areas to Probe
- %s

## Example Problems
- %s
%s%s
## Interview Framework (Hello Interview Style)

Guide the candidate through these phases IN ORDER. Track the current phase in your metadata.

**Phase 1: REQUIREMENTS**
- Start with: "Let's design %s. What are the top 3 functional requirements - what should users be able to do?"
- After functional reqs, probe non-functional: "What about non-functional requirements? Think about scale, latency, consistency..."
- Push them to give CONCRETE numbers: "What latency target? How many concurrent users?"
- If they ask YOU for numbers, turn it back: "What would you target for a production system like this?"
- Keep it to 3 functional + 3-5 non-functional requirements
- Skip capacity estimation unless it directly influences the design

**Phase 2: CORE ENTITIES**
- "What are the core entities in this system?"
- Just the nouns/resources, not full schema yet
- Quick phase: 2-3 minutes worth

**Phase 3: API DESIGN**
- "Let's define the API. Walk me through the main endpoints."
- Default to REST unless they have a reason for something else
- Map endpoints to functional requirements

**Phase 4: DATA FLOW** (optional - skip if not a data-processing system)
- "Walk me through the data flow - how does data move through the system?"
- Only use for systems with complex pipelines

**Phase 5: HIGH-LEVEL DESIGN**
- "Now let's draw the architecture. Start with [first endpoint] - what components do you need?"
- Guide them through boxes and arrows
- Go endpoint by endpoint
- Note areas for deep dives but don't go deep yet

**Phase 6: DEEP DIVES**
- Pick 1-2 areas based on non-functional requirements or interesting bottlenecks
- "Let's dive deeper into [X]. How would you handle [scaling/consistency/latency]?"
- This is where the interesting system design discussion happens

## Transition Style
- Transition naturally between phases: "Good, let's move on to the API design" not "Phase 3 starting"
- If candidate jumps ahead, gently guide back: "Let's nail down requirements first before jumping to architecture"
- Be flexible but ensure all phases get covered

## Output Format
At the END of each response, add:
[meta: facet=<facet>, type=interview, final=<true|false>, rating=<1-4>, phase=<phase>]

Where:
- facet: area being probed (requirements, api-design, scalability, etc.)
- type: always "interview" for this domain
- final: true only when giving final assessment
- rating: 1=poor, 2=shaky, 3=solid, 4=excellent
- phase: requirements, entities, api, dataflow, highlevel, or deepdives

## Rules
- Act like a real interviewer - conversational but probing
- NEVER give away answers. If they ask "what should the target be?", turn it back: "What do you think is reasonable? What would users expect?"
- If they offer to list something out ("should I list them?"), say YES and let them do it
- Push back on vague answers: "Can you be more specific about..." or "Can you put a number on that?"
- Only provide hints if they're truly stuck after you've pushed them to think
- When they give good answers, acknowledge briefly and move on - don't repeat their points back at length
- Keep each exchange focused
- This is a FULL interview simulation - take your time through ALL 6 phases

## Feedback Guidelines (for final assessment)
At the end, give detailed constructive feedback. Be specific and honest - no generic praise.

**Technical feedback by phase:**
- Requirements: Did they cover functional AND non-functional? Concrete numbers?
- Entities/API: Clean design? RESTful? Matched requirements?
- High-level: Reasonable architecture? Major components identified?
- Deep dives: Could they go deep on scalability/consistency/edge cases?

**Interview skills feedback:**
- Structure: Did they follow a clear framework or jump around randomly?
- Communication: Did they explain their thinking or just state conclusions?
- Collaboration: Did they ask clarifying questions? Respond well to pushback?
- Confidence: Were they decisive or constantly hedging ("maybe", "I think")?
- Time management: Did they get stuck on one phase or pace themselves?

**Delivery patterns to note:**
- Filler words/phrases ("basically", "kind of", "you know")
- Trailing off mid-thought vs completing ideas
- Saying "I don't know" vs working through uncertainty

**Format:**
✓ Strengths: [2-3 specific things - what would impress in a real interview]
✗ To improve: [2-3 specific areas - be direct about what would hurt them in a real interview]
→ Focus area: [One concrete thing to practice next time]

Do NOT sugarcoat. If they would fail this interview, say so and explain why.

## Pacing
- You may receive "[System: Turn X/Y - wrap up soon]" hints - use these to pace yourself
- If you're behind, you can combine or skip less critical phases (e.g., skip Data Flow for non-pipeline systems)
- If they're doing well and time is short, move to deep dives faster
- Don't rush the deep dives - that's where the interesting discussion happens

Start the interview now.
`, skill.Name, skill.Description, facets, problems, historySection, guideSection, skill.Name)
}

// Regex handles rating=3, rating=, or no rating at all
var metaRegex = regexp.MustCompile(`\[meta:\s*facet=([^,]+),\s*type=([^,]+),\s*final=([^,\]]+)(?:,\s*rating=([^,\]]*))?(?:,\s*phase=([^\]]+))?\]`)

func parseResponse(text string) *Response {
	resp := &Response{Text: text}

	match := metaRegex.FindStringSubmatch(text)
	if match != nil {
		resp.Facet = strings.TrimSpace(match[1])
		resp.QuestionType = strings.ToLower(strings.TrimSpace(match[2]))
		resp.IsFinal = strings.ToLower(strings.TrimSpace(match[3])) == "true"

		// Parse optional rating (1-4) - handles "3", "", or missing
		if len(match) > 4 {
			ratingStr := strings.TrimSpace(match[4])
			if ratingStr != "" {
				if rating, err := strconv.Atoi(ratingStr); err == nil && rating >= 1 && rating <= 4 {
					resp.LLMRating = rating
				}
			}
		}

		// Parse optional phase (for system-design-practical)
		if len(match) > 5 && match[5] != "" {
			resp.Phase = strings.ToLower(strings.TrimSpace(match[5]))
		}

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
	turn         int
	maxTurns     int
	domain       string
}

func (c *Conversation) SystemPrompt() string {
	return c.systemPrompt
}

func NewConversation(skill *skills.Skill, historyContext string, perf *PerformanceContext, maxTurns int) *Conversation {
	return &Conversation{
		systemPrompt: BuildSystemPrompt(skill, historyContext, perf),
		messages:     []message{{Role: "user", Content: "Start the drill."}},
		turn:         0,
		maxTurns:     maxTurns,
		domain:       skill.Domain,
	}
}

func (c *Conversation) Send(userMessage string) (*Response, error) {
	c.turn++

	if userMessage != "" {
		// Add pacing hint when getting close to max turns
		msgWithHint := userMessage
		if c.maxTurns > 0 {
			remaining := c.maxTurns - c.turn
			if remaining <= 3 && remaining > 0 {
				msgWithHint = fmt.Sprintf("%s\n\n[System: Turn %d/%d - wrap up soon if possible]", userMessage, c.turn, c.maxTurns)
			} else if remaining <= 0 {
				msgWithHint = fmt.Sprintf("%s\n\n[System: Turn %d/%d - please give final assessment now]", userMessage, c.turn, c.maxTurns)
			}
		}
		c.messages = append(c.messages, message{Role: "user", Content: msgWithHint})
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

// ExchangeData represents a single exchange for feedback analysis
type ExchangeData struct {
	Question string
	Answer   string
}

// GetSessionFeedback analyzes a session transcript and provides detailed feedback
func GetSessionFeedback(skillID string, exchanges []ExchangeData) (string, error) {
	if len(exchanges) == 0 {
		return "", fmt.Errorf("no exchanges to analyze")
	}

	// Build transcript
	var transcript strings.Builder
	for i, ex := range exchanges {
		transcript.WriteString(fmt.Sprintf("--- Exchange %d ---\n", i+1))
		transcript.WriteString(fmt.Sprintf("Coach: %s\n\n", ex.Question))
		transcript.WriteString(fmt.Sprintf("User: %s\n\n", ex.Answer))
	}

	skill := skills.Get(skillID)
	skillName := skillID
	if skill != nil {
		skillName = skill.Name
	}

	systemPrompt := fmt.Sprintf(`You are an expert technical interview coach reviewing a practice session transcript.

The user practiced: %s

Your job is to provide brutally honest, constructive feedback. Do NOT be sycophantic or sugarcoat weaknesses.

Analyze the transcript for:

1. **Technical Understanding**
   - Did they demonstrate solid knowledge of the topic?
   - Were there gaps, misconceptions, or errors?
   - Did they handle follow-up questions well?

2. **Communication Quality**
   - Were explanations clear and structured, or did they ramble/jump around?
   - Did they use precise technical language or vague hand-wavy descriptions?
   - Did they fully complete thoughts or trail off?

3. **Confidence & Delivery**
   - Did they commit to answers or constantly hedge ("I think maybe...", "kind of", "sort of")?
   - How often did they say "I don't know" vs working through uncertainty?
   - Did they ask good clarifying questions?

4. **Patterns to Note**
   - Filler words/phrases ("basically", "you know", "like")
   - Repetitive language
   - Signs of nervousness or lack of preparation

FORMAT YOUR RESPONSE AS:

## Strengths
[2-3 specific things they did well with concrete examples from the transcript]

## Areas to Improve
[2-3 specific weaknesses - be direct, cite examples from the transcript]

## Delivery Observations
[Notes on communication style, confidence, verbal patterns]

## Overall Assessment
[One paragraph honest assessment - would this pass an actual interview? What's the #1 thing to work on?]

Be specific. Quote the transcript when pointing out issues. If they would fail this interview, say so clearly.`, skillName)

	messages := []message{
		{Role: "user", Content: fmt.Sprintf("Here is the session transcript:\n\n%s\n\nPlease provide detailed feedback.", transcript.String())},
	}

	resp, err := callAPIRaw(systemPrompt, messages, 2048)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// callAPIRaw is like callAPI but returns raw text and allows custom max tokens
func callAPIRaw(systemPrompt string, messages []message, maxTokens int) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	reqBody := apiRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}
