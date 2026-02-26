package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"bonk/internal/buildinfo"
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
	rootCmd.Version = buildinfo.Version

	rootCmd.Flags().String("skill", "", "Specific skill ID to drill")
	rootCmd.Flags().BoolP("voice", "v", false, "Enable voice mode (TTS for coach questions)")

	// List command
	listCmd := &cobra.Command{
		Use:   "list [domain]",
		Short: "List available skills",
		Args:  cobra.MaximumNArgs(1),
		Run:   runList,
	}
	rootCmd.AddCommand(listCmd)

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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show bonk version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(buildinfo.Summary())
		},
	}
	rootCmd.AddCommand(versionCmd)

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
	allowDomainPicker := skillFlag == "" && len(args) == 0
	voiceEnabled, _ := cmd.Flags().GetBool("voice")
	for {
		m := tui.NewModel(database, skill, allowDomainPicker, voiceEnabled)
		p := tea.NewProgram(m, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if we should continue
		if fm, ok := finalModel.(tui.Model); ok {
			if pickedDomain := fm.SelectedDomain(); domainFilter == "" && pickedDomain != "" {
				domainFilter = pickedDomain
				allowDomainPicker = false
			}
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
