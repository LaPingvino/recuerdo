// Package charskeyboard provides functionality ported from Python module
//
// Provides a virtual character keyboard for entering special characters
// and diacritics that may not be easily accessible on a standard keyboard.
//
// This is an automated port - implementation may be incomplete.
package charskeyboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// CharsKeyboardModule is a Go port of the Python CharsKeyboardModule class
type CharsKeyboardModule struct {
	*core.BaseModule
	manager        *core.Manager
	widget         *qt.QWidget
	charButtons    []*qt.QPushButton
	targetLineEdit *qt.QLineEdit
	currentChars   []string
}

// NewCharsKeyboardModule creates a new CharsKeyboardModule instance
func NewCharsKeyboardModule() *CharsKeyboardModule {
	base := core.NewBaseModule("ui", "chars-keyboard-module")
	base.SetRequires("qtApp")

	return &CharsKeyboardModule{
		BaseModule:   base,
		charButtons:  make([]*qt.QPushButton, 0),
		currentChars: make([]string, 0),
	}
}

// GetWidget returns the main widget for this keyboard
func (mod *CharsKeyboardModule) GetWidget() *qt.QWidget {
	if mod.widget == nil {
		mod.createWidget()
	}
	return mod.widget
}

// createWidget creates the character keyboard widget
func (mod *CharsKeyboardModule) createWidget() {
	mod.widget = qt.NewQWidget(nil)
	layout := qt.NewQVBoxLayout(mod.widget)

	// Header
	headerLabel := qt.NewQLabel(mod.widget)
	headerLabel.SetText("Special Characters")
	headerFont := headerLabel.Font()
	headerFont.SetBold(true)
	headerLabel.SetFont(headerFont)
	layout.AddWidget(headerLabel.QWidget)

	// Character sets
	mod.createCharacterSet("Common", []string{"á", "é", "í", "ó", "ú", "ñ", "ü", "ç"}, layout)
	mod.createCharacterSet("German", []string{"ä", "ö", "ü", "ß", "Ä", "Ö", "Ü"}, layout)
	mod.createCharacterSet("French", []string{"à", "è", "ù", "â", "ê", "î", "ô", "û", "ë", "ï", "ÿ", "ç"}, layout)
	mod.createCharacterSet("Spanish", []string{"á", "é", "í", "ó", "ú", "ñ", "¿", "¡"}, layout)

	// Clear button
	clearButton := qt.NewQPushButton(nil)
	clearButton.SetText("Clear")
	clearButton.OnClicked(func() {
		mod.clearTarget()
	})
	layout.AddWidget(clearButton.QWidget)
}

// createCharacterSet creates a set of character buttons
func (mod *CharsKeyboardModule) createCharacterSet(setName string, chars []string, parentLayout *qt.QVBoxLayout) {
	groupBox := qt.NewQGroupBox(nil)
	groupBox.SetTitle(setName)
	groupLayout := qt.NewQHBoxLayout(groupBox.QWidget)

	for _, char := range chars {
		button := qt.NewQPushButton(nil)
		button.SetText(char)
		button.SetMaximumSize2(40, 40)
		button.SetMinimumSize2(40, 40)

		// Capture the character in the closure
		character := char
		button.OnClicked(func() {
			mod.insertCharacter(character)
		})

		groupLayout.AddWidget(button.QWidget)
		mod.charButtons = append(mod.charButtons, button)
	}

	parentLayout.AddWidget(groupBox.QWidget)
}

// SetTargetLineEdit sets the line edit widget that will receive the characters
func (mod *CharsKeyboardModule) SetTargetLineEdit(lineEdit *qt.QLineEdit) {
	mod.targetLineEdit = lineEdit
}

// insertCharacter inserts a character into the target line edit
func (mod *CharsKeyboardModule) insertCharacter(char string) {
	if mod.targetLineEdit == nil {
		fmt.Printf("Character typed: %s (no target set)\n", char)
		return
	}

	// Get current text and cursor position
	currentText := mod.targetLineEdit.Text()
	cursorPos := mod.targetLineEdit.CursorPosition()

	// Insert character at cursor position
	newText := currentText[:cursorPos] + char + currentText[cursorPos:]
	mod.targetLineEdit.SetText(newText)

	// Move cursor after inserted character
	mod.targetLineEdit.SetCursorPosition(cursorPos + len(char))

	// Give focus back to the line edit
	mod.targetLineEdit.SetFocus()
}

// clearTarget clears the target line edit
func (mod *CharsKeyboardModule) clearTarget() {
	if mod.targetLineEdit != nil {
		mod.targetLineEdit.Clear()
		mod.targetLineEdit.SetFocus()
	}
}

// AddCustomCharacters adds custom characters to the keyboard
func (mod *CharsKeyboardModule) AddCustomCharacters(setName string, chars []string) {
	// Store characters for later use, avoid layout casting issues for now
	mod.currentChars = append(mod.currentChars, chars...)
	fmt.Printf("Added custom character set '%s': %v\n", setName, chars)
}

// GetAvailableCharacters returns all available characters
func (mod *CharsKeyboardModule) GetAvailableCharacters() []string {
	chars := []string{}
	for _, button := range mod.charButtons {
		chars = append(chars, button.Text())
	}
	return chars
}

// FilterCharacters shows only characters matching the filter
func (mod *CharsKeyboardModule) FilterCharacters(filter string) {
	filter = strings.ToLower(filter)

	for _, button := range mod.charButtons {
		char := strings.ToLower(button.Text())
		if filter == "" || strings.Contains(char, filter) {
			button.SetVisible(true)
		} else {
			button.SetVisible(false)
		}
	}
}

// ResetFilter shows all characters
func (mod *CharsKeyboardModule) ResetFilter() {
	for _, button := range mod.charButtons {
		button.SetVisible(true)
	}
}

// Enable activates the module
func (mod *CharsKeyboardModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("CharsKeyboardModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *CharsKeyboardModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up widget
	if mod.widget != nil {
		mod.widget.Close()
		mod.widget = nil
		mod.charButtons = make([]*qt.QPushButton, 0)
	}

	mod.targetLineEdit = nil

	fmt.Println("CharsKeyboardModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *CharsKeyboardModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitCharsKeyboardModule creates and returns a new CharsKeyboardModule instance
func InitCharsKeyboardModule() core.Module {
	return NewCharsKeyboardModule()
}
