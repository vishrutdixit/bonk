package skills

import (
	"embed"
	"strings"
)

//go:embed guides/*.md
var guidesFS embed.FS

// GetGuide returns the guide content for a skill, or empty string if none exists.
func GetGuide(skillID string) string {
	data, err := guidesFS.ReadFile("guides/" + skillID + ".md")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
