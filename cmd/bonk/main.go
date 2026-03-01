package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"bonk/internal/buildinfo"
	"bonk/internal/db"
	"bonk/internal/llm"
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
	rand.Seed(time.Now().UnixNano())

	rootCmd := &cobra.Command{
		Use:   "bonk [domain]",
		Short: "Socratic drilling for technical skills",
		Long: `Bonk is an LLM-powered spaced repetition system that drills you on
technical concepts like a Socratic coach.

Domains:
  ds    - Data Structures (hash maps, trees, heaps, etc.)
  algo  - Algorithm Patterns (sliding window, binary search, etc.)
  sys   - System Design (load balancing, caching, etc.)
  sysp  - System Design Practical (interview simulations)
  lc    - LeetCode Patterns (problem-solving archetypes)`,
		Args: cobra.MaximumNArgs(1),
		Run:  runDrill,
	}
	rootCmd.Version = buildinfo.Version

	rootCmd.Flags().String("skill", "", "Specific skill ID to drill")
	rootCmd.Flags().BoolP("voice", "v", true, "Voice mode (TTS for coach, space to record). Use --voice=false to disable")

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

	// Setup command for voice mode dependencies
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up voice mode dependencies (macOS)",
		Long: `Install dependencies for voice mode:
  - sox (audio recording)
  - whisper-cpp (speech-to-text)
  - whisper model file

Requires Homebrew on macOS.`,
		Run: runSetup,
	}
	rootCmd.AddCommand(setupCmd)

	// Review command - view and get feedback on past sessions
	reviewCmd := &cobra.Command{
		Use:   "review [skill]",
		Short: "Review your last session and get AI feedback",
		Long: `View your last drill session and optionally get detailed AI feedback
on your performance, communication style, and areas to improve.

Examples:
  bonk review              Review your most recent session
  bonk review hash-maps    Review your last hash-maps session
  bonk review --feedback   Get AI feedback on your last session`,
		Args: cobra.MaximumNArgs(1),
		Run:  runReview,
	}
	reviewCmd.Flags().BoolP("feedback", "f", false, "Get AI feedback on the session")
	rootCmd.AddCommand(reviewCmd)

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
			fmt.Fprintf(os.Stderr, "Unknown domain: %s\nAvailable: ds, algo, sys, sysp, lc\n", args[0])
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
			fmt.Fprintf(os.Stderr, "Unknown domain: %s\nAvailable: ds, algo, sys, sysp, lc\n", args[0])
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

func runSetup(cmd *cobra.Command, args []string) {
	if runtime.GOOS != "darwin" {
		fmt.Println("Voice mode is currently only supported on macOS.")
		fmt.Println("TTS uses the 'say' command and STT uses whisper.cpp.")
		return
	}

	fmt.Println("Setting up voice mode for bonk...")
	fmt.Println()

	// Check for Homebrew
	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println("✗ Homebrew not found")
		fmt.Println("  Install from: https://brew.sh")
		os.Exit(1)
	}
	fmt.Println("✓ Homebrew found")

	// Check/install sox
	if _, err := exec.LookPath("sox"); err != nil {
		fmt.Println("  Installing sox...")
		installCmd := exec.Command("brew", "install", "sox")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to install sox: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("✓ sox installed")

	// Check/install whisper-cpp
	if _, err := exec.LookPath("whisper-cli"); err != nil {
		fmt.Println("  Installing whisper-cpp...")
		installCmd := exec.Command("brew", "install", "whisper-cpp")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to install whisper-cpp: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("✓ whisper-cpp installed")

	// Download whisper model
	homeDir, _ := os.UserHomeDir()
	bonkDir := filepath.Join(homeDir, ".bonk")
	modelPath := filepath.Join(bonkDir, "ggml-tiny.en.bin")

	if err := os.MkdirAll(bonkDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to create ~/.bonk directory: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Println("  Downloading whisper model (tiny.en, ~39MB)...")
		curlCmd := exec.Command("curl", "-sSL",
			"https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin",
			"-o", modelPath)
		curlCmd.Stdout = os.Stdout
		curlCmd.Stderr = os.Stderr
		if err := curlCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to download model: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("✓ Whisper model ready")

	fmt.Println("\nVoice mode setup complete!")
	fmt.Println("Run: bonk --voice")
}

func runReview(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Get skill filter if provided
	var skillID string
	if len(args) > 0 {
		skillID = args[0]
		if skills.Get(skillID) == nil {
			fmt.Fprintf(os.Stderr, "Unknown skill: %s\n", skillID)
			os.Exit(1)
		}
	}

	// Get last session
	session, err := database.GetLastSession(skillID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting session: %v\n", err)
		os.Exit(1)
	}
	if session == nil {
		fmt.Println("No completed sessions found.")
		return
	}

	// Get skill info
	skill := skills.Get(session.SkillID)
	skillName := session.SkillID
	if skill != nil {
		skillName = skill.Name
	}

	// Print session info
	fmt.Println()
	fmt.Printf("Session: %s\n", skillName)
	fmt.Printf("Date: %s\n", session.StartedAt[:10])
	fmt.Printf("Rating: %d/4\n", session.Rating)
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))

	// Print exchanges
	for _, ex := range session.Exchanges {
		fmt.Println()
		fmt.Printf("Coach:\n%s\n", ex.Question)
		fmt.Println()
		fmt.Printf("You:\n%s\n", ex.Answer)
		fmt.Println()
		fmt.Println(strings.Repeat("─", 40))
	}

	// Get AI feedback if requested
	wantFeedback, _ := cmd.Flags().GetBool("feedback")
	if wantFeedback {
		fmt.Println()
		fmt.Println("Getting AI feedback...")
		fmt.Println()

		// Convert exchanges to llm format
		exchanges := make([]llm.ExchangeData, len(session.Exchanges))
		for i, ex := range session.Exchanges {
			exchanges[i] = llm.ExchangeData{
				Question: ex.Question,
				Answer:   ex.Answer,
			}
		}

		feedback, err := llm.GetSessionFeedback(session.SkillID, exchanges)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting feedback: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(feedback)
	}
}
