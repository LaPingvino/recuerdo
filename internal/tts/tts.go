// Package tts speaks text with the speech the system has: espeak-ng (or
// espeak) on Linux and BSD, say on macOS and the System.Speech voices on
// Windows. It replaces OpenTeacher's textToSpeech modules, which used
// pyttsx for the same voices.
package tts

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"unicode"

	"github.com/LaPingvino/recuerdo/internal/langcode"
)

// ErrNoSpeech is returned when the system has no speech program.
var ErrNoSpeech = errors.New("no speech program found (install espeak-ng)")

// windowsScript speaks the text in RECUERDO_TTS_TEXT (an environment
// variable, so the text is never parsed as PowerShell).
const windowsScript = "Add-Type -AssemblyName System.Speech; " +
	"(New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak($env:RECUERDO_TTS_TEXT)"

// Command is how to speak text in language (a code like "nl", or a name
// like "Dutch"; empty for the default voice) on goos, given the programs
// found by lookPath: the program, its arguments and extra environment.
func Command(goos, text, language string, lookPath func(string) (string, error)) (name string, args, env []string, err error) {
	switch goos {
	case "windows":
		ps, err := lookPath("powershell")
		if err != nil {
			return "", nil, nil, ErrNoSpeech
		}
		return ps, []string{"-NoProfile", "-NonInteractive", "-Command", windowsScript}, []string{"RECUERDO_TTS_TEXT=" + text}, nil
	case "darwin":
		say, err := lookPath("say")
		if err != nil {
			return "", nil, nil, ErrNoSpeech
		}
		return say, []string{"--", text}, nil, nil
	}
	for _, prog := range []string{"espeak-ng", "espeak"} {
		if p, err := lookPath(prog); err == nil {
			args = []string{}
			if v := Voice(language); v != "" {
				args = append(args, "-v", v)
			}
			return p, append(args, "--", text), nil, nil
		}
	}
	return "", nil, nil, ErrNoSpeech
}

// Voice is espeak's voice for a language code or name ("" if unknown).
func Voice(language string) string {
	l := strings.TrimSpace(language)
	if l == "" {
		return ""
	}
	if isCode(l) {
		return strings.ToLower(l)
	}
	return langcode.Guess(l)
}

func isCode(s string) bool {
	if len(s) < 2 || len(s) > 3 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) || r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// Speaker speaks one text at a time: a new text stops the one before.
type Speaker struct {
	GOOS     string
	LookPath func(string) (string, error)
	mu       sync.Mutex
	current  *exec.Cmd
}

// New is a Speaker for this system.
func New(goos string) *Speaker { return &Speaker{GOOS: goos, LookPath: exec.LookPath} }

// Available reports whether the system can speak.
func (s *Speaker) Available() bool {
	_, _, _, err := Command(s.GOOS, "", "", s.LookPath)
	return err == nil
}

// Speak starts speaking text in language and returns without waiting.
func (s *Speaker) Speak(text, language string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	name, args, env, err := Command(s.GOOS, text, language, s.LookPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil && s.current.Process != nil {
		s.current.Process.Kill()
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.current = cmd
	go cmd.Wait()
	return nil
}
