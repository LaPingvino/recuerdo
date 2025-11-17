package words

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/mappu/miqt/qt"
)

// CharacterSet represents a named group of characters
type CharacterSet struct {
	Name       string
	Characters []rune
	Category   string
}

// UnicodeBlock represents a Unicode block with its range and name
type UnicodeBlock struct {
	Name  string
	Start int
	End   int
}

// IntegratedUnicodePicker is the main Unicode character picker widget
type IntegratedUnicodePicker struct {
	*qt.QWidget

	logger     *logging.Logger
	configPath string

	// Character sets loaded from configuration
	characterSets []CharacterSet
	unicodeBlocks []UnicodeBlock

	// UI components
	stackedWidget *qt.QStackedWidget
	modeButtons   *qt.QButtonGroup

	// Quick access mode widgets
	quickButton      *qt.QPushButton
	quickCharsPage   *qt.QWidget
	quickAccessReady bool

	// Browse Unicode mode widgets
	browseButton      *qt.QPushButton
	browseUnicPage    *qt.QWidget
	browseScrollArea  *qt.QScrollArea
	browseCharsWidget *qt.QWidget
	browseCharsLayout *qt.QGridLayout
	blockComboBox     *qt.QComboBox
	browseReady       bool

	// Signal for character selection
	characterSelected func(string)
}

// NewIntegratedUnicodePicker creates a new integrated Unicode character picker
func NewIntegratedUnicodePicker(configPath string, parent *qt.QWidget) *IntegratedUnicodePicker {
	picker := &IntegratedUnicodePicker{
		QWidget:    qt.NewQWidget(parent),
		logger:     logging.NewLogger("IntegratedUnicodePicker"),
		configPath: configPath,
	}

	// Only load config and setup basic UI - defer heavy initialization
	picker.loadConfig()
	picker.setupBasicUI()

	return picker
}

// setupBasicUI creates the basic UI structure with lazy loading
func (up *IntegratedUnicodePicker) setupBasicUI() {
	up.logger.Debug("Setting up basic UI for Unicode picker")

	// Set a reasonable minimum size and style
	up.SetMinimumSize2(400, 300)
	up.SetStyleSheet(`
		QWidget {
			font-family: 'Segoe UI', Arial, sans-serif;
			font-size: 11px;
		}
		QPushButton {
			padding: 6px 12px;
			border: 1px solid #ccc;
			border-radius: 4px;
			background-color: #f8f9fa;
			min-height: 28px;
		}
		QPushButton:hover {
			background-color: #e6f3ff;
			border-color: #0078d4;
		}
		QPushButton:pressed {
			background-color: #cce4f7;
		}
		QPushButton:checked {
			background-color: #0078d4;
			color: white;
			border-color: #005a9e;
		}
		QLineEdit, QComboBox {
			padding: 4px;
			border: 1px solid #ccc;
			border-radius: 3px;
			font-size: 12px;
		}
		QScrollArea {
			border: 1px solid #ccc;
			background-color: white;
		}
	`)

	// Main layout
	mainLayout := qt.NewQVBoxLayout(up.QWidget)

	// Mode selection buttons
	modeButtonsLayout := qt.NewQHBoxLayout2()
	up.modeButtons = qt.NewQButtonGroup()

	up.quickButton = qt.NewQPushButton(nil)
	up.quickButton.SetText("Quick Access")
	up.quickButton.SetCheckable(true)
	up.quickButton.SetChecked(true)
	up.modeButtons.AddButton2(up.quickButton.QAbstractButton, 0)
	modeButtonsLayout.AddWidget(up.quickButton.QWidget)

	up.browseButton = qt.NewQPushButton(nil)
	up.browseButton.SetText("Browse Unicode")
	up.browseButton.SetCheckable(true)
	up.modeButtons.AddButton2(up.browseButton.QAbstractButton, 1)
	modeButtonsLayout.AddWidget(up.browseButton.QWidget)

	// Add stretch to center buttons
	modeButtonsLayout.AddStretch()
	mainLayout.AddLayout2(modeButtonsLayout.QLayout, 0)

	// Create stacked widget for different modes
	up.stackedWidget = qt.NewQStackedWidget(up.QWidget)
	mainLayout.AddWidget(up.stackedWidget.QWidget)

	// Connect mode buttons with lazy loading
	up.modeButtons.OnIdClicked(func(id int) {
		up.stackedWidget.SetCurrentIndex(id)
		if id == 0 {
			up.logger.Debug("Switched to Quick Access mode")
			up.ensureQuickAccessPageLoaded()
		} else if id == 1 {
			up.logger.Debug("Switched to Browse Unicode mode")
			up.ensureBrowseUnicodePageLoaded()
		}
	})

	// Initialize with empty pages - content will be loaded on demand
	up.quickCharsPage = qt.NewQWidget2()
	up.browseUnicPage = qt.NewQWidget2()

	up.stackedWidget.AddWidget(up.quickCharsPage)
	up.stackedWidget.AddWidget(up.browseUnicPage)

	// Load Quick Access by default (since it's selected)
	up.ensureQuickAccessPageLoaded()

	up.logger.Debug("Basic UI setup complete")
}

// ensureQuickAccessPageLoaded loads the quick access page if not already loaded
func (up *IntegratedUnicodePicker) ensureQuickAccessPageLoaded() {
	if up.quickAccessReady {
		return
	}

	up.logger.Debug("Loading Quick Access page...")

	layout := qt.NewQVBoxLayout(up.quickCharsPage)

	scrollArea := qt.NewQScrollArea(up.quickCharsPage)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAsNeeded)
	scrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAsNeeded)

	scrollWidget := qt.NewQWidget2()
	scrollLayout := qt.NewQVBoxLayout(scrollWidget)

	// Add character sets
	for _, charSet := range up.characterSets {
		up.createCharacterSetWidget(&charSet, scrollLayout)
	}

	scrollArea.SetWidget(scrollWidget)
	layout.AddWidget(scrollArea.QWidget)

	up.quickAccessReady = true
	up.logger.Debug("Quick Access page loaded successfully")
}

// ensureBrowseUnicodePageLoaded loads the browse unicode page if not already loaded
func (up *IntegratedUnicodePicker) ensureBrowseUnicodePageLoaded() {
	if up.browseReady {
		return
	}

	up.logger.Debug("Loading Browse Unicode page...")

	layout := qt.NewQVBoxLayout(up.browseUnicPage)

	// Controls at the top
	controlsLayout := qt.NewQVBoxLayout2()

	// Direct codepoint input
	codepointLayout := qt.NewQHBoxLayout2()
	codepointLabel := qt.NewQLabel(up.browseUnicPage)
	codepointLabel.SetText("Enter Unicode codepoint (U+xxxx):")
	codepointLayout.AddWidget(codepointLabel.QWidget)

	codepointInput := qt.NewQLineEdit(up.browseUnicPage)
	codepointInput.SetPlaceholderText("U+1F600")
	codepointLayout.AddWidget(codepointInput.QWidget)

	addCodepointBtn := qt.NewQPushButton(nil)
	addCodepointBtn.SetText("Add Character")
	addCodepointBtn.OnClicked(func() {
		text := codepointInput.Text()
		if strings.HasPrefix(strings.ToUpper(text), "U+") {
			hexStr := strings.TrimPrefix(strings.ToUpper(text), "U+")
			if codepoint, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
				char := string(rune(codepoint))
				if up.characterSelected != nil {
					up.characterSelected(char)
				}
				codepointInput.Clear()
				up.logger.Debug("Added character from codepoint: %s -> %s", text, char)
			} else {
				up.logger.Warning("Invalid codepoint format: %s", text)
			}
		}
	})
	codepointLayout.AddWidget(addCodepointBtn.QWidget)
	controlsLayout.AddLayout2(codepointLayout.QLayout, 0)

	// Unicode block selection
	blockLayout := qt.NewQHBoxLayout2()
	blockLabel := qt.NewQLabel(up.browseUnicPage)
	blockLabel.SetText("Unicode Block:")
	blockLayout.AddWidget(blockLabel.QWidget)
	up.blockComboBox = qt.NewQComboBox(up.browseUnicPage)
	up.blockComboBox.AddItem("Select a block...")

	// Add common blocks first
	commonBlocks := []string{
		"Basic Latin", "Latin-1 Supplement", "Latin Extended-A", "Latin Extended-B",
		"Greek and Coptic", "Cyrillic", "Arabic", "Hebrew",
		"General Punctuation", "Currency Symbols", "Mathematical Operators",
		"Arrows", "Box Drawing", "Block Elements", "Geometric Shapes",
	}

	for _, blockName := range commonBlocks {
		for _, block := range up.unicodeBlocks {
			if block.Name == blockName {
				up.blockComboBox.AddItem(block.Name)
				break
			}
		}
	}

	// Add separator and remaining blocks
	up.blockComboBox.InsertSeparator(up.blockComboBox.Count())
	up.blockComboBox.AddItem("--- All Blocks ---")

	for _, block := range up.unicodeBlocks {
		// Skip if already added to common blocks
		isCommon := false
		for _, commonName := range commonBlocks {
			if block.Name == commonName {
				isCommon = true
				break
			}
		}
		if !isCommon {
			up.blockComboBox.AddItem(block.Name)
		}
	}

	blockLayout.AddWidget(up.blockComboBox.QWidget)
	controlsLayout.AddLayout2(blockLayout.QLayout, 0)
	layout.AddLayout2(controlsLayout.QLayout, 0)

	// Scrollable area for Unicode characters
	up.browseScrollArea = qt.NewQScrollArea(up.browseUnicPage)
	up.browseScrollArea.SetWidgetResizable(true)
	up.browseScrollArea.SetVerticalScrollBarPolicy(qt.ScrollBarAsNeeded)
	up.browseScrollArea.SetHorizontalScrollBarPolicy(qt.ScrollBarAsNeeded)

	up.browseCharsWidget = qt.NewQWidget2()
	up.browseCharsLayout = qt.NewQGridLayout(up.browseCharsWidget)
	up.browseCharsLayout.SetSpacing(2)
	up.browseScrollArea.SetWidget(up.browseCharsWidget)

	// Add placeholder text
	placeholderLabel := qt.NewQLabel(up.browseCharsWidget)
	placeholderLabel.SetText("Select a Unicode block above to view characters, or enter a codepoint directly.")
	placeholderLabel.SetStyleSheet("color: #666; font-style: italic; padding: 20px; text-align: center;")
	placeholderLabel.SetAlignment(qt.AlignCenter)
	up.browseCharsLayout.AddWidget(placeholderLabel.QWidget)

	layout.AddWidget(up.browseScrollArea.QWidget)

	// Connect block selection
	up.blockComboBox.OnCurrentTextChanged(func(text string) {
		if text != "Select a block..." && text != "--- All Blocks ---" && text != "" {
			up.browseUnicodeBlockPaginated(text, 0)
		}
	})

	up.browseReady = true
	up.logger.Debug("Browse Unicode page loaded successfully")
}

// createCharacterSetWidget creates a widget for a character set
func (up *IntegratedUnicodePicker) createCharacterSetWidget(charSet *CharacterSet, parentLayout *qt.QVBoxLayout) {
	if len(charSet.Characters) == 0 {
		return
	}

	// Group box for the character set
	groupBox := qt.NewQGroupBox(nil)
	groupBox.SetTitle(charSet.Name)
	groupBox.SetStyleSheet(`
		QGroupBox {
			font-weight: bold;
			border: 2px solid #ccc;
			border-radius: 8px;
			margin-top: 1ex;
		}
		QGroupBox::title {
			subcontrol-origin: margin;
			left: 10px;
			padding: 0 5px 0 5px;
		}
	`)

	// Grid layout for characters
	gridLayout := qt.NewQGridLayout(groupBox.QWidget)
	gridLayout.SetSpacing(2)

	const maxCols = 12
	row, col := 0, 0

	// Create buttons for each character
	for _, char := range charSet.Characters {
		if !unicode.IsPrint(char) && !unicode.IsSpace(char) {
			continue // Skip non-printable characters
		}

		charStr := string(char)
		btn := qt.NewQPushButton(nil)
		btn.SetText(charStr)
		btn.SetMinimumSize2(32, 32)
		btn.SetMaximumSize2(32, 32)
		btn.SetToolTip(fmt.Sprintf("U+%04X: %s", int(char), charStr))
		btn.SetStyleSheet(`
			QPushButton {
				font-size: 14px;
				font-weight: bold;
			}
		`)

		// Connect button click to emit character
		btn.OnClicked(func() {
			if up.characterSelected != nil {
				up.characterSelected(charStr)
			}
			up.logger.Debug("Character selected: %s (U+%04X)", charStr, int(char))
		})

		gridLayout.AddWidget2(btn.QWidget, row, col)

		col++
		if col >= maxCols {
			col = 0
			row++
		}
	}

	parentLayout.AddWidget(groupBox.QWidget)
}

// browseUnicodeBlockPaginated displays characters from a Unicode block with pagination
func (up *IntegratedUnicodePicker) browseUnicodeBlockPaginated(blockName string, page int) {
	up.logger.Debug("Browsing Unicode block: %s, page: %d", blockName, page)

	// Find the block
	var block *UnicodeBlock
	for i := range up.unicodeBlocks {
		if up.unicodeBlocks[i].Name == blockName {
			block = &up.unicodeBlocks[i]
			break
		}
	}

	if block == nil {
		up.logger.Warning("Unicode block not found: %s", blockName)
		return
	}

	// Clear existing content
	up.clearLayout(up.browseCharsLayout)

	const charsPerPage = 120
	const maxCols = 12

	start := block.Start + (page * charsPerPage)
	end := start + charsPerPage
	if end > block.End {
		end = block.End
	}

	row, col := 0, 0
	charCount := 0

	// Calculate total pages for pagination
	totalChars := block.End - block.Start + 1
	totalPages := (totalChars + charsPerPage - 1) / charsPerPage

	// Create grid layout for characters
	gridLayout := up.browseCharsLayout

	// Create character buttons
	for codepoint := start; codepoint <= end && codepoint <= block.End; codepoint++ {
		char := rune(codepoint)
		if !unicode.IsPrint(char) {
			continue
		}

		charStr := string(char)
		charBtn := qt.NewQPushButton(nil)
		charBtn.SetText(charStr)
		charBtn.SetMinimumSize2(32, 32)
		charBtn.SetMaximumSize2(32, 32)
		charBtn.SetToolTip(fmt.Sprintf("U+%04X: %s", codepoint, charStr))
		charBtn.SetStyleSheet(`
			QPushButton {
				font-size: 14px;
				font-weight: bold;
			}
		`)

		// Connect button click
		charBtn.OnClicked(func() {
			if up.characterSelected != nil {
				up.characterSelected(charStr)
			}
			up.logger.Debug("Character selected: %s (U+%04X)", charStr, codepoint)
		})

		gridLayout.AddWidget2(charBtn.QWidget, row, col)
		col++
		if col >= maxCols {
			col = 0
			row++
		}
		charCount++
	}

	// Add pagination controls if needed
	if totalPages > 1 {
		row++
		paginationLayout := qt.NewQHBoxLayout2()

		prevBtn := qt.NewQPushButton(nil)
		prevBtn.SetText("Previous")
		prevBtn.SetEnabled(page > 0)
		if page > 0 {
			prevBtn.OnClicked(func() {
				up.browseUnicodeBlockPaginated(blockName, page-1)
			})
		}
		paginationLayout.AddWidget(prevBtn.QWidget)

		// Page indicator
		pageLabel := qt.NewQLabel(up.browseCharsWidget)
		pageLabel.SetText(fmt.Sprintf("Page %d of %d", page+1, totalPages))
		pageLabel.SetAlignment(qt.AlignCenter)
		paginationLayout.AddWidget(pageLabel.QWidget)

		nextBtn := qt.NewQPushButton(nil)
		nextBtn.SetText("Next")
		nextBtn.SetEnabled(page < totalPages-1)
		if page < totalPages-1 {
			nextBtn.OnClicked(func() {
				up.browseUnicodeBlockPaginated(blockName, page+1)
			})
		}
		paginationLayout.AddWidget(nextBtn.QWidget)

		paginationWidget := qt.NewQWidget2()
		paginationWidget.SetLayout(paginationLayout.QLayout)
		gridLayout.AddWidget2(paginationWidget, row, 0)
	}

	up.logger.Debug("Displayed %d characters from block %s", charCount, blockName)
}

// clearLayout removes all widgets from a layout
func (up *IntegratedUnicodePicker) clearLayout(layout *qt.QGridLayout) {
	// Remove all items from the layout
	for {
		item := layout.TakeAt(0)
		if item == nil {
			break
		}
		widget := item.Widget()
		if widget != nil {
			widget.SetParent(nil)
		}
	}
}

// loadConfig loads character sets and Unicode blocks from configuration
func (up *IntegratedUnicodePicker) loadConfig() {
	up.logger.Debug("Loading Unicode picker configuration")

	// Load character sets
	up.loadCharacterSets()

	// Load Unicode blocks
	up.loadUnicodeBlocks()

	up.logger.Debug("Configuration loaded: %d character sets, %d Unicode blocks",
		len(up.characterSets), len(up.unicodeBlocks))
}

// loadCharacterSets loads predefined character sets
func (up *IntegratedUnicodePicker) loadCharacterSets() {
	// Define common character sets
	up.characterSets = []CharacterSet{
		{
			Name:       "Latin Accents",
			Category:   "Basic",
			Characters: []rune{'à', 'á', 'â', 'ã', 'ä', 'å', 'æ', 'ç', 'è', 'é', 'ê', 'ë', 'ì', 'í', 'î', 'ï', 'ñ', 'ò', 'ó', 'ô', 'õ', 'ö', 'ø', 'ù', 'ú', 'û', 'ü', 'ý', 'ÿ'},
		},
		{
			Name:       "Currency",
			Category:   "Symbols",
			Characters: []rune{'€', '$', '£', '¥', '¢', '₹', '₽', '₩', '₦', '₴', '₸', '₼'},
		},
		{
			Name:       "Math Symbols",
			Category:   "Math",
			Characters: []rune{'±', '×', '÷', '√', '∞', '≈', '≠', '≤', '≥', '∫', '∑', '∏', '∂', '∆', 'π', 'α', 'β', 'γ', 'δ', 'θ', 'λ', 'μ', 'σ', 'φ', 'ω'},
		},
		{
			Name:       "Arrows",
			Category:   "Symbols",
			Characters: []rune{'←', '↑', '→', '↓', '↔', '↕', '↖', '↗', '↘', '↙', '⇐', '⇑', '⇒', '⇓', '⇔'},
		},
		{
			Name:       "Punctuation",
			Category:   "Basic",
			Characters: []rune{0x201C, 0x201D, 0x2018, 0x2019, '«', '»', '‹', '›', '–', '—', '…', '•', '‚', '„', '‰', '′', '″'},
		},
	}

	// Load from file if it exists
	if up.configPath != "" && fileExists(up.configPath) {
		up.loadCharacterSetsFromFile()
	}
}

// loadCharacterSetsFromFile loads character sets from a JSON file
func (up *IntegratedUnicodePicker) loadCharacterSetsFromFile() {
	data, err := ioutil.ReadFile(up.configPath)
	if err != nil {
		up.logger.Warning("Failed to read character sets file: %v", err)
		return
	}

	var fileSets []CharacterSet
	if err := json.Unmarshal(data, &fileSets); err != nil {
		up.logger.Warning("Failed to parse character sets file: %v", err)
		return
	}

	// Append file sets to default sets
	up.characterSets = append(up.characterSets, fileSets...)
	up.logger.Debug("Loaded %d character sets from file", len(fileSets))
}

// loadUnicodeBlocks loads Unicode block definitions
func (up *IntegratedUnicodePicker) loadUnicodeBlocks() {
	// Define common Unicode blocks (simplified list)
	blocks := []UnicodeBlock{
		{"Basic Latin", 0x0000, 0x007F},
		{"Latin-1 Supplement", 0x0080, 0x00FF},
		{"Latin Extended-A", 0x0100, 0x017F},
		{"Latin Extended-B", 0x0180, 0x024F},
		{"Greek and Coptic", 0x0370, 0x03FF},
		{"Cyrillic", 0x0400, 0x04FF},
		{"Arabic", 0x0600, 0x06FF},
		{"Hebrew", 0x0590, 0x05FF},
		{"General Punctuation", 0x2000, 0x206F},
		{"Currency Symbols", 0x20A0, 0x20CF},
		{"Mathematical Operators", 0x2200, 0x22FF},
		{"Arrows", 0x2190, 0x21FF},
		{"Box Drawing", 0x2500, 0x257F},
		{"Block Elements", 0x2580, 0x259F},
		{"Geometric Shapes", 0x25A0, 0x25FF},
		{"Miscellaneous Symbols", 0x2600, 0x26FF},
		{"Emoticons", 0x1F600, 0x1F64F},
		{"Miscellaneous Symbols and Pictographs", 0x1F300, 0x1F5FF},
		{"Transport and Map Symbols", 0x1F680, 0x1F6FF},
	}

	up.unicodeBlocks = blocks

	// Sort by name for better UX
	sort.Slice(up.unicodeBlocks, func(i, j int) bool {
		return up.unicodeBlocks[i].Name < up.unicodeBlocks[j].Name
	})
}

// OnCharacterSelected sets the callback for when a character is selected
func (up *IntegratedUnicodePicker) OnCharacterSelected(callback func(string)) {
	up.characterSelected = callback
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// SaveCharacterSets saves the current character sets to a file
func (up *IntegratedUnicodePicker) SaveCharacterSets(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	data, err := json.MarshalIndent(up.characterSets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal character sets: %v", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	up.logger.Debug("Saved %d character sets to %s", len(up.characterSets), filename)
	return nil
}

// AddCustomCharacterSet adds a custom character set
func (up *IntegratedUnicodePicker) AddCustomCharacterSet(name, category string, characters []rune) {
	charSet := CharacterSet{
		Name:       name,
		Category:   category,
		Characters: characters,
	}

	up.characterSets = append(up.characterSets, charSet)

	// Refresh UI if already loaded
	if up.quickAccessReady {
		up.quickAccessReady = false
		up.ensureQuickAccessPageLoaded()
	}

	up.logger.Debug("Added custom character set: %s with %d characters", name, len(characters))
}
