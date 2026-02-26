package buildinfo

import "fmt"

// Populated via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Summary() string {
	return fmt.Sprintf("bonk %s (commit %s, built %s)", Version, Commit, Date)
}
