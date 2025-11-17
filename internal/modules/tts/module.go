// Package tts provides text-to-speech module for language learning lessons
package tts

import (
	"fmt"
	"time"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// TTSModule provides text-to-speech functionality as a core module
type TTSModule struct {
	*core.BaseModule
	engine *Engine
}

// NewTTSModule creates a new TTS module instance
func NewTTSModule() *TTSModule {
	return &TTSModule{
		BaseModule: core.NewBaseModule("tts", "Text-to-Speech Engine"),
		engine:     NewEngine(),
	}
}

// Initialize sets up the TTS module
func (m *TTSModule) Initialize(manager *core.Manager) error {
	m.SetManager(manager)

	// Check if TTS is available on this system
	if !m.engine.IsAvailable() {
		return fmt.Errorf("no text-to-speech backend available on this system")
	}

	return nil
}

// Override BaseModule methods to provide TTS-specific behavior
func (m *TTSModule) Priority() int {
	return 300 // Load after core systems but before UI
}

func (m *TTSModule) Requires() []string {
	return []string{"core"}
}

// SpeakWord speaks a single word, optionally in a specific language
func (m *TTSModule) SpeakWord(word, language string) error {
	if language != "" {
		m.selectVoiceForLanguage(language)
	}
	return m.engine.Speak(word)
}

// SpeakText speaks longer text, optionally in a specific language
func (m *TTSModule) SpeakText(text, language string) error {
	if language != "" {
		m.selectVoiceForLanguage(language)
	}
	return m.engine.Speak(text)
}

// selectVoiceForLanguage attempts to select an appropriate voice for the given language
func (m *TTSModule) selectVoiceForLanguage(language string) {
	voices := m.engine.GetVoices()

	// Try to find a voice that matches the language code
	for _, voice := range voices {
		if voice.Language == language ||
			(len(voice.Language) >= 2 && voice.Language[:2] == language) {
			m.engine.SetVoice(voice.ID)
			return
		}
	}

	// Try partial matches for common language codes
	langCode := language
	if len(language) > 2 {
		langCode = language[:2]
	}

	for _, voice := range voices {
		if len(voice.Language) >= 2 && voice.Language[:2] == langCode {
			m.engine.SetVoice(voice.ID)
			return
		}
	}
}

// Public API methods

// Speak synthesizes speech for text (blocking)
func (m *TTSModule) Speak(text string) error {
	return m.engine.SpeakAndWait(text, 30*time.Second)
}

// SpeakAsync synthesizes speech asynchronously
func (m *TTSModule) SpeakAsync(text string, callback func(error)) error {
	return m.engine.SpeakAsync(text, callback)
}

// Stop stops current speech synthesis
func (m *TTSModule) Stop() {
	m.engine.Stop()
}

// IsSpeaking returns true if currently synthesizing speech
func (m *TTSModule) IsSpeaking() bool {
	return m.engine.IsSpeaking()
}

// GetVoices returns available TTS voices
func (m *TTSModule) GetVoices() []Voice {
	return m.engine.GetVoices()
}

// SetVoice sets the current voice
func (m *TTSModule) SetVoice(voiceID string) error {
	return m.engine.SetVoice(voiceID)
}

// GetVoice returns the current voice ID
func (m *TTSModule) GetVoice() string {
	return m.engine.GetVoice()
}

// SetRate sets speaking rate (words per minute)
func (m *TTSModule) SetRate(rate int) {
	m.engine.SetRate(rate)
}

// GetRate returns current speaking rate
func (m *TTSModule) GetRate() int {
	return m.engine.GetRate()
}

// SetVolume sets volume (0.0 to 1.0)
func (m *TTSModule) SetVolume(volume float64) {
	m.engine.SetVolume(volume)
}

// GetVolume returns current volume
func (m *TTSModule) GetVolume() float64 {
	return m.engine.GetVolume()
}

// IsAvailable returns true if TTS is available on this system
func (m *TTSModule) IsAvailable() bool {
	return m.engine.IsAvailable()
}

// GetBackend returns the TTS backend being used
func (m *TTSModule) GetBackend() string {
	return m.engine.GetBackend()
}

// GetInfo returns module information
func (m *TTSModule) GetInfo() map[string]interface{} {
	voices := m.engine.GetVoices()
	voiceList := make([]map[string]interface{}, len(voices))
	for i, voice := range voices {
		voiceList[i] = map[string]interface{}{
			"id":       voice.ID,
			"name":     voice.Name,
			"language": voice.Language,
			"gender":   voice.Gender,
		}
	}

	return map[string]interface{}{
		"type":      "tts",
		"available": m.engine.IsAvailable(),
		"backend":   m.engine.GetBackend(),
		"voices":    voiceList,
		"settings": map[string]interface{}{
			"voice":  m.engine.GetVoice(),
			"rate":   m.engine.GetRate(),
			"volume": m.engine.GetVolume(),
		},
	}
}

// Cleanup performs module cleanup
func (m *TTSModule) Cleanup() error {
	if m.engine != nil {
		m.engine.Stop()
	}
	return nil
}
