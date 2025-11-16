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
	"unicode/utf8"

	"github.com/LaPingvino/recuerdo/internal/logging"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// CharacterSet represents a named group of characters
type CharacterSet struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Characters  []string `json:"characters"`
}

// UnicodePickerConfig holds the configuration for character sets
type UnicodePickerConfig struct {
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	CharacterSets []CharacterSet `json:"character_sets"`
}

// UnicodeBlockInfo represents a Unicode block with its range
type UnicodeBlockInfo struct {
	Name  string
	Start rune
	End   rune
}

// IntegratedUnicodePicker provides an integrated Unicode character picker
type IntegratedUnicodePicker struct {
	*widgets.QWidget

	logger     *logging.Logger
	config     *UnicodePickerConfig
	targetEdit *widgets.QLineEdit
	configPath string

	// UI components (lazy loaded)
	stackedWidget  *widgets.QStackedWidget
	quickCharsPage *widgets.QWidget
	browseUnicPage *widgets.QWidget

	// Browse Unicode components (lazy loaded)
	categoryComboBox  *widgets.QComboBox
	blockComboBox     *widgets.QComboBox
	rangeFromEdit     *widgets.QLineEdit
	rangeToEdit       *widgets.QLineEdit
	searchEdit        *widgets.QLineEdit
	browseCharsLayout *widgets.QGridLayout
	browseScrollArea  *widgets.QScrollArea
	browseWidget      *widgets.QWidget

	// Navigation
	modeButtons  *widgets.QButtonGroup
	quickButton  *widgets.QPushButton
	browseButton *widgets.QPushButton

	// Unicode data (lazy loaded)
	unicodeBlocks  []UnicodeBlockInfo
	categoryTables map[string]*unicode.RangeTable

	// Lazy loading flags
	quickPageInitialized   bool
	browsePageInitialized  bool
	unicodeDataInitialized bool
}

// NewIntegratedUnicodePicker creates a new integrated Unicode character picker
func NewIntegratedUnicodePicker(configPath string, parent widgets.QWidget_ITF) *IntegratedUnicodePicker {
	picker := &IntegratedUnicodePicker{
		QWidget:    widgets.NewQWidget(parent, 0),
		logger:     logging.NewLogger("IntegratedUnicodePicker"),
		configPath: configPath,
	}

	// Only load config and setup basic UI - defer heavy initialization
	picker.loadConfig()
	picker.setupBasicUI()

	return picker
}

// initializeUnicodeData initializes Unicode blocks and categories using Go's unicode package
func (up *IntegratedUnicodePicker) initializeUnicodeData() {
	// Initialize Unicode blocks with dynamic detection
	up.unicodeBlocks = []UnicodeBlockInfo{
		{"Basic Latin", 0x0000, 0x007F},
		{"Latin-1 Supplement", 0x0080, 0x00FF},
		{"Latin Extended-A", 0x0100, 0x017F},
		{"Latin Extended-B", 0x0180, 0x024F},
		{"IPA Extensions", 0x0250, 0x02AF},
		{"Spacing Modifier Letters", 0x02B0, 0x02FF},
		{"Combining Diacritical Marks", 0x0300, 0x036F},
		{"Greek and Coptic", 0x0370, 0x03FF},
		{"Cyrillic", 0x0400, 0x04FF},
		{"Cyrillic Supplement", 0x0500, 0x052F},
		{"Armenian", 0x0530, 0x058F},
		{"Hebrew", 0x0590, 0x05FF},
		{"Arabic", 0x0600, 0x06FF},
		{"Devanagari", 0x0900, 0x097F},
		{"Bengali", 0x0980, 0x09FF},
		{"Thai", 0x0E00, 0x0E7F},
		{"Tibetan", 0x0F00, 0x0FFF},
		{"Georgian", 0x10A0, 0x10FF},
		{"Hangul Jamo", 0x1100, 0x11FF},
		{"Latin Extended Additional", 0x1E00, 0x1EFF},
		{"Greek Extended", 0x1F00, 0x1FFF},
		{"General Punctuation", 0x2000, 0x206F},
		{"Superscripts and Subscripts", 0x2070, 0x209F},
		{"Currency Symbols", 0x20A0, 0x20CF},
		{"Letterlike Symbols", 0x2100, 0x214F},
		{"Number Forms", 0x2150, 0x218F},
		{"Arrows", 0x2190, 0x21FF},
		{"Mathematical Operators", 0x2200, 0x22FF},
		{"Miscellaneous Technical", 0x2300, 0x23FF},
		{"Control Pictures", 0x2400, 0x243F},
		{"Optical Character Recognition", 0x2440, 0x245F},
		{"Enclosed Alphanumerics", 0x2460, 0x24FF},
		{"Box Drawing", 0x2500, 0x257F},
		{"Block Elements", 0x2580, 0x259F},
		{"Geometric Shapes", 0x25A0, 0x25FF},
		{"Miscellaneous Symbols", 0x2600, 0x26FF},
		{"Dingbats", 0x2700, 0x27BF},
		{"Miscellaneous Mathematical Symbols-A", 0x27C0, 0x27EF},
		{"Supplemental Arrows-A", 0x27F0, 0x27FF},
		{"CJK Unified Ideographs", 0x4E00, 0x9FFF},
		{"Hangul Syllables", 0xAC00, 0xD7AF},
		{"Private Use Area", 0xE000, 0xF8FF},
		{"Alphabetic Presentation Forms", 0xFB00, 0xFB4F},
		{"Arabic Presentation Forms-A", 0xFB50, 0xFDFF},
		{"Combining Half Marks", 0xFE20, 0xFE2F},
		{"CJK Compatibility Forms", 0xFE30, 0xFE4F},
		{"Small Form Variants", 0xFE50, 0xFE6F},
		{"Arabic Presentation Forms-B", 0xFE70, 0xFEFF},
		{"Halfwidth and Fullwidth Forms", 0xFF00, 0xFFEF},
		{"Emoticons", 0x1F600, 0x1F64F},
		{"Miscellaneous Symbols and Pictographs", 0x1F300, 0x1F5FF},
		{"Transport and Map Symbols", 0x1F680, 0x1F6FF},
		{"Supplemental Symbols and Pictographs", 0x1F900, 0x1F9FF},
	}

	// Initialize category tables using Go's built-in Unicode categories
	up.categoryTables = map[string]*unicode.RangeTable{
		"Letters (Uppercase)":       unicode.Lu,
		"Letters (Lowercase)":       unicode.Ll,
		"Letters (Titlecase)":       unicode.Lt,
		"Letters (Modifier)":        unicode.Lm,
		"Letters (Other)":           unicode.Lo,
		"Marks (Nonspacing)":        unicode.Mn,
		"Marks (Spacing Combining)": unicode.Mc,
		"Marks (Enclosing)":         unicode.Me,
		"Numbers (Decimal)":         unicode.Nd,
		"Numbers (Letter)":          unicode.Nl,
		"Numbers (Other)":           unicode.No,
		"Punctuation (Connector)":   unicode.Pc,
		"Punctuation (Dash)":        unicode.Pd,
		"Punctuation (Open)":        unicode.Ps,
		"Punctuation (Close)":       unicode.Pe,
		"Punctuation (Initial)":     unicode.Pi,
		"Punctuation (Final)":       unicode.Pf,
		"Punctuation (Other)":       unicode.Po,
		"Symbols (Math)":            unicode.Sm,
		"Symbols (Currency)":        unicode.Sc,
		"Symbols (Modifier)":        unicode.Sk,
		"Symbols (Other)":           unicode.So,
	}

	up.logger.Success("Initialized Unicode data with %d blocks and %d categories", len(up.unicodeBlocks), len(up.categoryTables))
}

// loadConfig loads character sets from the configuration file
func (up *IntegratedUnicodePicker) loadConfig() {
	// Set default config path if not provided
	if up.configPath == "" {
		homeDir, _ := os.UserHomeDir()
		up.configPath = filepath.Join(homeDir, ".recuerdo", "character_sets.json")
	}

	// Create default config if file doesn't exist
	if _, err := os.Stat(up.configPath); os.IsNotExist(err) {
		up.createDefaultConfig()
	}

	// Load configuration
	data, err := ioutil.ReadFile(up.configPath)
	if err != nil {
		up.logger.Error("Failed to read config file: %v", err)
		up.createDefaultConfig()
		return
	}

	up.config = &UnicodePickerConfig{}
	if err := json.Unmarshal(data, up.config); err != nil {
		up.logger.Error("Failed to parse config file: %v", err)
		up.createDefaultConfig()
		return
	}

	up.logger.Success("Loaded character picker config with %d sets", len(up.config.CharacterSets))
}

// createDefaultConfig creates a default character picker configuration
func (up *IntegratedUnicodePicker) createDefaultConfig() {
	up.config = &UnicodePickerConfig{
		Version:     "1.0",
		Description: "Recuerdo Unicode Character Picker Configuration",
		CharacterSets: []CharacterSet{
			{
				Name:        "Spanish",
				Description: "Common Spanish characters",
				Characters:  []string{"á", "é", "í", "ó", "ú", "ñ", "ü", "Á", "É", "Í", "Ó", "Ú", "Ñ", "Ü", "¿", "¡"},
			},
			{
				Name:        "French",
				Description: "Common French characters",
				Characters:  []string{"à", "â", "ç", "è", "é", "ê", "ë", "î", "ï", "ô", "ù", "û", "ü", "ÿ", "À", "Â", "Ç", "È", "É", "Ê", "Ë", "Î", "Ï", "Ô", "Ù", "Û", "Ü", "Ÿ", "œ", "Œ", "æ", "Æ"},
			},
			{
				Name:        "German",
				Description: "Common German characters",
				Characters:  []string{"ä", "ö", "ü", "ß", "Ä", "Ö", "Ü"},
			},
			{
				Name:        "Portuguese",
				Description: "Common Portuguese characters",
				Characters:  []string{"á", "â", "à", "ã", "é", "ê", "í", "ó", "ô", "õ", "ú", "ç", "Á", "Â", "À", "Ã", "É", "Ê", "Í", "Ó", "Ô", "Õ", "Ú", "Ç"},
			},
			{
				Name:        "Italian",
				Description: "Common Italian characters",
				Characters:  []string{"à", "è", "é", "ì", "í", "î", "ò", "ó", "ù", "ú", "À", "È", "É", "Ì", "Í", "Î", "Ò", "Ó", "Ù", "Ú"},
			},
			{
				Name:        "Math & Symbols",
				Description: "Mathematical and special symbols",
				Characters:  []string{"±", "×", "÷", "≈", "≠", "≤", "≥", "∞", "π", "Σ", "Δ", "√", "∫", "∂", "€", "£", "¥", "¢", "°", "µ", "α", "β", "γ", "θ", "λ", "φ", "ω"},
			},
			{
				Name:        "Punctuation",
				Description: "Extended punctuation marks",
				Characters:  []string{"–", "—", "\u201c", "\u201d", "\u2018", "\u2019", "…", "‚", "„", "‹", "›", "«", "»", "¨", "´", "`", "ˆ", "˜", "¯", "˙", "¸", "˚", "˝"},
			},
		},
	}

	// Save default config
	up.saveConfig()
}

// saveConfig saves the current configuration to file
func (up *IntegratedUnicodePicker) saveConfig() {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(up.configPath), 0755); err != nil {
		up.logger.Error("Failed to create config directory: %v", err)
		return
	}

	data, err := json.MarshalIndent(up.config, "", "  ")
	if err != nil {
		up.logger.Error("Failed to marshal config: %v", err)
		return
	}

	if err := ioutil.WriteFile(up.configPath, data, 0644); err != nil {
		up.logger.Error("Failed to write config file: %v", err)
		return
	}

	up.logger.Success("Saved character picker config to: %s", up.configPath)
}

// setupBasicUI creates the basic UI structure without heavy components
func (up *IntegratedUnicodePicker) setupBasicUI() {
	up.SetMaximumHeight(350)
	up.SetStyleSheet(`
		QWidget {
			background-color: #f8f8f8;
			border: 1px solid #ddd;
			border-radius: 6px;
		}
		QPushButton {
			background-color: white;
			border: 1px solid #ccc;
			border-radius: 3px;
			padding: 4px;
			margin: 1px;
			font-size: 13px;
			min-width: 28px;
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

	mainLayout := widgets.NewQVBoxLayout()
	mainLayout.SetContentsMargins(8, 8, 8, 8)
	up.SetLayout(mainLayout)

	// Mode selection buttons
	modeLayout := widgets.NewQHBoxLayout()
	up.modeButtons = widgets.NewQButtonGroup(up)

	up.quickButton = widgets.NewQPushButton2("Quick Access", up)
	up.quickButton.SetCheckable(true)
	up.quickButton.SetChecked(true)
	up.modeButtons.AddButton(up.quickButton, 0)
	modeLayout.AddWidget(up.quickButton, 0, 0)

	up.browseButton = widgets.NewQPushButton2("Browse Unicode", up)
	up.browseButton.SetCheckable(true)
	up.modeButtons.AddButton(up.browseButton, 1)
	modeLayout.AddWidget(up.browseButton, 0, 0)

	modeLayout.AddStretch(0)
	mainLayout.AddLayout(modeLayout, 0)

	// Stacked widget for different modes
	up.stackedWidget = widgets.NewQStackedWidget(up)
	mainLayout.AddWidget(up.stackedWidget, 1, 0)

	// Connect mode buttons with lazy loading
	up.modeButtons.ConnectButtonClicked2(func(id int) {
		up.stackedWidget.SetCurrentIndex(id)
		if id == 0 {
			up.logger.Debug("Switched to Quick Access mode")
			up.ensureQuickAccessPageLoaded()
		} else {
			up.logger.Debug("Switched to Browse Unicode mode")
			up.ensureBrowseUnicodePageLoaded()
			up.stackedWidget.SetCurrentIndex(1)
		}
	})

	// Create initial quick access page only
	up.ensureQuickAccessPageLoaded()

	up.logger.Success("Basic Unicode picker UI created (lazy loading enabled)")
}

// ensureQuickAccessPageLoaded creates the quick access page if not already loaded
func (up *IntegratedUnicodePicker) ensureQuickAccessPageLoaded() {
	if up.quickPageInitialized {
		return
	}

	up.logger.Debug("Lazy loading Quick Access page...")

	up.quickCharsPage = widgets.NewQWidget(nil, 0)

	scrollArea := widgets.NewQScrollArea(up.quickCharsPage)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetVerticalScrollBarPolicy(core.Qt__ScrollBarAsNeeded)
	scrollArea.SetHorizontalScrollBarPolicy(core.Qt__ScrollBarAsNeeded)

	scrollWidget := widgets.NewQWidget(nil, 0)
	scrollLayout := widgets.NewQVBoxLayout()
	scrollWidget.SetLayout(scrollLayout)

	// Add character sets
	for _, charSet := range up.config.CharacterSets {
		if len(charSet.Characters) == 0 {
			continue
		}

		// Set title
		titleLabel := widgets.NewQLabel2(charSet.Name, scrollWidget, 0)
		titleLabel.SetStyleSheet("font-weight: bold; color: #333; margin: 4px 0px 2px 0px;")
		scrollLayout.AddWidget(titleLabel, 0, 0)

		// Characters grid
		charWidget := widgets.NewQWidget(scrollWidget, 0)
		charLayout := widgets.NewQGridLayout2()
		charLayout.SetSpacing(2)
		charWidget.SetLayout(charLayout)

		maxCols := 14
		row, col := 0, 0

		for _, char := range charSet.Characters {
			if strings.TrimSpace(char) == "" {
				continue
			}

			button := widgets.NewQPushButton2(char, charWidget)
			button.SetToolTip(fmt.Sprintf("Insert: %s (U+%04X)", char, []rune(char)[0]))

			// Capture character in closure
			character := char
			button.ConnectClicked(func(checked bool) {
				up.insertCharacter(character)
			})

			charLayout.AddWidget3(button, row, col, 1, 1, 0)

			col++
			if col >= maxCols {
				col = 0
				row++
			}
		}

		scrollLayout.AddWidget(charWidget, 0, 0)
	}

	scrollLayout.AddStretch(0)
	scrollArea.SetWidget(scrollWidget)

	pageLayout := widgets.NewQVBoxLayout()
	pageLayout.SetContentsMargins(0, 0, 0, 0)
	up.quickCharsPage.SetLayout(pageLayout)
	pageLayout.AddWidget(scrollArea, 1, 0)

	up.stackedWidget.AddWidget(up.quickCharsPage)
	up.quickPageInitialized = true

	up.logger.Success("Quick Access page loaded")
}

// ensureBrowseUnicodePageLoaded creates the browse unicode page if not already loaded
func (up *IntegratedUnicodePicker) ensureBrowseUnicodePageLoaded() {
	if up.browsePageInitialized {
		return
	}

	up.logger.Debug("Lazy loading Browse Unicode page...")

	// Ensure Unicode data is loaded
	up.ensureUnicodeDataLoaded()

	up.browseUnicPage = widgets.NewQWidget(nil, 0)

	layout := widgets.NewQVBoxLayout()
	layout.SetContentsMargins(4, 4, 4, 4)
	up.browseUnicPage.SetLayout(layout)

	// Simple controls layout
	controlsLayout := widgets.NewQVBoxLayout()

	// Unicode codepoint input (primary method)
	codepointLayout := widgets.NewQHBoxLayout()
	codepointLayout.AddWidget(widgets.NewQLabel2("Unicode Codepoint:", up.browseUnicPage, 0), 0, 0)
	up.rangeFromEdit = widgets.NewQLineEdit(up.browseUnicPage)
	up.rangeFromEdit.SetPlaceholderText("Enter hex code (e.g., 00E1, 1F600)")
	up.rangeFromEdit.SetMaximumWidth(150)
	codepointLayout.AddWidget(up.rangeFromEdit, 0, 0)

	insertCodepointBtn := widgets.NewQPushButton2("Insert", up.browseUnicPage)
	insertCodepointBtn.ConnectClicked(func(checked bool) {
		hexText := up.rangeFromEdit.Text()
		if hexText != "" {
			up.insertCharacterFromHex(hexText)
		}
	})
	codepointLayout.AddWidget(insertCodepointBtn, 0, 0)
	codepointLayout.AddStretch(0)
	controlsLayout.AddLayout(codepointLayout, 0)

	// Unicode block selection
	blockLayout := widgets.NewQHBoxLayout()
	blockLayout.AddWidget(widgets.NewQLabel2("Unicode Block:", up.browseUnicPage, 0), 0, 0)
	up.blockComboBox = widgets.NewQComboBox(up.browseUnicPage)
	up.blockComboBox.AddItem("Select a block...", core.NewQVariant())

	// Add common blocks first
	commonBlocks := []string{
		"Basic Latin", "Latin-1 Supplement", "Latin Extended-A", "Latin Extended-B",
		"Greek and Coptic", "Cyrillic", "Arabic", "Hebrew",
		"CJK Unified Ideographs", "Hiragana", "Katakana",
		"Mathematical Operators", "Arrows", "Box Drawing",
		"General Punctuation", "Currency Symbols", "Emoticons",
	}

	for _, blockName := range commonBlocks {
		for _, block := range up.unicodeBlocks {
			if block.Name == blockName {
				up.blockComboBox.AddItem(block.Name, core.NewQVariant())
				break
			}
		}
	}

	// Add separator and remaining blocks
	up.blockComboBox.InsertSeparator(up.blockComboBox.Count())
	up.blockComboBox.AddItem("--- All Blocks ---", core.NewQVariant())

	for _, block := range up.unicodeBlocks {
		// Skip if already added to common blocks
		isCommon := false
		for _, common := range commonBlocks {
			if block.Name == common {
				isCommon = true
				break
			}
		}
		if !isCommon {
			up.blockComboBox.AddItem(block.Name, core.NewQVariant())
		}
	}

	blockLayout.AddWidget(up.blockComboBox, 0, 0)
	blockLayout.AddStretch(0)
	controlsLayout.AddLayout(blockLayout, 0)

	layout.AddLayout(controlsLayout, 0)

	// Characters display area (initially empty)
	up.browseScrollArea = widgets.NewQScrollArea(up.browseUnicPage)
	up.browseScrollArea.SetWidgetResizable(true)
	up.browseWidget = widgets.NewQWidget(nil, 0)
	up.browseCharsLayout = widgets.NewQGridLayout2()
	up.browseCharsLayout.SetSpacing(2)
	up.browseWidget.SetLayout(up.browseCharsLayout)
	up.browseScrollArea.SetWidget(up.browseWidget)

	// Add placeholder text
	placeholderLabel := widgets.NewQLabel2("Select a Unicode block above to view characters, or enter a codepoint directly.", nil, 0)
	placeholderLabel.SetStyleSheet("color: #666; font-style: italic; padding: 20px; text-align: center;")
	placeholderLabel.SetAlignment(core.Qt__AlignCenter)
	up.browseCharsLayout.AddWidget3(placeholderLabel, 0, 0, 1, 1, 0)

	layout.AddWidget(up.browseScrollArea, 1, 0)

	// Connect block selection
	up.blockComboBox.ConnectCurrentTextChanged(func(text string) {
		if text != "Select a block..." && text != "--- All Blocks ---" && text != "" {
			up.browseUnicodeBlockPaginated(text, 0)
		}
	})

	// Connect codepoint input
	up.rangeFromEdit.ConnectReturnPressed(func() {
		hexText := up.rangeFromEdit.Text()
		if hexText != "" {
			up.insertCharacterFromHex(hexText)
		}
	})

	up.stackedWidget.AddWidget(up.browseUnicPage)

	// If this is the first time loading and we're supposed to show browse mode, switch to it
	if up.stackedWidget.CurrentIndex() == 1 {
		up.stackedWidget.SetCurrentIndex(1)
	}

	up.browsePageInitialized = true
	up.logger.Success("Browse Unicode page loaded with simplified interface")
}

// ensureUnicodeDataLoaded initializes Unicode data if not already loaded
func (up *IntegratedUnicodePicker) ensureUnicodeDataLoaded() {
	if up.unicodeDataInitialized {
		return
	}

	up.logger.Debug("Lazy loading Unicode data...")
	up.initializeUnicodeData()
	up.unicodeDataInitialized = true
	up.logger.Success("Unicode data loaded")
}

// browseUnicodeBlock displays characters from a specific Unicode block
func (up *IntegratedUnicodePicker) browseUnicodeBlock(blockName string) {
	up.ensureUnicodeDataLoaded()

	// Only clear if browse page is initialized
	if up.browsePageInitialized {
		up.clearBrowseResults()
	}

	for _, block := range up.unicodeBlocks {
		if block.Name == blockName {
			up.displayUnicodeRange(block.Start, block.End, fmt.Sprintf("Unicode Block: %s", blockName))
			break
		}
	}
}

// browseUnicodeBlockPaginated displays characters from a Unicode block with pagination
func (up *IntegratedUnicodePicker) browseUnicodeBlockPaginated(blockName string, page int) {
	up.ensureUnicodeDataLoaded()

	if up.browsePageInitialized {
		up.clearBrowseResults()
	}

	var targetBlock *UnicodeBlockInfo
	for i := range up.unicodeBlocks {
		if up.unicodeBlocks[i].Name == blockName {
			targetBlock = &up.unicodeBlocks[i]
			break
		}
	}

	if targetBlock == nil {
		up.logger.Warning("Unicode block not found: %s", blockName)
		return
	}

	const charsPerPage = 64 // 8x8 grid
	totalChars := int(targetBlock.End - targetBlock.Start + 1)
	totalPages := (totalChars + charsPerPage - 1) / charsPerPage

	startIdx := page * charsPerPage
	endIdx := startIdx + charsPerPage
	if endIdx > totalChars {
		endIdx = totalChars
	}

	// Display page info
	pageInfo := widgets.NewQLabel2(fmt.Sprintf("Block: %s | Page %d of %d | Characters %d-%d of %d",
		blockName, page+1, totalPages, startIdx+1, endIdx, totalChars), nil, 0)
	pageInfo.SetStyleSheet("font-weight: bold; padding: 5px; background-color: #f0f0f0; border: 1px solid #ccc;")
	up.browseCharsLayout.AddWidget3(pageInfo, 0, 0, 1, 8, 0)

	// Add pagination controls if more than one page
	if totalPages > 1 {
		paginationLayout := widgets.NewQHBoxLayout()

		prevBtn := widgets.NewQPushButton2("← Previous", nil)
		prevBtn.SetEnabled(page > 0)
		if page > 0 {
			prevBtn.ConnectClicked(func(checked bool) {
				up.browseUnicodeBlockPaginated(blockName, page-1)
			})
		}
		paginationLayout.AddWidget(prevBtn, 0, 0)

		pageLabel := widgets.NewQLabel2(fmt.Sprintf("Page %d of %d", page+1, totalPages), nil, 0)
		pageLabel.SetAlignment(core.Qt__AlignCenter)
		paginationLayout.AddWidget(pageLabel, 1, 0)

		nextBtn := widgets.NewQPushButton2("Next →", nil)
		nextBtn.SetEnabled(page < totalPages-1)
		if page < totalPages-1 {
			nextBtn.ConnectClicked(func(checked bool) {
				up.browseUnicodeBlockPaginated(blockName, page+1)
			})
		}
		paginationLayout.AddWidget(nextBtn, 0, 0)

		paginationWidget := widgets.NewQWidget(nil, 0)
		paginationWidget.SetLayout(paginationLayout)
		up.browseCharsLayout.AddWidget3(paginationWidget, 1, 0, 1, 8, 0)
	}

	// Display characters for this page
	row := 2 // Start after info and pagination
	col := 0
	charCount := 0

	for i := startIdx; i < endIdx; i++ {
		char := rune(int(targetBlock.Start) + i)

		// Skip non-graphic characters
		if !unicode.IsGraphic(char) {
			continue
		}

		button := widgets.NewQPushButton2(string(char), nil)
		button.SetMinimumSize2(32, 32)
		button.SetMaximumSize2(32, 32)
		button.SetStyleSheet(`
			QPushButton {
				font-size: 14px;
				border: 1px solid #ccc;
				background-color: white;
			}
			QPushButton:hover {
				background-color: #e0e0e0;
				border: 2px solid #0078d4;
			}
			QPushButton:pressed {
				background-color: #0078d4;
				color: white;
			}
		`)

		// Set tooltip with character info
		hexCode := fmt.Sprintf("U+%04X", char)
		charName := up.getUnicodeCharName(char)
		if charName == "" {
			charName = "Unknown"
		}
		button.SetToolTip(fmt.Sprintf("%s\n%s\nClick to insert", hexCode, charName))

		// Connect click handler
		charToInsert := string(char)
		button.ConnectClicked(func(checked bool) {
			up.insertCharacter(charToInsert)
		})

		up.browseCharsLayout.AddWidget3(button, row, col, 1, 1, 0)

		col++
		if col >= 8 { // 8 columns
			col = 0
			row++
		}
		charCount++
	}

	up.logger.Debug("Displayed page %d of Unicode block '%s' (%d characters)", page+1, blockName, charCount)
}

// browseUnicodeCategory displays characters from a Unicode category
func (up *IntegratedUnicodePicker) browseUnicodeCategory(categoryName string) {
	up.ensureUnicodeDataLoaded()

	// Only clear if browse page is initialized
	if up.browsePageInitialized {
		up.clearBrowseResults()
	}

	rangeTable, exists := up.categoryTables[categoryName]
	if !exists {
		return
	}

	var chars []rune

	// Iterate through the range table to find characters
	for _, r16 := range rangeTable.R16 {
		for r := rune(r16.Lo); r <= rune(r16.Hi); r += rune(r16.Stride) {
			if unicode.IsPrint(r) && utf8.ValidRune(r) && !unicode.IsControl(r) {
				chars = append(chars, r)
				if len(chars) >= 200 { // Reduced limit for better performance
					break
				}
			}
		}
		if len(chars) >= 200 {
			break
		}
	}

	for _, r32 := range rangeTable.R32 {
		for r := rune(r32.Lo); r <= rune(r32.Hi); r += rune(r32.Stride) {
			if unicode.IsPrint(r) && utf8.ValidRune(r) && !unicode.IsControl(r) {
				chars = append(chars, r)
				if len(chars) >= 200 { // Reduced limit for better performance
					break
				}
			}
		}
		if len(chars) >= 200 {
			break
		}
	}

	up.displaySearchResults(chars, fmt.Sprintf("Unicode Category: %s", categoryName))
}

// browseUnicodeRange displays characters from a custom range
func (up *IntegratedUnicodePicker) browseUnicodeRange() {
	fromText := strings.TrimSpace(up.rangeFromEdit.Text())
	toText := strings.TrimSpace(up.rangeToEdit.Text())

	if fromText == "" || toText == "" {
		return
	}

	// Parse hex values
	fromVal, err1 := strconv.ParseInt(fromText, 16, 32)
	toVal, err2 := strconv.ParseInt(toText, 16, 32)

	if err1 != nil || err2 != nil {
		return
	}

	if fromVal > toVal || fromVal < 0 || toVal > 0x10FFFF {
		return
	}

	// Only clear if browse page is initialized
	if up.browsePageInitialized {
		up.clearBrowseResults()
	}
	up.displayUnicodeRange(rune(fromVal), rune(toVal), fmt.Sprintf("Range: U+%04X - U+%04X", fromVal, toVal))
}

// performSearch searches for characters by name or code
func (up *IntegratedUnicodePicker) performSearch(query string) {
	query = strings.TrimSpace(strings.ToLower(query))
	if len(query) < 2 {
		return
	}

	// Only clear if browse page is initialized
	if up.browsePageInitialized {
		up.clearBrowseResults()
	}

	var results []rune

	// Check if it's a hex search
	if strings.HasPrefix(query, "u+") {
		query = query[2:]
	}

	// Try to parse as hex
	if hexVal, err := strconv.ParseInt(query, 16, 32); err == nil && hexVal <= 0x10FFFF {
		r := rune(hexVal)
		if unicode.IsPrint(r) && utf8.ValidRune(r) {
			results = append(results, r)
		}
	} else {
		// Search by character name patterns
		searchPatterns := map[string][]rune{
			"alpha":    {0x03B1, 0x0391},
			"beta":     {0x03B2, 0x0392},
			"gamma":    {0x03B3, 0x0393},
			"delta":    {0x03B4, 0x0394},
			"epsilon":  {0x03B5, 0x0395},
			"zeta":     {0x03B6, 0x0396},
			"eta":      {0x03B7, 0x0397},
			"theta":    {0x03B8, 0x0398},
			"iota":     {0x03B9, 0x0399},
			"kappa":    {0x03BA, 0x039A},
			"lambda":   {0x03BB, 0x039B},
			"mu":       {0x03BC, 0x039C},
			"nu":       {0x03BD, 0x039D},
			"xi":       {0x03BE, 0x039E},
			"omicron":  {0x03BF, 0x039F},
			"pi":       {0x03C0, 0x03A0},
			"rho":      {0x03C1, 0x03A1},
			"sigma":    {0x03C3, 0x03A3, 0x03C2},
			"tau":      {0x03C4, 0x03A4},
			"upsilon":  {0x03C5, 0x03A5},
			"phi":      {0x03C6, 0x03A6},
			"chi":      {0x03C7, 0x03A7},
			"psi":      {0x03C8, 0x03A8},
			"omega":    {0x03C9, 0x03A9},
			"euro":     {0x20AC},
			"pound":    {0x00A3},
			"yen":      {0x00A5},
			"dollar":   {0x0024},
			"cent":     {0x00A2},
			"degree":   {0x00B0},
			"infinity": {0x221E},
			"arrow":    {0x2190, 0x2191, 0x2192, 0x2193},
			"heart":    {0x2665, 0x2764},
			"star":     {0x2605, 0x2606},
			"music":    {0x266A, 0x266B},
		}

		for pattern, runes := range searchPatterns {
			if strings.Contains(pattern, query) {
				results = append(results, runes...)
			}
		}

		// Also search in our configured character sets
		for _, charSet := range up.config.CharacterSets {
			if strings.Contains(strings.ToLower(charSet.Name), query) {
				for _, char := range charSet.Characters {
					if len(char) > 0 {
						results = append(results, []rune(char)[0])
					}
				}
			}
		}
	}

	// Remove duplicates and sort
	uniqueResults := make(map[rune]bool)
	var finalResults []rune
	for _, r := range results {
		if !uniqueResults[r] {
			uniqueResults[r] = true
			finalResults = append(finalResults, r)
		}
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i] < finalResults[j]
	})

	up.displaySearchResults(finalResults, fmt.Sprintf("Search results for: %s", query))
}

// clearBrowseResults clears the browse results area
func (up *IntegratedUnicodePicker) clearBrowseResults() {
	if up.browseCharsLayout == nil || up.browseWidget == nil {
		return
	}

	// Simple approach: delete the browse widget and recreate it
	up.browseWidget.DeleteLater()
	up.browseWidget = widgets.NewQWidget(nil, 0)
	up.browseCharsLayout = widgets.NewQGridLayout2()
	up.browseCharsLayout.SetSpacing(2)
	up.browseWidget.SetLayout(up.browseCharsLayout)
	up.browseScrollArea.SetWidget(up.browseWidget)
}

// displayUnicodeRange displays a range of Unicode characters
func (up *IntegratedUnicodePicker) displayUnicodeRange(startRune, endRune rune, title string) {
	var chars []rune

	for r := startRune; r <= endRune; r++ {
		if unicode.IsPrint(r) && utf8.ValidRune(r) && !unicode.IsControl(r) {
			chars = append(chars, r)
		}
	}

	up.displaySearchResults(chars, title)
}

// displaySearchResults displays a list of Unicode characters
func (up *IntegratedUnicodePicker) displaySearchResults(chars []rune, title string) {
	if len(chars) == 0 {
		noResultsLabel := widgets.NewQLabel2("No characters found", up.browseWidget, 0)
		noResultsLabel.SetStyleSheet("color: #666; font-style: italic; padding: 20px;")
		up.browseCharsLayout.AddWidget3(noResultsLabel, 0, 0, 1, 1, 0)
		return
	}

	// Add title if provided
	if title != "" {
		titleLabel := widgets.NewQLabel2(fmt.Sprintf("%s (%d characters)", title, len(chars)), up.browseWidget, 0)
		titleLabel.SetStyleSheet("font-weight: bold; color: #333; margin: 4px 0px;")
		up.browseCharsLayout.AddWidget3(titleLabel, 0, 0, 1, 14, 0)
	}

	maxCols := 14
	row, col := 1, 0

	for _, r := range chars {
		char := string(r)
		button := widgets.NewQPushButton2(char, up.browseWidget)

		// Detailed tooltip with Unicode information
		unicodeName := up.getUnicodeCharName(r)
		button.SetToolTip(fmt.Sprintf("Character: %s\nUnicode: U+%04X\nUTF-8: %X\nName: %s", char, r, []byte(char), unicodeName))

		// Capture character in closure
		character := char
		button.ConnectClicked(func(checked bool) {
			up.insertCharacter(character)
		})

		up.browseCharsLayout.AddWidget3(button, row, col, 1, 1, 0)

		col++
		if col >= maxCols {
			col = 0
			row++
		}
	}

	up.logger.Debug("Displayed %d Unicode characters: %s", len(chars), title)
}

// getUnicodeCharName returns a descriptive name for a Unicode character
func (up *IntegratedUnicodePicker) getUnicodeCharName(r rune) string {
	// Basic character identification
	if r >= 'A' && r <= 'Z' {
		return fmt.Sprintf("Latin Capital Letter %c", r)
	}
	if r >= 'a' && r <= 'z' {
		return fmt.Sprintf("Latin Small Letter %c", r)
	}
	if r >= '0' && r <= '9' {
		return fmt.Sprintf("Digit %c", r)
	}

	// Common Unicode characters with descriptive names
	commonNames := map[rune]string{
		0x0020: "Space",
		0x00A0: "Non-Breaking Space",
		0x00A1: "Inverted Exclamation Mark",
		0x00A3: "Pound Sign",
		0x00A5: "Yen Sign",
		0x00A9: "Copyright Sign",
		0x00AE: "Registered Sign",
		0x00B0: "Degree Sign",
		0x00B5: "Micro Sign",
		0x00BF: "Inverted Question Mark",
		0x20AC: "Euro Sign",
		0x2190: "Leftwards Arrow",
		0x2191: "Upwards Arrow",
		0x2192: "Rightwards Arrow",
		0x2193: "Downwards Arrow",
		0x221E: "Infinity",
		0x2665: "Black Heart Suit",
		0x2764: "Heavy Black Heart",
		0x03B1: "Greek Small Letter Alpha",
		0x03B2: "Greek Small Letter Beta",
		0x03C0: "Greek Small Letter Pi",
		0x03A9: "Greek Capital Letter Omega",
	}

	if name, exists := commonNames[r]; exists {
		return name
	}

	// Use Unicode category as fallback
	if unicode.IsLetter(r) {
		return "Letter"
	}
	if unicode.IsDigit(r) {
		return "Digit"
	}
	if unicode.IsPunct(r) {
		return "Punctuation"
	}
	if unicode.IsSymbol(r) {
		return "Symbol"
	}
	if unicode.IsMark(r) {
		return "Mark"
	}

	return "Character"
}

// insertCharacter inserts a character into the target edit widget
func (up *IntegratedUnicodePicker) insertCharacter(char string) {
	if up.targetEdit == nil {
		up.logger.Warning("No target edit widget set")
		return
	}

	cursorPos := up.targetEdit.CursorPosition()
	currentText := up.targetEdit.Text()

	// Convert to runes for proper Unicode handling
	currentRunes := []rune(currentText)
	charRunes := []rune(char)

	// Ensure cursor position is within bounds
	if cursorPos > len(currentRunes) {
		cursorPos = len(currentRunes)
	}
	if cursorPos < 0 {
		cursorPos = 0
	}

	// Insert character at cursor position using rune slicing
	newRunes := make([]rune, 0, len(currentRunes)+len(charRunes))
	newRunes = append(newRunes, currentRunes[:cursorPos]...)
	newRunes = append(newRunes, charRunes...)
	newRunes = append(newRunes, currentRunes[cursorPos:]...)

	newText := string(newRunes)
	up.targetEdit.SetText(newText)

	// Move cursor to after inserted character (in rune positions)
	newCursorPos := cursorPos + len(charRunes)
	up.targetEdit.SetCursorPosition(newCursorPos)

	// Focus back on edit widget
	up.targetEdit.SetFocus2()

	up.logger.Info("Inserted '%s' at position %d (runes: %d)", char, cursorPos, len(charRunes))
}

// SetTargetEdit sets the target line edit for character insertion
func (up *IntegratedUnicodePicker) SetTargetEdit(edit *widgets.QLineEdit) {
	up.targetEdit = edit
	up.logger.Debug("Set target edit widget")
}

// ReloadConfig reloads the configuration file
func (up *IntegratedUnicodePicker) ReloadConfig() {
	up.loadConfig()

	// Reset lazy loading flags and recreate quick access page if it was loaded
	if up.quickPageInitialized {
		up.stackedWidget.RemoveWidget(up.quickCharsPage)
		up.quickCharsPage.DeleteLater()
		up.quickPageInitialized = false
		up.ensureQuickAccessPageLoaded()
	}

	up.logger.Success("Reloaded Unicode picker configuration")
}

// GetConfigPath returns the current configuration file path
func (up *IntegratedUnicodePicker) GetConfigPath() string {
	return up.configPath
}

// insertCharacterFromHex inserts a character from a hex code string
func (up *IntegratedUnicodePicker) insertCharacterFromHex(hexText string) {
	// Clean the hex string
	hexText = strings.TrimSpace(strings.ToUpper(hexText))
	hexText = strings.TrimPrefix(hexText, "U+")
	hexText = strings.TrimPrefix(hexText, "0X")

	if hexText == "" {
		up.logger.Warning("Empty hex code provided")
		return
	}

	// Parse hex code
	codepoint, err := strconv.ParseInt(hexText, 16, 32)
	if err != nil {
		up.logger.Warning("Invalid hex code: %s", hexText)
		return
	}

	if codepoint < 0 || codepoint > 0x10FFFF {
		up.logger.Warning("Codepoint out of valid Unicode range: U+%s", hexText)
		return
	}

	char := rune(codepoint)
	if !unicode.IsGraphic(char) && !unicode.IsSpace(char) {
		up.logger.Warning("Codepoint U+%s is not a graphic character", hexText)
		return
	}

	up.insertCharacter(string(char))
	up.logger.Debug("Inserted character from hex code U+%s: %c", hexText, char)

	// Clear the input field for next use
	up.rangeFromEdit.Clear()
}
