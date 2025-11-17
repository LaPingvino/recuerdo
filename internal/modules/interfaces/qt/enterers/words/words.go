// Package words provides a stub word entry system for creating vocabulary lessons
package words

import (
	"fmt"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// WordPair represents a question/answer pair for vocabulary learning
type WordPair struct {
	Question string
	Answer   string
	Known    bool
}

// WordsEntererModule provides basic word entry functionality (stub implementation)
type WordsEntererModule struct {
	*core.BaseModule
	wordPairs []WordPair
}

// NewWordsEntererModule creates a new words enterer module
func NewWordsEntererModule() *WordsEntererModule {
	return &WordsEntererModule{
		BaseModule: core.NewBaseModule("wordsEnterer", "Words Entry System"),
		wordPairs:  make([]WordPair, 0),
	}
}

// Priority returns module loading priority
func (m *WordsEntererModule) Priority() int {
	return 250 // Load after core systems
}

// Requires returns module dependencies
func (m *WordsEntererModule) Requires() []string {
	return []string{"core"}
}

// AddWordPair adds a new word pair to the collection
func (m *WordsEntererModule) AddWordPair(question, answer string) {
	pair := WordPair{
		Question: question,
		Answer:   answer,
		Known:    false,
	}
	m.wordPairs = append(m.wordPairs, pair)
}

// RemoveWordPair removes a word pair at the specified index
func (m *WordsEntererModule) RemoveWordPair(index int) bool {
	if index < 0 || index >= len(m.wordPairs) {
		return false
	}

	// Remove by slicing
	m.wordPairs = append(m.wordPairs[:index], m.wordPairs[index+1:]...)
	return true
}

// GetWordPairs returns all current word pairs
func (m *WordsEntererModule) GetWordPairs() []WordPair {
	return append([]WordPair(nil), m.wordPairs...) // Return copy
}

// UpdateWordPair updates a word pair at the specified index
func (m *WordsEntererModule) UpdateWordPair(index int, question, answer string) bool {
	if index < 0 || index >= len(m.wordPairs) {
		return false
	}

	m.wordPairs[index].Question = question
	m.wordPairs[index].Answer = answer
	return true
}

// Clear removes all word pairs
func (m *WordsEntererModule) Clear() {
	m.wordPairs = m.wordPairs[:0]
}

// Count returns the number of word pairs
func (m *WordsEntererModule) Count() int {
	return len(m.wordPairs)
}

// CreateLesson creates a lesson from the current word pairs
func (m *WordsEntererModule) CreateLesson(title string) *lesson.Lesson {
	// Create a basic lesson structure
	// This is a stub - in a full implementation, this would create
	// a proper lesson.Lesson with all the vocabulary data

	wordList := lesson.NewWordList()
	wordList.Title = title

	// Convert word pairs to WordItems (stub implementation)
	for i, pair := range m.wordPairs {
		item := lesson.WordItem{
			ID:        i + 1,
			Questions: []string{pair.Question},
			Answers:   []string{pair.Answer},
		}
		wordList.Items = append(wordList.Items, item)
	}

	lessonData := &lesson.Lesson{
		Data: lesson.LessonData{
			List:      *wordList,
			Resources: make(map[string]interface{}),
		},
		DataType: "words",
	}

	return lessonData
}

// LoadFromText loads word pairs from plain text (stub implementation)
func (m *WordsEntererModule) LoadFromText(text string) error {
	// TODO: Parse text input into word pairs
	// This would typically handle formats like:
	// question = answer
	// question\tanswer
	// question,answer

	// For now, just return success
	return nil
}

// ExportToText exports word pairs as plain text
func (m *WordsEntererModule) ExportToText() string {
	var result string
	for i, pair := range m.wordPairs {
		if i > 0 {
			result += "\n"
		}
		result += pair.Question + " = " + pair.Answer
	}
	return result
}

// GetInfo returns module information
func (m *WordsEntererModule) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":        "wordsEnterer",
		"name":        "Words Entry System",
		"description": "Basic word pair entry for vocabulary lessons",
		"features": []string{
			"Add/remove word pairs",
			"Create vocabulary lessons",
			"Import/export text format",
		},
		"status":    "stub",
		"wordCount": len(m.wordPairs),
	}
}

// Validate checks if the current word pairs are valid for creating a lesson
func (m *WordsEntererModule) Validate() []string {
	var errors []string

	if len(m.wordPairs) == 0 {
		errors = append(errors, "No word pairs defined")
		return errors
	}

	for i, pair := range m.wordPairs {
		if pair.Question == "" {
			errors = append(errors, fmt.Sprintf("Word pair %d has empty question", i+1))
		}
		if pair.Answer == "" {
			errors = append(errors, fmt.Sprintf("Word pair %d has empty answer", i+1))
		}
	}

	return errors
}
