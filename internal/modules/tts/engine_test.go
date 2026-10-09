// Package tts provides text-to-speech functionality for language learning lessons
package tts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	// Check default values
	if engine.GetRate() != 120 {
		t.Errorf("Expected default rate 120, got %d", engine.GetRate())
	}

	if engine.GetVolume() != 1.0 {
		t.Errorf("Expected default volume 1.0, got %f", engine.GetVolume())
	}

	if engine.IsSpeaking() {
		t.Error("New engine should not be speaking")
	}
}

func TestBackendDetection(t *testing.T) {
	engine := NewEngine()
	backend := engine.GetBackend()

	// Backend should be detected (may be "none" if no TTS available)
	if backend == "" {
		t.Error("Backend should not be empty")
	}

	t.Logf("Detected TTS backend: %s", backend)
	t.Logf("TTS available: %v", engine.IsAvailable())
}

func TestVoiceManagement(t *testing.T) {
	engine := NewEngine()
	voices := engine.GetVoices()

	t.Logf("Found %d voices", len(voices))

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping voice tests")
	}

	if len(voices) == 0 {
		t.Skip("No voices available, skipping voice tests")
	}

	// Test voice information
	for i, voice := range voices {
		if i >= 3 { // Limit output for readability
			break
		}
		t.Logf("Voice %d: ID=%s, Name=%s, Language=%s, Gender=%s",
			i, voice.ID, voice.Name, voice.Language, voice.Gender)

		if voice.ID == "" {
			t.Errorf("Voice %d has empty ID", i)
		}
		if voice.Name == "" {
			t.Errorf("Voice %d has empty Name", i)
		}
	}

	// Test voice selection
	originalVoice := engine.GetVoice()
	if originalVoice == "" && len(voices) > 0 {
		t.Error("Engine should have a default voice when voices are available")
	}

	// Try setting a different voice
	if len(voices) > 1 {
		newVoiceID := voices[1].ID
		err := engine.SetVoice(newVoiceID)
		if err != nil {
			t.Errorf("Failed to set voice: %v", err)
		}

		if engine.GetVoice() != newVoiceID {
			t.Errorf("Expected voice %s, got %s", newVoiceID, engine.GetVoice())
		}

		// Test invalid voice
		err = engine.SetVoice("invalid-voice-id")
		if err == nil {
			t.Error("Setting invalid voice should return error")
		}
	}
}

func TestRateAndVolume(t *testing.T) {
	engine := NewEngine()

	// Test rate setting
	testRates := []int{50, 120, 200, 300, 400} // 400 should be clamped to 300
	for _, rate := range testRates {
		engine.SetRate(rate)
		actualRate := engine.GetRate()

		expectedRate := rate
		if rate > 300 {
			expectedRate = 300
		} else if rate < 50 {
			expectedRate = 50
		}

		if actualRate != expectedRate {
			t.Errorf("Rate: expected %d, got %d", expectedRate, actualRate)
		}
	}

	// Test volume setting
	testVolumes := []float64{0.0, 0.5, 1.0, 1.5} // 1.5 should be clamped to 1.0
	for _, volume := range testVolumes {
		engine.SetVolume(volume)
		actualVolume := engine.GetVolume()

		expectedVolume := volume
		if volume > 1.0 {
			expectedVolume = 1.0
		} else if volume < 0.0 {
			expectedVolume = 0.0
		}

		if actualVolume != expectedVolume {
			t.Errorf("Volume: expected %f, got %f", expectedVolume, actualVolume)
		}
	}
}

func TestSpeechSynthesis(t *testing.T) {
	needAudio(t)
	engine := NewEngine()

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping speech tests")
	}

	// Test basic speech
	testText := "Hello, this is a test."
	err := engine.Speak(testText)
	if err != nil {
		t.Errorf("Speech synthesis failed: %v", err)
	}

	// Wait a moment to ensure speech completes
	time.Sleep(100 * time.Millisecond)
}

func TestAsyncSpeech(t *testing.T) {
	needAudio(t)
	engine := NewEngine()

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping async speech tests")
	}

	// Test async speech with callback (on another goroutine: a channel)
	done := make(chan error, 1)
	if err := engine.SpeakAsync("Async test speech.", func(err error) { done <- err }); err != nil {
		t.Errorf("Async speech start failed: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Async speech callback error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Error("Async speech callback timeout")
	}
}

func TestSpeechControl(t *testing.T) {
	needAudio(t)
	engine := NewEngine()

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping speech control tests")
	}

	// Start speech
	testText := "This is a long speech that should be stopped before completion."
	err := engine.SpeakAsync(testText, nil)
	if err != nil {
		t.Errorf("Failed to start speech: %v", err)
	}

	// Check if speaking
	time.Sleep(100 * time.Millisecond) // Give it time to start
	if !engine.IsSpeaking() {
		t.Log("Warning: Speech may have completed too quickly to test stopping")
	}

	// Stop speech
	engine.Stop()

	// Wait and check if stopped
	time.Sleep(200 * time.Millisecond)
	if engine.IsSpeaking() {
		t.Error("Speech should have stopped")
	}
}

func TestSpeakAndWait(t *testing.T) {
	needAudio(t)
	engine := NewEngine()

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping SpeakAndWait test")
	}

	// Test with timeout
	testText := "Quick test."
	err := engine.SpeakAndWait(testText, 5*time.Second)
	if err != nil {
		t.Errorf("SpeakAndWait failed: %v", err)
	}

	// Engine should not be speaking after completion
	if engine.IsSpeaking() {
		t.Error("Engine should not be speaking after SpeakAndWait completes")
	}
}

func TestWindowsRateConversion(t *testing.T) {
	engine := NewEngine()

	testCases := []struct {
		wpm      int
		expected int
	}{
		{50, -10}, // Minimum
		{120, 0},  // Default
		{180, 5},  // Above default
		{300, 10}, // Maximum
		{400, 10}, // Clamped maximum
	}

	for _, tc := range testCases {
		engine.SetRate(tc.wpm)
		result := engine.rateToWindowsRange()
		if result != tc.expected {
			t.Errorf("Windows rate conversion: WPM %d should convert to %d, got %d",
				tc.wpm, tc.expected, result)
		}
	}
}

func TestWaitUntilDone(t *testing.T) {
	needAudio(t)
	engine := NewEngine()

	if !engine.IsAvailable() {
		t.Skip("TTS not available on this system, skipping WaitUntilDone test")
	}

	// Test when not speaking
	err := engine.WaitUntilDone(1 * time.Second)
	if err != nil {
		t.Errorf("WaitUntilDone should not error when not speaking: %v", err)
	}

	// Test timeout
	engine.SpeakAsync("This is a longer text that might take some time to speak completely.", nil)

	err = engine.WaitUntilDone(100 * time.Millisecond) // Very short timeout
	if err == nil {
		t.Log("Note: Speech completed faster than expected timeout")
	} else if err.Error() != "speech synthesis timeout" {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	// Clean up
	engine.Stop()
}

// Benchmark tests
func BenchmarkEngineCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engine := NewEngine()
		_ = engine
	}
}

func BenchmarkVoiceLoading(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engine := NewEngine()
		engine.loadVoices()
	}
}

// needAudio skips tests that play speech out loud unless
// RECUERDO_TEST_AUDIO=1: a test run must not start talking (and CI
// runners have no sound at all).
func needAudio(t *testing.T) {
	t.Helper()
	// a test run must not start talking out loud (on CI there is no
	// sound at all)
	if os.Getenv("RECUERDO_TEST_AUDIO") == "" {
		t.Skip("plays speech: set RECUERDO_TEST_AUDIO=1 to run")
	}
}

// An utterance interrupted by the next does not clear that one's state:
// Stop still stops it.
func TestStopAfterInterruptedUtterance(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "espeak"), []byte("#!/bin/sh\nexec sleep 10\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	e := NewEngine()
	e.backend, e.voiceID = "espeak", "en"
	if err := e.Speak("first"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := e.Speak("second"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond) // the first ends (killed), its cleanup runs
	if !e.IsSpeaking() {
		t.Fatal("the second utterance no longer counts as speaking")
	}
	e.Stop()
	deadline := time.Now().Add(3 * time.Second)
	for e.IsSpeaking() && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if e.IsSpeaking() {
		t.Error("Stop did not stop the second utterance")
	}
}
