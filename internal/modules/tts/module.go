// Package tts provides text-to-speech module for language learning lessons
package tts

import (
	"context"
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
		BaseModule: core.NewBaseModule(),
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

	// Register event handlers
	m.registerEventHandlers()

	return nil
}

// GetType returns the module type
func (m *TTSModule) GetType() string {
	return "tts"
}

// GetPriority returns module priority for loading order
func (m *TTSModule) GetPriority() int {
	return 300 // Load after core systems but before UI
}

// GetRequires returns module dependencies
func (m *TTSModule) GetRequires() []string {
	return []string{"core"}
}

// GetProvides returns services this module provides
func (m *TTSModule) GetProvides() []string {
	return []string{"tts", "text-to-speech"}
}

// registerEventHandlers sets up event listeners
func (m *TTSModule) registerEventHandlers() {
	if manager := m.GetManager(); manager != nil {
		// Listen for lesson events to provide pronunciation help
		manager.RegisterEventHandler("lesson.word.speak", m.handleSpeakWord)
		manager.RegisterEventHandler("lesson.text.speak", m.handleSpeakText)
		manager.RegisterEventHandler("tts.stop", m.handleStopSpeech)
		manager.RegisterEventHandler("tts.configure", m.handleConfigure)
	}
}

// handleSpeakWord handles requests to speak a word
func (m *TTSModule) handleSpeakWord(ctx context.Context, data map[string]interface{}) error {
	word, ok := data["word"].(string)
	if !ok {
		return fmt.Errorf("invalid word parameter")
	}

	// Optional language parameter for voice selection
	if lang, exists := data["language"].(string); exists {
		m.selectVoiceForLanguage(lang)
	}

	return m.engine.SpeakAsync(word, func(err error) {
		if err != nil {
			m.GetManager().EmitEvent("tts.error", map[string]interface{}{
				"error": err.Error(),
				"text":  word,
			})
		} else {
			m.GetManager().EmitEvent("tts.completed", map[string]interface{}{
				"text": word,
			})
		}
	})
}

// handleSpeakText handles requests to speak longer text
func (m *TTSModule) handleSpeakText(ctx context.Context, data map[string]interface{}) error {
	text, ok := data["text"].(string)
	if !ok {
		return fmt.Errorf("invalid text parameter")
	}

	// Optional language parameter
	if lang, exists := data["language"].(string); exists {
		m.selectVoiceForLanguage(lang)
	}

	return m.engine.SpeakAsync(text, func(err error) {
		if err != nil {
			m.GetManager().EmitEvent("tts.error", map[string]interface{}{
				"error": err.Error(),
				"text":  text,
			})
		} else {
			m.GetManager().EmitEvent("tts.completed", map[string]interface{}{
				"text": text,
			})
		}
	})
}

// handleStopSpeech handles requests to stop current speech
func (m *TTSModule) handleStopSpeech(ctx context.Context, data map[string]interface{}) error {
	m.engine.Stop()
	m.GetManager().EmitEvent("tts.stopped", map[string]interface{}{})
	return nil
}

// handleConfigure handles TTS configuration changes
func (m *TTSModule) handleConfigure(ctx context.Context, data map[string]interface{}) error {
	if voice, exists := data["voice"].(string); exists {
		if err := m.engine.SetVoice(voice); err != nil {
			return fmt.Errorf("failed to set voice: %v", err)
		}
	}

	if rate, exists := data["rate"].(int); exists {
		m.engine.SetRate(rate)
	}

	if volume, exists := data["volume"].(float64); exists {
		m.engine.SetVolume(volume)
	}

	return nil
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
