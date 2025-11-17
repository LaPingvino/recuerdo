package words

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/mappu/miqt/qt"
)

// EnhancedLineEdit is a QLineEdit that handles keyboard input properly
// by intercepting key events and processing dead keys manually
type EnhancedLineEdit struct {
	*qt.QLineEdit
	logger          *logging.Logger
	deadKeyState    rune
	keyboardLayout  string
	deadKeyMappings map[rune]map[rune]rune
	altgrMappings   map[int]rune
	debugMode       bool
}

// NewEnhancedLineEdit creates a new enhanced line edit with proper keyboard handling
func NewEnhancedLineEdit(parent *qt.QWidget) *EnhancedLineEdit {
	edit := &EnhancedLineEdit{
		QLineEdit:       qt.NewQLineEdit(parent),
		logger:          logging.NewLogger("EnhancedLineEdit"),
		deadKeyMappings: make(map[rune]map[rune]rune),
		altgrMappings:   make(map[int]rune),
		debugMode:       true,
	}

	edit.initializeKeyboardMappings()
	edit.connectKeyEventHandler()

	return edit
}

// initializeKeyboardMappings sets up dead key and AltGr mappings
func (e *EnhancedLineEdit) initializeKeyboardMappings() {
	// Detect current keyboard layout
	e.detectKeyboardLayout()

	// Initialize dead key mappings
	e.deadKeyMappings = map[rune]map[rune]rune{
		// Acute accent (') - most common for Portuguese
		'\'': {
			'a': 'á', 'e': 'é', 'i': 'í', 'o': 'ó', 'u': 'ú', 'y': 'ý',
			'A': 'Á', 'E': 'É', 'I': 'Í', 'O': 'Ó', 'U': 'Ú', 'Y': 'Ý',
		},
		// Acute accent (´) - alternative form
		'´': {
			'a': 'á', 'e': 'é', 'i': 'í', 'o': 'ó', 'u': 'ú', 'y': 'ý',
			'A': 'Á', 'E': 'É', 'I': 'Í', 'O': 'Ó', 'U': 'Ú', 'Y': 'Ý',
		},
		// Grave accent (`)
		'`': {
			'a': 'à', 'e': 'è', 'i': 'ì', 'o': 'ò', 'u': 'ù',
			'A': 'À', 'E': 'È', 'I': 'Ì', 'O': 'Ò', 'U': 'Ù',
		},
		// Circumflex (^)
		'^': {
			'a': 'â', 'e': 'ê', 'i': 'î', 'o': 'ô', 'u': 'û',
			'A': 'Â', 'E': 'Ê', 'I': 'Î', 'O': 'Ô', 'U': 'Û',
		},
		// Tilde (~) - important for Portuguese
		'~': {
			'a': 'ã', 'n': 'ñ', 'o': 'õ',
			'A': 'Ã', 'N': 'Ñ', 'O': 'Õ',
		},
		// Diaeresis (¨)
		'¨': {
			'a': 'ä', 'e': 'ë', 'i': 'ï', 'o': 'ö', 'u': 'ü', 'y': 'ÿ',
			'A': 'Ä', 'E': 'Ë', 'I': 'Ï', 'O': 'Ö', 'U': 'Ü', 'Y': 'Ÿ',
		},
		// Cedilla - for ç
		',': {
			'c': 'ç', 'C': 'Ç',
		},
	}

	// Initialize AltGr mappings (common ones)
	e.altgrMappings = map[int]rune{
		int(qt.Key_2):     '@',  // AltGr + 2 = @
		int(qt.Key_3):     '#',  // AltGr + 3 = #
		int(qt.Key_4):     '€',  // AltGr + 4 = € (Euro)
		int(qt.Key_5):     '£',  // AltGr + 5 = £ (Pound)
		int(qt.Key_7):     '{',  // AltGr + 7 = {
		int(qt.Key_8):     '[',  // AltGr + 8 = [
		int(qt.Key_9):     ']',  // AltGr + 9 = ]
		int(qt.Key_0):     '}',  // AltGr + 0 = }
		int(qt.Key_Minus): '\\', // AltGr + - = \
		int(qt.Key_Equal): '|',  // AltGr + = = |
	}

	e.logger.Debug("Initialized keyboard mappings for layout: %s", e.keyboardLayout)
}

// detectKeyboardLayout queries the system for current keyboard layout
func (e *EnhancedLineEdit) detectKeyboardLayout() {
	cmd := exec.Command("setxkbmap", "-query")
	output, err := cmd.Output()
	if err != nil {
		e.logger.Warning("Failed to detect keyboard layout: %v", err)
		e.keyboardLayout = "us"
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "layout:") {
			e.keyboardLayout = strings.TrimSpace(strings.Split(line, ":")[1])
			break
		}
	}

	if e.keyboardLayout == "" {
		e.keyboardLayout = "us"
	}

	e.logger.Info("Detected keyboard layout: %s", e.keyboardLayout)
}

// connectKeyEventHandler sets up the key event interception
func (e *EnhancedLineEdit) connectKeyEventHandler() {
	// Override keyPressEvent by connecting to key press
	e.OnKeyPressEvent(func(super func(param1 *qt.QKeyEvent), event *qt.QKeyEvent) {
		e.handleKeyPress(event)
	})
}

// handleKeyPress processes key press events with dead key support
func (e *EnhancedLineEdit) handleKeyPress(event *qt.QKeyEvent) {
	key := event.Key()
	modifiers := event.Modifiers()
	text := event.Text()

	if e.debugMode {
		e.logger.Debug("Key event: key=%d, mods=%d, text='%s'", key, int(modifiers), text)
	}

	// Check for AltGr combinations (Ctrl+Alt in Qt)
	isAltGr := (modifiers&qt.AltModifier != 0) && (modifiers&qt.ControlModifier != 0)
	if isAltGr {
		if char, exists := e.altgrMappings[key]; exists {
			e.insertCharacterAtCursor(string(char))
			if e.debugMode {
				e.logger.Debug("AltGr combination: %c", char)
			}
			return
		}
	}

	// Handle regular text input
	if len(text) == 1 {
		char := rune(text[0])

		// Check if this is a potential dead key
		if _, isDeadKey := e.deadKeyMappings[char]; isDeadKey {
			// This might be a dead key - store it and wait for next character
			e.deadKeyState = char
			if e.debugMode {
				e.logger.Debug("Potential dead key: %c", char)
			}
			return // Don't insert the dead key itself
		}

		// Check if we have a pending dead key
		if e.deadKeyState != 0 {
			if deadMap, exists := e.deadKeyMappings[e.deadKeyState]; exists {
				if combinedChar, exists := deadMap[char]; exists {
					// Valid dead key combination
					e.insertCharacterAtCursor(string(combinedChar))
					if e.debugMode {
						e.logger.Debug("Dead key combination: %c + %c = %c", e.deadKeyState, char, combinedChar)
					}
					e.deadKeyState = 0 // Reset dead key state
					return
				}
			}

			// Invalid dead key combination - insert both characters
			e.insertCharacterAtCursor(string(e.deadKeyState))
			e.insertCharacterAtCursor(string(char))
			e.deadKeyState = 0
			if e.debugMode {
				e.logger.Debug("Invalid dead key combo - inserted both: %c %c", e.deadKeyState, char)
			}
			return
		}
	}

	// Handle special keys
	switch key {
	case int(qt.Key_Backspace):
		e.deadKeyState = 0 // Reset dead key on backspace
		// Let Qt handle normally
	case int(qt.Key_Escape), int(qt.Key_Tab), int(qt.Key_Return), int(qt.Key_Enter):
		e.deadKeyState = 0 // Reset dead key on navigation
		// Let Qt handle normally
	default:
		// Let Qt handle normally for other cases
	}
}

// insertCharacterAtCursor inserts text at the current cursor position
func (e *EnhancedLineEdit) insertCharacterAtCursor(text string) {
	cursorPos := e.CursorPosition()
	currentText := e.Text()

	// Split at cursor position
	before := currentText[:cursorPos]
	after := currentText[cursorPos:]

	// Insert new text
	newText := before + text + after
	e.SetText(newText)

	// Move cursor to after inserted text
	newCursorPos := cursorPos + len([]rune(text))
	e.SetCursorPosition(newCursorPos)

	if e.debugMode {
		e.logger.Debug("Inserted '%s' at position %d", text, cursorPos)
	}
}

// SetDebugMode enables or disables debug logging
func (e *EnhancedLineEdit) SetDebugMode(enabled bool) {
	e.debugMode = enabled
}

// ResetDeadKeyState clears any pending dead key state
func (e *EnhancedLineEdit) ResetDeadKeyState() {
	e.deadKeyState = 0
}

// GetKeyboardLayout returns the detected keyboard layout
func (e *EnhancedLineEdit) GetKeyboardLayout() string {
	return e.keyboardLayout
}

// TestDeadKeys tests the dead key functionality
func (e *EnhancedLineEdit) TestDeadKeys() {
	fmt.Println("Testing Enhanced Input Dead Keys:")

	testCases := []struct {
		deadKey  rune
		base     rune
		expected rune
		desc     string
	}{
		{'\'', 'a', 'á', "apostrophe + a"},
		{'\'', 'e', 'é', "apostrophe + e"},
		{'`', 'a', 'à', "grave + a"},
		{'^', 'e', 'ê', "circumflex + e"},
		{'~', 'a', 'ã', "tilde + a"},
		{'~', 'o', 'õ', "tilde + o"},
		{'~', 'n', 'ñ', "tilde + n"},
		{',', 'c', 'ç', "comma + c"},
	}

	for _, tc := range testCases {
		if deadMap, exists := e.deadKeyMappings[tc.deadKey]; exists {
			if result, exists := deadMap[tc.base]; exists {
				status := "✓"
				if result != tc.expected {
					status = "✗"
				}
				fmt.Printf("  %s %s = %c\n", status, tc.desc, result)
			} else {
				fmt.Printf("  - %s = (not mapped)\n", tc.desc)
			}
		}
	}
}

// AddCustomDeadKey adds a custom dead key mapping
func (e *EnhancedLineEdit) AddCustomDeadKey(deadKey rune, base rune, result rune) {
	if _, exists := e.deadKeyMappings[deadKey]; !exists {
		e.deadKeyMappings[deadKey] = make(map[rune]rune)
	}
	e.deadKeyMappings[deadKey][base] = result
	e.logger.Debug("Added custom dead key: %c + %c = %c", deadKey, base, result)
}

// AddCustomAltGr adds a custom AltGr mapping
func (e *EnhancedLineEdit) AddCustomAltGr(keyCode int, result rune) {
	e.altgrMappings[keyCode] = result
	e.logger.Debug("Added custom AltGr: key %d = %c", keyCode, result)
}
