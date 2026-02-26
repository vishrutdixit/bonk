package voice

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Speech rate in words per minute (default ~175, fast ~250, very fast ~350)
const speechRate = "280"

// SpeechProcess represents an in-progress TTS that can be stopped.
type SpeechProcess struct {
	cmd *exec.Cmd
}

// Stop kills the speech process.
func (s *SpeechProcess) Stop() {
	if s != nil && s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
}

// Speak uses macOS `say` command to speak text asynchronously.
// Returns a SpeechProcess that can be stopped, or nil if nothing to speak.
func Speak(text string) *SpeechProcess {
	clean := StripMarkdown(text)
	if clean == "" {
		return nil
	}
	cmd := exec.Command("say", "-r", speechRate, clean)
	if err := cmd.Start(); err != nil {
		return nil
	}
	return &SpeechProcess{cmd: cmd}
}

// StripMarkdown removes markdown formatting for cleaner TTS output.
func StripMarkdown(text string) string {
	// Remove code blocks
	text = regexp.MustCompile(`(?s)\x60\x60\x60.*?\x60\x60\x60`).ReplaceAllString(text, "")
	// Remove inline code
	text = regexp.MustCompile(`\x60[^\x60]+\x60`).ReplaceAllString(text, "")
	// Remove markdown links [text](url) -> text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(text, "$1")
	// Remove emphasis markers
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", " ")
	// Remove headers
	text = regexp.MustCompile(`(?m)^#+\s*`).ReplaceAllString(text, "")
	// Remove bullet points
	text = regexp.MustCompile(`(?m)^[\-\*]\s*`).ReplaceAllString(text, "")
	// Collapse whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// Recording holds state for an in-progress audio recording.
type Recording struct {
	cmd       *exec.Cmd
	AudioPath string
}

// StartRecording begins recording audio using sox.
// Returns a Recording that can be stopped later.
func StartRecording() (*Recording, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("bonk-%d.wav", time.Now().UnixNano()))
	// sox -d: default input device, -q: quiet, -r 16000: sample rate, -c 1: mono
	cmd := exec.Command("sox", "-d", "-q", "-r", "16000", "-c", "1", path)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}
	return &Recording{cmd: cmd, AudioPath: path}, nil
}

// Stop ends the recording and returns the audio file path.
func (r *Recording) Stop() (string, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return "", fmt.Errorf("no recording in progress")
	}
	// Send SIGINT to gracefully stop sox
	r.cmd.Process.Signal(syscall.SIGINT)
	r.cmd.Wait()
	return r.AudioPath, nil
}

// Transcribe runs whisper.cpp on an audio file and returns the transcription.
func Transcribe(audioPath string) (string, error) {
	modelPath := filepath.Join(os.Getenv("HOME"), ".bonk", "ggml-tiny.en.bin")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("whisper model not found at %s - run: curl -L https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin -o %s", modelPath, modelPath)
	}
	cmd := exec.Command("whisper-cli", "-m", modelPath, "-f", audioPath, "--no-timestamps")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("whisper transcription failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
