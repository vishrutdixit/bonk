package llm

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"bonk/internal/skills"
)

type fakeHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (f fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return f.do(req)
}

func testHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func withHTTPTestEnv(t *testing.T, client httpDoer) {
	t.Helper()
	oldAPIKey := apiKey
	oldModel := model
	oldClient := httpClient
	apiKey = "test-key"
	model = "test-model"
	httpClient = client
	t.Cleanup(func() {
		apiKey = oldAPIKey
		model = oldModel
		httpClient = oldClient
	})
}

func mustSkill(t *testing.T, id string) *skills.Skill {
	t.Helper()
	s := skills.Get(id)
	if s == nil {
		t.Fatalf("skills.Get(%q) = nil", id)
	}
	return s
}

func TestDifficultyLevel(t *testing.T) {
	tests := []struct {
		name string
		perf *PerformanceContext
		want string
	}{
		{name: "nil perf", perf: nil, want: "medium"},
		{name: "few sessions defaults medium", perf: &PerformanceContext{OverallSessions: 2, OverallAvgRating: 1.0}, want: "medium"},
		{name: "hard", perf: &PerformanceContext{OverallSessions: 3, OverallAvgRating: 3.5}, want: "hard"},
		{name: "medium", perf: &PerformanceContext{OverallSessions: 10, OverallAvgRating: 2.8}, want: "medium"},
		{name: "easy", perf: &PerformanceContext{OverallSessions: 6, OverallAvgRating: 2.2}, want: "easy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DifficultyLevel(tt.perf)
			if got != tt.want {
				t.Fatalf("DifficultyLevel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSystemPromptRoutesByDomain(t *testing.T) {
	regular := BuildSystemPrompt(mustSkill(t, "hash-maps"), "", nil)
	if !strings.Contains(regular, "You are a Socratic coding coach.") {
		t.Fatalf("regular prompt missing default coach template: %q", regular)
	}

	lc := BuildSystemPrompt(mustSkill(t, "sliding-window-hash"), "", nil)
	if !strings.Contains(lc, "You are a mock interview coach.") {
		t.Fatalf("lc prompt missing LC template: %q", lc)
	}
	if !strings.Contains(lc, "Do NOT write code.") {
		t.Fatalf("lc prompt missing LC-specific instruction: %q", lc)
	}

	sysp := BuildSystemPrompt(mustSkill(t, "design-twitter"), "", nil)
	if !strings.Contains(sysp, "You are a senior engineer conducting a system design interview.") {
		t.Fatalf("sysp prompt missing interview template: %q", sysp)
	}
	if !strings.Contains(sysp, "phase=<phase>") {
		t.Fatalf("sysp prompt missing phase metadata instruction: %q", sysp)
	}
}

func TestBuildSystemPromptIncludesHistoryAndDifficulty(t *testing.T) {
	history := "Recent questions:\n- Asked about mechanics: struggled\n"
	perf := &PerformanceContext{OverallSessions: 8, OverallAvgRating: 3.7}
	prompt := BuildSystemPrompt(mustSkill(t, "hash-maps"), history, perf)

	if !strings.Contains(prompt, "## Recent History for This Skill") {
		t.Fatalf("prompt missing history section: %q", prompt)
	}
	if !strings.Contains(prompt, history) {
		t.Fatalf("prompt missing history context text: %q", prompt)
	}
	if !strings.Contains(prompt, "## Difficulty Adjustment") {
		t.Fatalf("prompt missing difficulty section: %q", prompt)
	}
	if !strings.Contains(prompt, "level: hard") {
		t.Fatalf("prompt missing hard level marker: %q", prompt)
	}
}

func TestBuildSystemPromptOmitsDifficultyForFewSessions(t *testing.T) {
	perf := &PerformanceContext{OverallSessions: 2, OverallAvgRating: 4.0}
	prompt := BuildSystemPrompt(mustSkill(t, "hash-maps"), "", perf)
	if strings.Contains(prompt, "## Difficulty Adjustment") {
		t.Fatalf("prompt should omit difficulty section for few sessions: %q", prompt)
	}
}

func TestParseResponseWithFullMetadata(t *testing.T) {
	in := "Question text here\n[meta: facet=complexity, type=problem, final=true, rating=4, phase=highlevel, struggled=true]"
	resp := parseResponse(in)

	if resp.Facet != "complexity" {
		t.Fatalf("Facet = %q, want complexity", resp.Facet)
	}
	if resp.QuestionType != "problem" {
		t.Fatalf("QuestionType = %q, want problem", resp.QuestionType)
	}
	if !resp.IsFinal {
		t.Fatal("IsFinal = false, want true")
	}
	if resp.LLMRating != 4 {
		t.Fatalf("LLMRating = %d, want 4", resp.LLMRating)
	}
	if resp.Phase != "highlevel" {
		t.Fatalf("Phase = %q, want highlevel", resp.Phase)
	}
	if !resp.Struggled {
		t.Fatal("Struggled = false, want true")
	}
	if strings.Contains(resp.Text, "[meta:") {
		t.Fatalf("Text should not include meta line: %q", resp.Text)
	}
	if resp.Assessment != resp.Text {
		t.Fatalf("Assessment should mirror text for final responses; got %q text %q", resp.Assessment, resp.Text)
	}
}

func TestParseResponseHandlesOptionalAndInvalidMetadata(t *testing.T) {
	t.Run("invalid rating ignored", func(t *testing.T) {
		in := "Try this\n[meta: facet=application, type=conceptual, final=false, rating=9, struggled=false]"
		resp := parseResponse(in)
		if resp.LLMRating != 0 {
			t.Fatalf("LLMRating = %d, want 0", resp.LLMRating)
		}
		if resp.Struggled {
			t.Fatal("Struggled = true, want false")
		}
		if resp.IsFinal {
			t.Fatal("IsFinal = true, want false")
		}
	})

	t.Run("no metadata leaves text untouched", func(t *testing.T) {
		in := "plain response without metadata"
		resp := parseResponse(in)
		if resp.Text != in {
			t.Fatalf("Text = %q, want %q", resp.Text, in)
		}
		if resp.Facet != "" || resp.QuestionType != "" || resp.LLMRating != 0 || resp.Phase != "" || resp.Struggled {
			t.Fatalf("unexpected parsed values: %#v", resp)
		}
	})
}

func TestParseResponseTrimsAndLowercasesFields(t *testing.T) {
	in := "Hello\n[meta: facet= mechanics , type= CONCEPTUAL , final= TRUE , rating= 3 , phase= API , struggled= TRUE ]"
	resp := parseResponse(in)
	if resp.Facet != "mechanics" {
		t.Fatalf("Facet = %q, want mechanics", resp.Facet)
	}
	if resp.QuestionType != "conceptual" {
		t.Fatalf("QuestionType = %q, want conceptual", resp.QuestionType)
	}
	if !resp.IsFinal {
		t.Fatal("IsFinal = false, want true")
	}
	if resp.LLMRating != 3 {
		t.Fatalf("LLMRating = %d, want 3", resp.LLMRating)
	}
	if resp.Phase != "api" {
		t.Fatalf("Phase = %q, want api", resp.Phase)
	}
	if !resp.Struggled {
		t.Fatal("Struggled = false, want true")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Setenv("BONK_TEST_ENV", "")
	if got := getEnvOrDefault("BONK_TEST_ENV", "default"); got != "default" {
		t.Fatalf("getEnvOrDefault empty = %q, want default", got)
	}

	t.Setenv("BONK_TEST_ENV", "set-value")
	if got := getEnvOrDefault("BONK_TEST_ENV", "default"); got != "set-value" {
		t.Fatalf("getEnvOrDefault set = %q, want set-value", got)
	}
}

func TestNewConversationInitialState(t *testing.T) {
	skill := mustSkill(t, "hash-maps")
	c := NewConversation(skill, "history", &PerformanceContext{OverallSessions: 4, OverallAvgRating: 3.0}, 12)
	if c.turn != 0 {
		t.Fatalf("turn = %d, want 0", c.turn)
	}
	if c.maxTurns != 12 {
		t.Fatalf("maxTurns = %d, want 12", c.maxTurns)
	}
	if c.domain != skill.Domain {
		t.Fatalf("domain = %q, want %q", c.domain, skill.Domain)
	}
	if len(c.messages) != 1 || c.messages[0].Role != "user" || c.messages[0].Content != "Start the drill." {
		t.Fatalf("messages = %#v, want one seed user message", c.messages)
	}
	if c.SystemPrompt() == "" {
		t.Fatal("SystemPrompt() should not be empty")
	}
}

func TestGetSessionFeedbackRejectsEmptyExchanges(t *testing.T) {
	_, err := GetSessionFeedback("hash-maps", nil)
	if err == nil {
		t.Fatal("GetSessionFeedback() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "no exchanges") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallAPIParsesResponseAndSetsHeaders(t *testing.T) {
	withHTTPTestEnv(t, fakeHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", req.Method)
			}
			if req.URL.String() != "https://api.anthropic.com/v1/messages" {
				t.Fatalf("url = %s, want anthropic messages endpoint", req.URL.String())
			}
			if req.Header.Get("x-api-key") != "test-key" {
				t.Fatalf("x-api-key header = %q, want test-key", req.Header.Get("x-api-key"))
			}
			if req.Header.Get("anthropic-version") != "2023-06-01" {
				t.Fatalf("anthropic-version = %q, want 2023-06-01", req.Header.Get("anthropic-version"))
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading request body: %v", err)
			}
			bs := string(body)
			if !strings.Contains(bs, `"model":"test-model"`) {
				t.Fatalf("request body missing model: %s", bs)
			}
			if !strings.Contains(bs, `"max_tokens":1024`) {
				t.Fatalf("request body missing default max_tokens: %s", bs)
			}
			if !strings.Contains(bs, `"system":"sys"`) {
				t.Fatalf("request body missing system prompt: %s", bs)
			}

			resp := `{"content":[{"text":"Explain complexity\n[meta: facet=complexity, type=conceptual, final=false, rating=3, struggled=false]"}]}`
			return testHTTPResponse(200, resp), nil
		},
	})

	out, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
	if err != nil {
		t.Fatalf("callAPI() error = %v", err)
	}
	if out.Text != "Explain complexity" {
		t.Fatalf("Text = %q, want parsed response without meta", out.Text)
	}
	if out.Facet != "complexity" || out.QuestionType != "conceptual" || out.LLMRating != 3 || out.Struggled {
		t.Fatalf("parsed response fields unexpected: %#v", out)
	}
}

func TestCallAPIErrorPaths(t *testing.T) {
	t.Run("missing api key", func(t *testing.T) {
		oldAPIKey := apiKey
		apiKey = ""
		t.Cleanup(func() { apiKey = oldAPIKey })

		_, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
		if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY not set") {
			t.Fatalf("callAPI missing key err = %v, want ANTHROPIC_API_KEY not set", err)
		}
	})

	t.Run("http transport error", func(t *testing.T) {
		withHTTPTestEnv(t, fakeHTTPClient{
			do: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("boom")
			},
		})

		_, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
		if err == nil || !strings.Contains(err.Error(), "send request: boom") {
			t.Fatalf("callAPI transport err = %v, want send request: boom", err)
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		withHTTPTestEnv(t, fakeHTTPClient{
			do: func(req *http.Request) (*http.Response, error) {
				return testHTTPResponse(401, `{"error":"unauthorized"}`), nil
			},
		})

		_, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
		if err == nil || !strings.Contains(err.Error(), "API error 401") {
			t.Fatalf("callAPI status err = %v, want API error 401", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		withHTTPTestEnv(t, fakeHTTPClient{
			do: func(req *http.Request) (*http.Response, error) {
				return testHTTPResponse(200, `not-json`), nil
			},
		})

		_, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
		if err == nil || !strings.Contains(err.Error(), "unmarshal response") {
			t.Fatalf("callAPI json err = %v, want unmarshal response error", err)
		}
	})

	t.Run("empty content response", func(t *testing.T) {
		withHTTPTestEnv(t, fakeHTTPClient{
			do: func(req *http.Request) (*http.Response, error) {
				return testHTTPResponse(200, `{"content":[]}`), nil
			},
		})

		_, err := callAPI("sys", []message{{Role: "user", Content: "hello"}})
		if err == nil || !strings.Contains(err.Error(), "empty response from API") {
			t.Fatalf("callAPI empty err = %v, want empty response from API", err)
		}
	})
}

func TestCallAPIRawUsesCustomMaxTokens(t *testing.T) {
	withHTTPTestEnv(t, fakeHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading request body: %v", err)
			}
			bs := string(body)
			if !strings.Contains(bs, `"max_tokens":2048`) {
				t.Fatalf("request body missing custom max_tokens: %s", bs)
			}
			return testHTTPResponse(200, `{"content":[{"text":"raw output"}]}`), nil
		},
	})

	out, err := callAPIRaw("sys", []message{{Role: "user", Content: "hello"}}, 2048)
	if err != nil {
		t.Fatalf("callAPIRaw() error = %v", err)
	}
	if out != "raw output" {
		t.Fatalf("callAPIRaw() = %q, want raw output", out)
	}
}
