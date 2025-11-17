// Package tts provides text-to-speech functionality for language learning lessons
package tts

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Voice represents a text-to-speech voice
type Voice struct {
	ID       string
	Name     string
	Language string
	Gender   string
}

// Engine provides text-to-speech functionality
type Engine struct {
	*core.BaseModule

	// Configuration
	rate    int     // Words per minute (default: 120)
	volume  float64 // Volume 0.0-1.0 (default: 1.0)
	voiceID string  // Current voice ID

	// Available voices
	voices []Voice

	// Backend selection
	backend string // "espeak", "say" (macOS), "powershell" (Windows)

	// State management
	mu       sync.RWMutex
	speaking bool
	cancel   context.CancelFunc
}

// NewEngine creates a new text-to-speech engine
func NewEngine() *Engine {
	engine := &Engine{
		BaseModule: core.NewBaseModule(),
		rate:       120, // Default speaking rate
		volume:     1.0, // Default volume
		speaking:   false,
	}

	// Detect and initialize the appropriate backend
	engine.initializeBackend()
	engine.loadVoices()

	return engine
}

// initializeBackend detects and configures the TTS backend based on OS
func (e *Engine) initializeBackend() {
	switch runtime.GOOS {
	case "darwin":
		// macOS - use 'say' command
		if _, err := exec.LookPath("say"); err == nil {
			e.backend = "say"
			return
		}
	case "windows":
		// Windows - use PowerShell SAPI
		e.backend = "powershell"
		return
	case "linux":
		// Linux - prefer espeak-ng, fallback to espeak
		if _, err := exec.LookPath("espeak-ng"); err == nil {
			e.backend = "espeak-ng"
			return
		}
		if _, err := exec.LookPath("espeak"); err == nil {
			e.backend = "espeak"
			return
		}
	}

	// No backend available
	e.backend = "none"
}

// loadVoices discovers available voices for the current backend
func (e *Engine) loadVoices() {
	e.voices = []Voice{}

	switch e.backend {
	case "say":
		e.loadMacOSVoices()
	case "espeak", "espeak-ng":
		e.loadEspeakVoices()
	case "powershell":
		e.loadWindowsVoices()
	}

	// Set default voice if available
	if len(e.voices) > 0 {
		e.voiceID = e.voices[0].ID
	}
}

// loadMacOSVoices loads voices from macOS 'say' command
func (e *Engine) loadMacOSVoices() {
	cmd := exec.Command("say", "-v", "?")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse voice line format: "Alex                en_US    # Most people recognize me by my voice."
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			voice := Voice{
				ID:       parts[0],
				Name:     parts[0],
				Language: parts[1],
				Gender:   "unknown",
			}
			e.voices = append(e.voices, voice)
		}
	}
}

// loadEspeakVoices loads voices from espeak/espeak-ng
func (e *Engine) loadEspeakVoices() {
	cmd := exec.Command(e.backend, "--voices")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		// Parse espeak voice format
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			voice := Voice{
				ID:       fields[1],
				Name:     strings.Join(fields[3:], " "),
				Language: fields[1],
				Gender:   "unknown",
			}

			// Detect gender from voice name
			name := strings.ToLower(voice.Name)
			if strings.Contains(name, "female") || strings.Contains(name, "woman") {
				voice.Gender = "female"
			} else if strings.Contains(name, "male") || strings.Contains(name, "man") {
				voice.Gender = "male"
			}

			e.voices = append(e.voices, voice)
		}
	}
}

// loadWindowsVoices loads voices from Windows SAPI
func (e *Engine) loadWindowsVoices() {
	// PowerShell command to list voices
	script := `Add-Type -AssemblyName System.Speech; $synth = New-Object System.Speech.Synthesis.SpeechSynthesizer; $synth.GetInstalledVoices() | ForEach-Object { Write-Output "$($_.VoiceInfo.Name)|$($_.VoiceInfo.Culture)|$($_.VoiceInfo.Gender)" }`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			voice := Voice{
				ID:       parts[0],
				Name:     parts[0],
				Language: parts[1],
				Gender:   strings.ToLower(parts[2]),
			}
			e.voices = append(e.voices, voice)
		}
	}
}

// Speak synthesizes speech for the given text
func (e *Engine) Speak(text string) error {
	return e.SpeakAsync(text, nil)
}

// SpeakAsync synthesizes speech asynchronously with optional callback
func (e *Engine) SpeakAsync(text string, callback func(error)) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.backend == "none" {
		err := fmt.Errorf("no text-to-speech backend available")
		if callback != nil {
			go callback(err)
		}
		return err
	}

	// Cancel any ongoing speech
	if e.speaking && e.cancel != nil {
		e.cancel()
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.speaking = true

	// Start speech in goroutine
	go func() {
		defer func() {
			e.mu.Lock()
			e.speaking = false
			e.cancel = nil
			e.mu.Unlock()
		}()

		err := e.speakWithBackend(ctx, text)
		if callback != nil {
			callback(err)
		}
	}()

	return nil
}

// speakWithBackend performs the actual TTS using the configured backend
func (e *Engine) speakWithBackend(ctx context.Context, text string) error {
	var cmd *exec.Cmd

	switch e.backend {
	case "say":
		args := []string{"-v", e.voiceID, "-r", strconv.Itoa(e.rate)}
		cmd = exec.CommandContext(ctx, "say", append(args, text)...)

	case "espeak", "espeak-ng":
		args := []string{
			"-v", e.voiceID,
			"-s", strconv.Itoa(e.rate),
			"-a", strconv.Itoa(int(e.volume * 100)),
			text,
		}
		cmd = exec.CommandContext(ctx, e.backend, args...)

	case "powershell":
		// Windows SAPI via PowerShell
		script := fmt.Sprintf(`
			Add-Type -AssemblyName System.Speech;
			$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;
			$synth.SelectVoice('%s');
			$synth.Rate = %d;
			$synth.Volume = %d;
			$synth.Speak('%s')
		`, e.voiceID, e.rateToWindowsRange(), int(e.volume*100), strings.ReplaceAll(text, "'", "''"))

		cmd = exec.CommandContext(ctx, "powershell", "-Command", script)

	default:
		return fmt.Errorf("unsupported TTS backend: %s", e.backend)
	}

	// Redirect stderr to avoid cluttering output
	cmd.Stderr = nil

	return cmd.Run()
}

// Stop stops any ongoing speech synthesis
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.speaking && e.cancel != nil {
		e.cancel()
	}
}

// IsSpeaking returns true if speech synthesis is currently active
func (e *Engine) IsSpeaking() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.speaking
}

// GetVoices returns available voices
func (e *Engine) GetVoices() []Voice {
	return append([]Voice(nil), e.voices...) // Return copy
}

// SetVoice sets the current voice by ID
func (e *Engine) SetVoice(voiceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate voice exists
	for _, voice := range e.voices {
		if voice.ID == voiceID {
			e.voiceID = voiceID
			return nil
		}
	}

	return fmt.Errorf("voice not found: %s", voiceID)
}

// GetVoice returns the current voice ID
func (e *Engine) GetVoice() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.voiceID
}

// SetRate sets the speaking rate in words per minute
func (e *Engine) SetRate(rate int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clamp rate to reasonable bounds
	if rate < 50 {
		rate = 50
	} else if rate > 300 {
		rate = 300
	}

	e.rate = rate
}

// GetRate returns the current speaking rate
func (e *Engine) GetRate() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.rate
}

// SetVolume sets the volume (0.0 to 1.0)
func (e *Engine) SetVolume(volume float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clamp volume to valid range
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	e.volume = volume
}

// GetVolume returns the current volume
func (e *Engine) GetVolume() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.volume
}

// IsAvailable returns true if TTS functionality is available
func (e *Engine) IsAvailable() bool {
	return e.backend != "none"
}

// GetBackend returns the current TTS backend name
func (e *Engine) GetBackend() string {
	return e.backend
}

// rateToWindowsRange converts WPM rate to Windows SAPI range (-10 to 10)
func (e *Engine) rateToWindowsRange() int {
	// Map 50-300 WPM to -10 to 10 range
	// 120 WPM (default) maps to 0
	rate := ((e.rate - 120) * 10) / 120
	if rate < -10 {
		rate = -10
	} else if rate > 10 {
		rate = 10
	}
	return rate
}

// WaitUntilDone blocks until speech synthesis completes
func (e *Engine) WaitUntilDone(timeout time.Duration) error {
	start := time.Now()

	for e.IsSpeaking() {
		if timeout > 0 && time.Since(start) > timeout {
			e.Stop()
			return fmt.Errorf("speech synthesis timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// SpeakAndWait synthesizes speech and waits for completion
func (e *Engine) SpeakAndWait(text string, timeout time.Duration) error {
	if err := e.Speak(text); err != nil {
		return err
	}
	return e.WaitUntilDone(timeout)
}
