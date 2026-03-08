package main

import (
	"testing"

	"bonk/internal/db"
)

func TestBuildTranscriptSummaryFindsStrongestWeakestAndFocus(t *testing.T) {
	session := &db.SessionDetail{
		SkillID: "hash-maps",
		Exchanges: []db.Exchange{
			{
				Turn:      1,
				Facet:     "invariants",
				Answer:    "not sure maybe use a map",
				Struggled: true,
			},
			{
				Turn:   2,
				Facet:  "complexity",
				Answer: "Use frequency buckets and a hash map; this gives O(n) time and O(n) space with straightforward updates.",
			},
			{
				Turn:      3,
				Facet:     "invariants",
				Answer:    "idk",
				Struggled: true,
			},
		},
	}

	s := buildTranscriptSummary(session)
	if s.Strongest == nil || s.Strongest.Turn != 2 {
		t.Fatalf("Strongest turn = %#v, want turn 2", s.Strongest)
	}
	if s.Weakest == nil || s.Weakest.Turn != 3 {
		t.Fatalf("Weakest turn = %#v, want turn 3", s.Weakest)
	}
	if s.MissedFacet != "invariants" {
		t.Fatalf("MissedFacet = %q, want invariants", s.MissedFacet)
	}
	if s.NextDrillFocus != "invariants" {
		t.Fatalf("NextDrillFocus = %q, want invariants", s.NextDrillFocus)
	}
}

func TestBuildTranscriptSummaryHandlesNoExchanges(t *testing.T) {
	s := buildTranscriptSummary(&db.SessionDetail{SkillID: "hash-maps"})
	if s.Strongest != nil || s.Weakest != nil || s.MissedFacet != "" || s.NextDrillFocus != "" {
		t.Fatalf("summary = %#v, want zero-value summary", s)
	}
}

func TestSummarizeTextNormalizesWhitespaceAndTruncates(t *testing.T) {
	text := "  this   answer has   extra   spaces "
	got := summarizeText(text, 18)
	if got != "this answer has..." {
		t.Fatalf("summarizeText() = %q, want %q", got, "this answer has...")
	}
}
