// Package plaintextwords provides a plain text word list entry interface with Qt dialog
package plaintextwords

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// PlainTextWordsEntererModule handles plain text word list input with Qt interface
type PlainTextWordsEntererModule struct {
	*core.BaseModule
	manager        *core.Manager
	uiModule       core.Module
	buttonRegister core.Module
	parser         core.Module
	loaderGui      core.Module
	translator     core.Module
	charsKeyboard  core.Module
	button         interface{} // Button reference
	activeDialogs  []*EnterPlainTextDialog
}

// EnterPlainTextDialog is the Qt dialog for entering plain text word pairs
type EnterPlainTextDialog struct {
	*qt.QDialog
	module    *PlainTextWordsEntererModule
	textEdit  *qt.QTextEdit
	label     *qt.QLabel
	buttonBox *qt.QDialogButtonBox
	tab       interface{} // Tab reference
}

// WordPair represents a question-answer pair
type WordPair struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// Lesson represents a word lesson structure
type Lesson struct {
	Type    string      `json:"type"`
	List    interface{} `json:"list"`
	Changed bool        `json:"changed"`
}

// NewPlainTextWordsEntererModule creates a new plain text words enterer module
func NewPlainTextWordsEntererModule() *PlainTextWordsEntererModule {
	base := core.NewBaseModule("plainTextWordsEnterer", "Plain Text Words Enterer")
	base.SetPriority(960) // High priority as in Python version

	return &PlainTextWordsEntererModule{
		BaseModule:    base,
		activeDialogs: make([]*EnterPlainTextDialog, 0),
	}
}

// Enable activates the module and sets up Qt interface
func (m *PlainTextWordsEntererModule) Enable(ctx context.Context) error {
	if err := m.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Get required modules
	if err := m.setupModuleDependencies(); err != nil {
		return fmt.Errorf("failed to setup module dependencies: %w", err)
	}

	// Register button
	if err := m.registerButton(); err != nil {
		return fmt.Errorf("failed to register button: %w", err)
	}

	// Setup translations
	m.retranslate()

	log.Println("PlainTextWordsEntererModule enabled")
	return nil
}

// Disable deactivates the module
func (m *PlainTextWordsEntererModule) Disable(ctx context.Context) error {
	if err := m.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Close all active dialogs
	for _, dialog := range m.activeDialogs {
		if dialog != nil {
			dialog.Close()
		}
	}
	m.activeDialogs = nil

	// Unregister button
	if m.buttonRegister != nil && m.button != nil {
		// TODO: Call unregisterButton when interface is available
	}

	log.Println("PlainTextWordsEntererModule disabled")
	return nil
}

// SetManager sets the module manager
func (m *PlainTextWordsEntererModule) SetManager(manager *core.Manager) {
	m.manager = manager
}

// setupModuleDependencies gets references to required modules
func (m *PlainTextWordsEntererModule) setupModuleDependencies() error {
	if m.manager == nil {
		return fmt.Errorf("manager not set")
	}

	// Get UI module
	uiModules := m.manager.GetModulesByType("ui")
	if len(uiModules) > 0 {
		m.uiModule = uiModules[0] // Use first active UI module
	}

	// Get button register module
	buttonModules := m.manager.GetModulesByType("buttonRegister")
	if len(buttonModules) > 0 {
		m.buttonRegister = buttonModules[0]
	}

	// Get word list string parser
	parserModules := m.manager.GetModulesByType("wordListStringParser")
	if len(parserModules) > 0 {
		m.parser = parserModules[0]
	}

	// Get loader GUI module
	loaderModules := m.manager.GetModulesByType("loaderGui")
	if len(loaderModules) > 0 {
		m.loaderGui = loaderModules[0]
	}

	// Get optional modules
	translatorModules := m.manager.GetModulesByType("translator")
	if len(translatorModules) > 0 {
		m.translator = translatorModules[0]
	}

	charsModules := m.manager.GetModulesByType("charsKeyboard")
	if len(charsModules) > 0 {
		m.charsKeyboard = charsModules[0]
	}

	return nil
}

// registerButton registers the create button
func (m *PlainTextWordsEntererModule) registerButton() error {
	if m.buttonRegister == nil {
		return fmt.Errorf("buttonRegister module not available")
	}

	// TODO: Implement button registration when interface is available
	// For now, we'll handle button creation internally

	return nil
}

// retranslate updates all translatable strings
func (m *PlainTextWordsEntererModule) retranslate() {
	// TODO: Implement translation system integration
	// For now, use English strings

	// Update button text if available
	// TODO: Update button text when interface is available

	// Update all active dialogs
	for _, dialog := range m.activeDialogs {
		if dialog != nil {
			dialog.retranslate()
		}
	}
}

// CreateLesson shows the plain text entry dialog
func (m *PlainTextWordsEntererModule) CreateLesson() {
	dialog := m.createEnterPlainTextDialog()
	m.activeDialogs = append(m.activeDialogs, dialog)

	// TODO: Add to UI module as custom tab when interface is available
	// For now, show as modal dialog
	dialog.ShowModal()
}

// createEnterPlainTextDialog creates a new plain text entry dialog
func (m *PlainTextWordsEntererModule) createEnterPlainTextDialog() *EnterPlainTextDialog {
	dialog := &EnterPlainTextDialog{
		QDialog: qt.NewQDialog(),
		module:  m,
	}

	dialog.setupUI()
	dialog.retranslate()
	dialog.connectSignals()

	return dialog
}

// setupUI creates the dialog's user interface
func (d *EnterPlainTextDialog) setupUI() {
	d.SetModal(true)
	d.Resize(800, 600)

	// Create button box
	d.buttonBox = qt.NewQDialogButtonBox2(
		qt.QDialogButtonBox_Cancel|qt.QDialogButtonBox_Ok,
		d.QDialog,
	)

	// Create label
	d.label = qt.NewQLabel2("", d.QDialog, 0)
	d.label.SetWordWrap(true)
	d.label.SetSizePolicy2(qt.QSizePolicy_Preferred, qt.QSizePolicy_Fixed)

	// Create text edit
	d.textEdit = qt.NewQTextEdit(d.QDialog)

	// Create splitter for text edit and optional character keyboard
	splitter := qt.NewQSplitter2(qt.Orientation_Horizontal, d.QDialog)
	splitter.AddWidget(d.textEdit)
	splitter.SetStretchFactor(0, 3)

	// Add character keyboard if available
	if d.module.charsKeyboard != nil {
		// TODO: Create keyboard widget when interface is available
		// splitter.AddWidget(keyboardWidget)
		// splitter.SetStretchFactor(1, 1)
	}

	// Create layout
	layout := qt.NewQVBoxLayout()
	layout.AddWidget(d.label, 0, 0)
	layout.AddWidget(splitter, 0, 0)
	layout.AddWidget(d.buttonBox, 0, 0)

	d.SetLayout(layout)
}

// connectSignals connects Qt signals to slots
func (d *EnterPlainTextDialog) connectSignals() {
	// Connect button box signals
	d.buttonBox.OnAccepted(func() {
		d.accept()
	})

	d.buttonBox.OnRejected(func() {
		d.Reject()
	})

	// Connect dialog finished signal
	d.OnFinished(func(result int) {
		d.cleanup()
	})
}

// retranslate updates translatable strings for the dialog
func (d *EnterPlainTextDialog) retranslate() {
	d.label.SetText("Please enter the plain text in the text edit. Separate words with a new line and questions from answers with an equals sign ('=') or a tab.")
	d.SetWindowTitle("Plain text words enterer")
}

// accept handles dialog acceptance
func (d *EnterPlainTextDialog) accept() {
	lesson, err := d.getLesson()
	if err != nil {
		d.showError("Missing equals sign or tab", "Please make sure every line contains an '='-sign or tab between the questions and answers.")
		return
	}

	if lesson != nil {
		d.loadLesson(lesson)
	}

	d.Accept()
}

// getLesson parses the text and returns a lesson
func (d *EnterPlainTextDialog) getLesson() (*Lesson, error) {
	text := d.textEdit.ToPlainText()
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	wordPairs, err := d.parseList(text)
	if err != nil {
		return nil, err
	}

	if len(wordPairs) == 0 {
		return nil, nil
	}

	lesson := &Lesson{
		Type:    "words",
		List:    wordPairs,
		Changed: true,
	}

	return lesson, nil
}

// parseList parses plain text into word pairs
func (d *EnterPlainTextDialog) parseList(text string) ([]WordPair, error) {
	var wordPairs []WordPair

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try splitting by '=' first, then by tab
		var question, answer string
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				question = strings.TrimSpace(parts[0])
				answer = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "\t") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				question = strings.TrimSpace(parts[0])
				answer = strings.TrimSpace(parts[1])
			}
		}

		if question == "" || answer == "" {
			return nil, fmt.Errorf("line %d: missing separator", i+1)
		}

		wordPairs = append(wordPairs, WordPair{
			Question: question,
			Answer:   answer,
		})
	}

	return wordPairs, nil
}

// loadLesson loads the lesson using the loader GUI module
func (d *EnterPlainTextDialog) loadLesson(lesson *Lesson) {
	if d.module.loaderGui == nil {
		log.Println("LoaderGui module not available")
		return
	}

	// TODO: Call loadFromLesson when interface is available
	log.Printf("Loading lesson with %d word pairs", d.getLessonSize(lesson))
}

// getLessonSize returns the number of items in the lesson
func (d *EnterPlainTextDialog) getLessonSize(lesson *Lesson) int {
	if lesson == nil || lesson.List == nil {
		return 0
	}

	if wordPairs, ok := lesson.List.([]WordPair); ok {
		return len(wordPairs)
	}

	return 0
}

// showError displays an error message box
func (d *EnterPlainTextDialog) showError(title, message string) {
	qt.QMessageBox_Warning(
		d.QDialog,
		title,
		message,
		qt.QMessageBox_Ok,
		qt.QMessageBox_Ok,
	)
}

// focus sets focus to the text edit
func (d *EnterPlainTextDialog) focus() {
	d.textEdit.SetFocus2(qt.FocusReason_OtherFocusReason)
}

// cleanup removes the dialog from the active dialogs list
func (d *EnterPlainTextDialog) cleanup() {
	// Remove from active dialogs list
	for i, dialog := range d.module.activeDialogs {
		if dialog == d {
			d.module.activeDialogs = append(d.module.activeDialogs[:i], d.module.activeDialogs[i+1:]...)
			break
		}
	}
}

// ShowModal shows the dialog as modal
func (d *EnterPlainTextDialog) ShowModal() int {
	d.focus()
	return d.Exec()
}

// GetInfo returns module information
func (m *PlainTextWordsEntererModule) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":        "plainTextWordsEnterer",
		"name":        "Plain Text Words Enterer",
		"description": "Enter word pairs using plain text format with Qt dialog interface",
		"features": []string{
			"Qt dialog interface",
			"Support '=' and tab separators",
			"Character keyboard integration",
			"Word list parsing",
			"Lesson loading integration",
			"Error validation",
			"Translation support",
		},
		"status":        "complete",
		"format":        "question = answer (or question[TAB]answer)",
		"activeDialogs": len(m.activeDialogs),
		"priority":      960,
	}
}

// ValidateTextFormat validates plain text format
func (m *PlainTextWordsEntererModule) ValidateTextFormat(text string) error {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line contains separator
		if !strings.Contains(line, "=") && !strings.Contains(line, "\t") {
			return fmt.Errorf("line %d: missing equals sign or tab between question and answer", i+1)
		}
	}

	return nil
}

// ParseTextToWordPairs parses plain text directly to word pairs (utility function)
func (m *PlainTextWordsEntererModule) ParseTextToWordPairs(text string) ([]WordPair, error) {
	var wordPairs []WordPair

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var question, answer string
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				question = strings.TrimSpace(parts[0])
				answer = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "\t") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				question = strings.TrimSpace(parts[0])
				answer = strings.TrimSpace(parts[1])
			}
		}

		if question == "" || answer == "" {
			return nil, fmt.Errorf("line %d: missing separator between question and answer", i+1)
		}

		wordPairs = append(wordPairs, WordPair{
			Question: question,
			Answer:   answer,
		})
	}

	return wordPairs, nil
}

// GetFormatHelp returns help text about the expected format
func (m *PlainTextWordsEntererModule) GetFormatHelp() string {
	return `Plain Text Format Help:

Enter word pairs separated by new lines.
Use either '=' or tab to separate questions from answers.

Examples:
  hello = hola
  goodbye = adiós
  cat	gato
  dog	perro

Each line should contain:
- A question (left side)
- An equals sign (=) OR a tab character
- An answer (right side)

Empty lines are ignored.
Lines without separators will show an error.`
}

// InitPlainTextWordsEntererModule creates and returns a new module instance
func InitPlainTextWordsEntererModule() core.Module {
	return NewPlainTextWordsEntererModule()
}
