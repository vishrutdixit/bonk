package serve

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run starts a web terminal server using ttyd
func Run(port string) error {
	// Check if ttyd is installed
	if _, err := exec.LookPath("ttyd"); err != nil {
		fmt.Fprintln(os.Stderr, "ttyd not found. Install with:")
		fmt.Fprintln(os.Stderr, "  brew install ttyd  # macOS")
		fmt.Fprintln(os.Stderr, "  apt install ttyd   # Linux")
		return fmt.Errorf("ttyd not installed")
	}

	// Get the path to our own executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get local IP for phone access
	localIP := getLocalIP()

	// Build ttyd command
	// -W enables writable mode (allows input from browser)
	// -t options set xterm.js terminal options for better mobile experience
	cmd := exec.Command("ttyd",
		"-W",
		"-p", port,
		"-t", "fontSize=32",
		"-t", "cursorBlink=true",
		exe,
	)
	cmd.Env = os.Environ() // Pass through env vars (ANTHROPIC_API_KEY)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Print access URLs
	fmt.Println("bonk web terminal starting...")
	fmt.Println()
	fmt.Printf("  Local:   http://localhost:%s\n", port)
	if localIP != "" {
		fmt.Printf("  Network: http://%s:%s\n", localIP, port)
	}
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start ttyd
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ttyd: %w", err)
	}

	// Wait for signal or process exit
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cmd.Process.Signal(syscall.SIGTERM)
	}()

	return cmd.Wait()
}

// getLocalIP returns the local IP address for LAN access
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
