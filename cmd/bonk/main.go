package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"bonk/internal/db"
	"bonk/internal/serve"
	"bonk/internal/skills"
	"bonk/internal/tui"
)

// selectSkill picks the next skill to drill using SM-2 priority:
// 1. Due skills (overdue based on scheduling)
// 2. New skills (never reviewed)
// 3. Random (fallback)
func selectSkill(database *db.DB, domainFilter string) *skills.Skill {
	// Check for due skills first
	dueSkills, _ := database.GetDueSkills()
	for _, due := range dueSkills {
		if s := skills.Get(due.SkillID); s != nil {
			if domainFilter == "" || s.Domain == domainFilter {
				return s
			}
		}
	}

	// Check for new (never reviewed) skills
	var allIDs []string
	for _, s := range skills.List() {
		if domainFilter == "" || s.Domain == domainFilter {
			allIDs = append(allIDs, s.ID)
		}
	}
	newSkills := database.GetNewSkills(allIDs)
	if len(newSkills) > 0 {
		return skills.Get(newSkills[rand.Intn(len(newSkills))])
	}

	// Fallback to random
	var candidates []*skills.Skill
	if domainFilter != "" {
		candidates = skills.ListByDomain(domainFilter)
	} else {
		candidates = skills.List()
	}
	if len(candidates) == 0 {
		return nil
	}
	return candidates[rand.Intn(len(candidates))]
}

func main() {

	rootCmd := &cobra.Command{
		Use:   "bonk [domain]",
		Short: "Socratic drilling for technical skills",
		Long: `Bonk is an LLM-powered spaced repetition system that drills you on
technical concepts like a Socratic coach.

Domains:
  ds    - Data Structures (hash maps, trees, heaps, etc.)
  algo  - Algorithm Patterns (sliding window, binary search, etc.)
  sys   - System Design (load balancing, caching, etc.)
  lc    - LeetCode Patterns (problem-solving archetypes)`,
		Args: cobra.MaximumNArgs(1),
		Run:  runDrill,
	}

	rootCmd.Flags().String("skill", "", "Specific skill ID to drill")
	rootCmd.Flags().Bool("dev", false, "Enable dev mode (debug UI and prompt inspection)")

	// List command
	listCmd := &cobra.Command{
		Use:   "list [domain]",
		Short: "List available skills",
		Args:  cobra.MaximumNArgs(1),
		Run:   runList,
	}
	rootCmd.AddCommand(listCmd)

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show progress statistics",
		Run:   runStats,
	}
	rootCmd.AddCommand(statsCmd)

	// Serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start web terminal for mobile access",
		Run:   runServe,
	}
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(serveCmd)

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info [skill]",
		Short: "Show skill details (facets, example problems)",
		Long: `Show detailed information about a skill including its facets and example problems.

Examples:
  bonk info hash-maps     Show details for hash-maps skill
  bonk info --all         List all skills with full details`,
		Args: cobra.MaximumNArgs(1),
		Run:  runInfo,
	}
	infoCmd.Flags().Bool("all", false, "Show all skills with full details")
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDrill(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Get skill
	var skill *skills.Skill
	var domainFilter string

	skillFlag, _ := cmd.Flags().GetString("skill")
	if skillFlag != "" {
		skill = skills.Get(skillFlag)
		if skill == nil {
			fmt.Fprintf(os.Stderr, "Unknown skill: %s\nUse 'bonk list' to see available skills\n", skillFlag)
			os.Exit(1)
		}
	} else if len(args) > 0 {
		// Domain filter specified
		domain, ok := skills.DomainMap[args[0]]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown domain: %s\nAvailable: ds, algo, sys, lc\n", args[0])
			os.Exit(1)
		}
		domainFilter = domain
		skill = selectSkill(database, domainFilter)
	} else {
		// No filter - use smart selection
		skill = selectSkill(database, "")
	}

	if skill == nil {
		fmt.Fprintf(os.Stderr, "No skills found\n")
		os.Exit(1)
	}

	// Run drill loop
	for {
		devMode, _ := cmd.Flags().GetBool("dev")
		m := tui.NewModel(database, skill, devMode)
		p := tea.NewProgram(m, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if we should continue
		if fm, ok := finalModel.(tui.Model); ok {
			if !fm.ShouldContinue() {
				break
			}
			// Pick next skill using smart selection
			skill = selectSkill(database, domainFilter)
			if skill == nil {
				break
			}
		} else {
			break
		}
	}
}

func runList(cmd *cobra.Command, args []string) {
	var domainFilter string
	if len(args) > 0 {
		domain, ok := skills.DomainMap[args[0]]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown domain: %s\nAvailable: ds, algo, sys, lc\n", args[0])
			os.Exit(1)
		}
		domainFilter = domain
	}

	// Group by domain
	byDomain := make(map[string][]*skills.Skill)
	for _, s := range skills.List() {
		if domainFilter != "" && s.Domain != domainFilter {
			continue
		}
		byDomain[s.Domain] = append(byDomain[s.Domain], s)
	}

	fmt.Println()
	total := 0
	for _, domain := range skills.Domains() {
		domainSkills := byDomain[domain]
		if len(domainSkills) == 0 {
			continue
		}

		if domainFilter == "" {
			fmt.Printf("[%s]\n", domain)
		}
		for _, s := range domainSkills {
			fmt.Printf("  %-25s %s\n", s.ID, s.Name)
			total++
		}
		if domainFilter == "" {
			fmt.Println()
		}
	}

	fmt.Printf("\nTotal: %d skills\n", total)
	fmt.Println("\nUsage: bonk [domain] or bonk --skill <id>")
}

func runStats(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Gather all stats
	total, _ := database.GetTotalSessions()
	dueNow, _ := database.GetDueCount()
	dueWeek, _ := database.GetDueThisWeek()
	newSkills := database.GetNewSkills(skills.ListIDs())
	currentStreak, longestStreak, _ := database.GetStreak()
	avgRating, _, _ := database.GetOverallAvgRating()
	domainStats, _ := database.GetDomainStats(skills.ListIDsByDomain())
	recentSessions, _ := database.GetRecentSessions(5)

	// Header
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                        BONK STATS                           │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Overview row
	fmt.Println("  OVERVIEW")
	fmt.Println("  ────────────────────────────────────────────────────────────")
	fmt.Printf("  Sessions: %-8d  Avg Rating: %-6s  Streak: %d days (best: %d)\n",
		total, formatRating(avgRating), currentStreak, longestStreak)
	fmt.Printf("  Due now:  %-8d  Due this week: %-4d  New skills: %d\n",
		dueNow, dueWeek, len(newSkills))
	fmt.Println()

	// Domain breakdown
	fmt.Println("  DOMAINS")
	fmt.Println("  ────────────────────────────────────────────────────────────")
	domainOrder := []string{"data-structures", "algorithm-patterns", "system-design"}
	domainShort := map[string]string{
		"data-structures":    "DS",
		"algorithm-patterns": "Algo",
		"system-design":      "Sys",
	}
	for _, domain := range domainOrder {
		for _, ds := range domainStats {
			if ds.Domain == domain {
				pct := 0
				if ds.TotalSkills > 0 {
					pct = ds.Practiced * 100 / ds.TotalSkills
				}
				bar := progressBar(pct, 20)
				fmt.Printf("  %-5s %s %3d%% (%d/%d skills, %d sessions, avg %.1f)\n",
					domainShort[domain], bar, pct, ds.Practiced, ds.TotalSkills, ds.SessionCount, ds.AvgRating)
				break
			}
		}
	}
	fmt.Println()

	// Recent sessions
	if len(recentSessions) > 0 {
		fmt.Println("  RECENT")
		fmt.Println("  ────────────────────────────────────────────────────────────")
		for _, s := range recentSessions {
			skill := skills.Get(s.SkillID)
			name := s.SkillID
			if skill != nil {
				name = skill.Name
			}
			date := formatDate(s.FinishedAt)
			fmt.Printf("  %s  %-28s %s\n", ratingEmoji(s.Rating), truncate(name, 28), date)
		}
		fmt.Println()
	}

	// Top skills
	skillStats, _ := database.GetSkillStats(5)
	if len(skillStats) > 0 {
		fmt.Println("  MOST PRACTICED")
		fmt.Println("  ────────────────────────────────────────────────────────────")
		for _, s := range skillStats {
			skill := skills.Get(s.SkillID)
			name := s.SkillID
			if skill != nil {
				name = skill.Name
			}
			fmt.Printf("  %-28s %3d sessions  avg %.1f\n", truncate(name, 28), s.Count, s.AvgRating)
		}
		fmt.Println()
	}
}

func formatRating(r float64) string {
	if r == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", r)
}

func progressBar(pct, width int) string {
	filled := pct * width / 100
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func ratingEmoji(rating int) string {
	switch rating {
	case 1:
		return "🔴"
	case 2:
		return "🟡"
	case 3:
		return "🟢"
	case 4:
		return "⭐"
	default:
		return "  "
	}
}

func formatDate(s string) string {
	// Input: 2024-01-15 14:30:00
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func runServe(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetString("port")
	if err := serve.Run(port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInfo(cmd *cobra.Command, args []string) {
	showAll, _ := cmd.Flags().GetBool("all")

	if showAll {
		// Show all skills with full details
		for _, domain := range skills.Domains() {
			domainSkills := skills.ListByDomain(domain)
			if len(domainSkills) == 0 {
				continue
			}
			fmt.Printf("\n[%s]\n", domain)
			fmt.Println(strings.Repeat("─", 60))
			for _, s := range domainSkills {
				printSkillInfo(s)
			}
		}
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: bonk info <skill> or bonk info --all\n")
		fmt.Fprintf(os.Stderr, "Use 'bonk list' to see available skills\n")
		os.Exit(1)
	}

	skill := skills.Get(args[0])
	if skill == nil {
		fmt.Fprintf(os.Stderr, "Unknown skill: %s\n", args[0])
		fmt.Fprintf(os.Stderr, "Use 'bonk list' to see available skills\n")
		os.Exit(1)
	}

	fmt.Println()
	printSkillInfo(skill)
}

func printSkillInfo(s *skills.Skill) {
	fmt.Printf("%-20s %s\n", "Skill:", s.Name)
	fmt.Printf("%-20s %s\n", "ID:", s.ID)
	fmt.Printf("%-20s %s\n", "Domain:", s.Domain)
	fmt.Printf("%-20s %s\n", "Description:", s.Description)
	fmt.Println()

	if len(s.Facets) > 0 {
		fmt.Println("Facets:")
		for _, f := range s.Facets {
			fmt.Printf("  • %s\n", f)
		}
		fmt.Println()
	}

	if len(s.ExampleProblems) > 0 {
		fmt.Println("Example Problems:")
		for _, p := range s.ExampleProblems {
			fmt.Printf("  • %s\n", p)
		}
		fmt.Println()
	}
}
