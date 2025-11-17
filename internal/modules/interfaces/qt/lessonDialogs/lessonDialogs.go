// Package lessonDialogs provides functionality ported from Python module
//
// Provides dialogs for lesson management including new lesson creation,
// lesson properties, and lesson import/export dialogs.
//
// This is an automated port - implementation may be incomplete.
package lessonDialogs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/mappu/miqt/qt"
)

// LessonDialogsModule is a Go port of the Python LessonDialogsModule class
type LessonDialogsModule struct {
	*core.BaseModule
	manager          *core.Manager
	newLessonDialog  *qt.QDialog
	propertiesDialog *qt.QDialog
	importDialog     *qt.QDialog

	// Widget references for new lesson dialog
	nameEdit          *qt.QLineEdit
	descEdit          *qt.QTextEdit
	wordsRadio        *qt.QRadioButton
	topoRadio         *qt.QRadioButton
	mediaRadio        *qt.QRadioButton
	questionLangCombo *qt.QComboBox
	answerLangCombo   *qt.QComboBox

	// Widget references for properties dialog
	propNameEdit    *qt.QLineEdit
	propDescEdit    *qt.QTextEdit
	propAuthorEdit  *qt.QLineEdit
	propVersionEdit *qt.QLineEdit
	itemCountLabel  *qt.QLabel

	// Widget references for import dialog
	importFileEdit *qt.QLineEdit
	encodingCombo  *qt.QComboBox
	separatorCombo *qt.QComboBox
	firstRowCheck  *qt.QCheckBox
}

// NewLessonDialogsModule creates a new LessonDialogsModule instance
func NewLessonDialogsModule() *LessonDialogsModule {
	base := core.NewBaseModule("lessonDialogs", "lesson-dialogs-module")
	base.SetRequires("qtApp")

	return &LessonDialogsModule{
		BaseModule: base,
	}
}

// ShowNewLessonDialog displays the new lesson creation dialog
func (mod *LessonDialogsModule) ShowNewLessonDialog() map[string]interface{} {
	log.Printf("[SUCCESS] LessonDialogsModule.ShowNewLessonDialog() - creating and showing new lesson dialog")

	if mod.manager == nil {
		log.Printf("[ERROR] LessonDialogsModule.ShowNewLessonDialog() - manager is nil")
		return nil
	}

	// Get the main window as parent
	var parentWidget *qt.QWidget
	uiModules := mod.manager.GetModulesByType("ui")
	if len(uiModules) > 0 {
		if guiMod, ok := uiModules[0].(interface{ GetMainWindow() *qt.QMainWindow }); ok {
			parentWidget = guiMod.GetMainWindow().QWidget
			log.Printf("[SUCCESS] LessonDialogsModule got parent window from GUI module")
		}
	}

	if mod.newLessonDialog == nil {
		mod.createNewLessonDialog(parentWidget)
	}

	// Reset form
	mod.resetNewLessonForm()

	log.Printf("[SUCCESS] LessonDialogsModule showing new lesson dialog")
	result := mod.newLessonDialog.Exec()
	log.Printf("[SUCCESS] LessonDialogsModule dialog closed with result: %d", result)

	if result == int(qt.QDialog__Accepted) {
		data := mod.getNewLessonData()
		log.Printf("[SUCCESS] New lesson dialog returned data: %v", data)
		return data
	} else {
		log.Printf("[INFO] New lesson dialog was cancelled")
	}

	return nil
}

// ShowPropertiesDialog displays the lesson properties dialog
func (mod *LessonDialogsModule) ShowPropertiesDialog(parent *qt.QWidget, lessonData map[string]interface{}) map[string]interface{} {
	if mod.propertiesDialog == nil {
		mod.createPropertiesDialog(parent)
	}

	// Load current lesson data
	mod.loadPropertiesData(lessonData)

	if mod.propertiesDialog.Exec() == int(qt.QDialog__Accepted) {
		return mod.getPropertiesData()
	}

	return nil
}

// ShowImportDialog displays the lesson import dialog
func (mod *LessonDialogsModule) ShowImportDialog(parent *qt.QWidget) map[string]interface{} {
	if mod.importDialog == nil {
		mod.createImportDialog(parent)
	}

	if mod.importDialog.Exec() == int(qt.QDialog__Accepted) {
		return mod.getImportData()
	}

	return nil
}

// createNewLessonDialog creates the new lesson dialog
func (mod *LessonDialogsModule) createNewLessonDialog(parent *qt.QWidget) {
	mod.newLessonDialog = qt.NewQDialog(parent)
	mod.newLessonDialog.SetWindowTitle("Create New Lesson")
	mod.newLessonDialog.SetFixedSize2(400, 350)
	mod.newLessonDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout(mod.newLessonDialog.QWidget)

	// Lesson name
	nameGroup := qt.NewQGroupBox(mod.newLessonDialog.QWidget)
	nameGroup.SetTitle("Lesson Information")
	nameLayout := qt.NewQFormLayout(nameGroup.QWidget)

	mod.nameEdit = qt.NewQLineEdit(nameGroup.QWidget)
	mod.nameEdit.SetObjectName("lessonName")
	mod.nameEdit.SetPlaceholderText("Enter lesson name...")
	nameLayout.AddRow3("Name:", mod.nameEdit.QWidget)

	mod.descEdit = qt.NewQTextEdit(nameGroup.QWidget)
	mod.descEdit.SetObjectName("lessonDescription")
	mod.descEdit.SetPlaceholderText("Enter lesson description...")
	mod.descEdit.SetMaximumHeight(80)
	nameLayout.AddRow3("Description:", mod.descEdit.QWidget)

	layout.AddWidget(nameGroup.QWidget)

	// Lesson type
	typeGroup := qt.NewQGroupBox(mod.newLessonDialog.QWidget)
	typeGroup.SetTitle("Lesson Type")
	typeLayout := qt.NewQVBoxLayout(typeGroup.QWidget)

	mod.wordsRadio = qt.NewQRadioButton(typeGroup.QWidget)
	mod.wordsRadio.SetText("Words List")
	mod.wordsRadio.SetObjectName("wordsRadio")
	mod.wordsRadio.SetChecked(true)
	typeLayout.AddWidget(mod.wordsRadio.QWidget)

	mod.topoRadio = qt.NewQRadioButton(typeGroup.QWidget)
	mod.topoRadio.SetText("Topology")
	mod.topoRadio.SetObjectName("topoRadio")
	typeLayout.AddWidget(mod.topoRadio.QWidget)

	mod.mediaRadio = qt.NewQRadioButton(typeGroup.QWidget)
	mod.mediaRadio.SetText("Media")
	mod.mediaRadio.SetObjectName("mediaRadio")
	typeLayout.AddWidget(mod.mediaRadio.QWidget)

	layout.AddWidget(typeGroup.QWidget)

	// Language settings
	langGroup := qt.NewQGroupBox(mod.newLessonDialog.QWidget)
	langGroup.SetTitle("Languages")
	langLayout := qt.NewQFormLayout(langGroup.QWidget)

	mod.questionLangCombo = qt.NewQComboBox(langGroup.QWidget)
	mod.questionLangCombo.SetObjectName("questionLanguage")
	mod.questionLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	langLayout.AddRow3("Question language:", mod.questionLangCombo.QWidget)

	mod.answerLangCombo = qt.NewQComboBox(langGroup.QWidget)
	mod.answerLangCombo.SetObjectName("answerLanguage")
	mod.answerLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	mod.answerLangCombo.SetCurrentIndex(1) // Default to Dutch
	langLayout.AddRow3("Answer language:", mod.answerLangCombo.QWidget)

	layout.AddWidget(langGroup.QWidget)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox(mod.newLessonDialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	layout.AddWidget(buttonBox.QWidget)

	buttonBox.OnAccepted(func() {
		mod.newLessonDialog.Accept()
	})

	buttonBox.OnRejected(func() {
		mod.newLessonDialog.Reject()
	})
}

// createPropertiesDialog creates the lesson properties dialog
func (mod *LessonDialogsModule) createPropertiesDialog(parent *qt.QWidget) {
	mod.propertiesDialog = qt.NewQDialog(parent)
	mod.propertiesDialog.SetWindowTitle("Lesson Properties")
	mod.propertiesDialog.SetFixedSize2(450, 400)
	mod.propertiesDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout(mod.propertiesDialog.QWidget)

	// Create tab widget
	tabWidget := qt.NewQTabWidget(mod.propertiesDialog.QWidget)
	layout.AddWidget(tabWidget.QWidget)

	// General tab
	generalTab := qt.NewQWidget2()
	generalLayout := qt.NewQFormLayout(generalTab)

	mod.propNameEdit = qt.NewQLineEdit(nil)
	mod.propNameEdit.SetObjectName("propLessonName")
	generalLayout.AddRow3("Name:", mod.propNameEdit.QWidget)

	mod.propDescEdit = qt.NewQTextEdit(nil)
	mod.propDescEdit.SetObjectName("propLessonDescription")
	mod.propDescEdit.SetMaximumHeight(100)
	generalLayout.AddRow3("Description:", mod.propDescEdit.QWidget)

	mod.propAuthorEdit = qt.NewQLineEdit(nil)
	mod.propAuthorEdit.SetObjectName("propAuthor")
	generalLayout.AddRow3("Author:", mod.propAuthorEdit.QWidget)

	mod.propVersionEdit = qt.NewQLineEdit(nil)
	mod.propVersionEdit.SetObjectName("propVersion")
	generalLayout.AddRow3("Version:", mod.propVersionEdit.QWidget)

	tabWidget.AddTab(generalTab, "General")

	// Statistics tab
	statsTab := qt.NewQWidget2()
	statsLayout := qt.NewQFormLayout(statsTab)

	itemCountLabel := qt.NewQLabel2()
	itemCountLabel.SetText("0")
	itemCountLabel.SetObjectName("itemCount")
	statsLayout.AddRow3("Number of items:", itemCountLabel.QWidget)

	createdLabel := qt.NewQLabel2()
	createdLabel.SetText("Unknown")
	createdLabel.SetObjectName("createdDate")
	statsLayout.AddRow3("Created:", createdLabel.QWidget)

	modifiedLabel := qt.NewQLabel2()
	modifiedLabel.SetText("Unknown")
	modifiedLabel.SetObjectName("modifiedDate")
	statsLayout.AddRow3("Last modified:", modifiedLabel.QWidget)

	fileSizeLabel := qt.NewQLabel2()
	fileSizeLabel.SetText("0 KB")
	fileSizeLabel.SetObjectName("fileSize")
	statsLayout.AddRow3("File size:", fileSizeLabel.QWidget)

	tabWidget.AddTab(statsTab, "Statistics")

	// Buttons
	buttonBox := qt.NewQDialogButtonBox(mod.propertiesDialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	layout.AddWidget(buttonBox.QWidget)

	buttonBox.OnAccepted(func() {
		mod.propertiesDialog.Accept()
	})

	buttonBox.OnRejected(func() {
		mod.propertiesDialog.Reject()
	})
}

// createImportDialog creates the import dialog
func (mod *LessonDialogsModule) createImportDialog(parent *qt.QWidget) {
	mod.importDialog = qt.NewQDialog(parent)
	mod.importDialog.SetWindowTitle("Import Lesson")
	mod.importDialog.SetFixedSize2(500, 300)
	mod.importDialog.SetWindowModality(qt.ApplicationModal)

	layout := qt.NewQVBoxLayout(mod.importDialog.QWidget)

	// File selection
	fileGroup := qt.NewQGroupBox(mod.importDialog.QWidget)
	fileGroup.SetTitle("Import File")
	fileLayout := qt.NewQHBoxLayout(fileGroup.QWidget)

	mod.importFileEdit = qt.NewQLineEdit(fileGroup.QWidget)
	mod.importFileEdit.SetObjectName("filePath")
	mod.importFileEdit.SetPlaceholderText("Select file to import...")
	fileLayout.AddWidget(mod.importFileEdit.QWidget)

	browseBtn := qt.NewQPushButton2()
	browseBtn.SetText("Browse...")
	browseBtn.OnClicked(func() {
		fileName := qt.QFileDialog_GetOpenFileName4(mod.importDialog.QWidget,
			"Select lesson file",
			"",
			"All supported files (*.ot *.txt *.csv);;OpenTeacher files (*.ot);;Text files (*.txt);;CSV files (*.csv);;All files (*.*)")
		if fileName != "" {
			mod.importFileEdit.SetText(fileName)
		}
	})
	fileLayout.AddWidget(browseBtn.QWidget)

	layout.AddWidget(fileGroup.QWidget)

	// Import options
	optionsGroup := qt.NewQGroupBox(mod.importDialog.QWidget)
	optionsGroup.SetTitle("Import Options")
	optionsLayout := qt.NewQVBoxLayout(optionsGroup.QWidget)

	encodingLayout := qt.NewQHBoxLayout2()
	encodingLabel := qt.NewQLabel2()
	encodingLabel.SetText("File encoding:")
	encodingLayout.AddWidget(encodingLabel.QWidget)

	mod.encodingCombo = qt.NewQComboBox(optionsGroup.QWidget)
	mod.encodingCombo.SetObjectName("encoding")
	mod.encodingCombo.AddItems([]string{"UTF-8", "UTF-16", "ISO-8859-1", "ASCII"})
	encodingLayout.AddWidget(mod.encodingCombo.QWidget)
	optionsLayout.AddLayout2(encodingLayout.QLayout, 0)

	separatorLayout := qt.NewQHBoxLayout2()
	separatorLabel := qt.NewQLabel2()
	separatorLabel.SetText("Field separator:")
	separatorLayout.AddWidget(separatorLabel.QWidget)

	mod.separatorCombo = qt.NewQComboBox(optionsGroup.QWidget)
	mod.separatorCombo.SetObjectName("separator")
	mod.separatorCombo.AddItems([]string{"Tab", "Comma", "Semicolon", "Space"})
	separatorLayout.AddWidget(mod.separatorCombo.QWidget)
	optionsLayout.AddLayout2(separatorLayout.QLayout, 0)

	mod.firstRowCheck = qt.NewQCheckBox2()
	mod.firstRowCheck.SetText("First row contains headers")
	mod.firstRowCheck.SetObjectName("firstRowHeaders")
	mod.firstRowCheck.SetChecked(true)
	optionsLayout.AddWidget(mod.firstRowCheck.QWidget)

	layout.AddWidget(optionsGroup.QWidget)

	// Preview area
	previewGroup := qt.NewQGroupBox(mod.importDialog.QWidget)
	previewGroup.SetTitle("Preview")
	previewLayout := qt.NewQVBoxLayout(previewGroup.QWidget)

	previewText := qt.NewQTextEdit(previewGroup.QWidget)
	previewText.SetObjectName("previewText")
	previewText.SetReadOnly(true)
	previewText.SetMaximumHeight(150)
	previewText.SetPlainText("Select a file to see preview...")

	previewLayout.AddWidget(previewText.QWidget)

	layout.AddWidget(previewGroup.QWidget)

	// Buttons
	buttonBox := qt.NewQDialogButtonBox(mod.importDialog.QWidget)
	buttonBox.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	layout.AddWidget(buttonBox.QWidget)

	buttonBox.OnAccepted(func() {
		mod.importDialog.Accept()
	})

	buttonBox.OnRejected(func() {
		mod.importDialog.Reject()
	})
}

// resetNewLessonForm resets the new lesson dialog form
func (mod *LessonDialogsModule) resetNewLessonForm() {
	if mod.newLessonDialog == nil {
		return
	}

	if mod.nameEdit != nil {
		mod.nameEdit.Clear()
	}

	if mod.descEdit != nil {
		mod.descEdit.Clear()
	}

	if mod.wordsRadio != nil {
		mod.wordsRadio.SetChecked(true)
	}
}

// getNewLessonData retrieves data from the new lesson dialog form
func (mod *LessonDialogsModule) getNewLessonData() map[string]interface{} {
	data := make(map[string]interface{})

	if mod.newLessonDialog == nil {
		return nil
	}

	// Get lesson name and description
	if mod.nameEdit != nil {
		data["name"] = mod.nameEdit.Text()
	}

	if mod.descEdit != nil {
		data["description"] = mod.descEdit.ToPlainText()
	}

	// Determine lesson type
	if mod.wordsRadio != nil && mod.wordsRadio.IsChecked() {
		data["type"] = "words"
	} else if mod.topoRadio != nil && mod.topoRadio.IsChecked() {
		data["type"] = "topology"
	} else if mod.mediaRadio != nil && mod.mediaRadio.IsChecked() {
		data["type"] = "media"
	} else {
		data["type"] = "words" // Default
	}

	// Get languages
	if mod.questionLangCombo != nil {
		data["questionLanguage"] = mod.questionLangCombo.CurrentText()
	}

	if mod.answerLangCombo != nil {
		data["answerLanguage"] = mod.answerLangCombo.CurrentText()
	}

	return data
}

// loadPropertiesData loads lesson data into the properties dialog
func (mod *LessonDialogsModule) loadPropertiesData(lessonData map[string]interface{}) {
	if mod.propertiesDialog == nil || lessonData == nil {
		return
	}

	if name, ok := lessonData["name"].(string); ok {
		if mod.propNameEdit != nil {
			mod.propNameEdit.SetText(name)
		}
	}

	if desc, ok := lessonData["description"].(string); ok {
		if mod.propDescEdit != nil {
			mod.propDescEdit.SetPlainText(desc)
		}
	}

	if author, ok := lessonData["author"].(string); ok {
		if mod.propAuthorEdit != nil {
			mod.propAuthorEdit.SetText(author)
		}
	}

	if version, ok := lessonData["version"].(string); ok {
		if mod.propVersionEdit != nil {
			mod.propVersionEdit.SetText(version)
		}
	}

	// Update statistics
	if itemCount, ok := lessonData["itemCount"].(int); ok {
		if mod.itemCountLabel != nil {
			mod.itemCountLabel.SetText(fmt.Sprintf("%d", itemCount))
		}
	}
}

// getPropertiesData extracts data from the properties dialog
func (mod *LessonDialogsModule) getPropertiesData() map[string]interface{} {
	if mod.propertiesDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	if mod.propNameEdit != nil {
		data["name"] = strings.TrimSpace(mod.propNameEdit.Text())
	}

	if mod.propDescEdit != nil {
		data["description"] = strings.TrimSpace(mod.propDescEdit.ToPlainText())
	}

	if mod.propAuthorEdit != nil {
		data["author"] = strings.TrimSpace(mod.propAuthorEdit.Text())
	}

	if mod.propVersionEdit != nil {
		data["version"] = strings.TrimSpace(mod.propVersionEdit.Text())
	}

	return data
}

// getImportData extracts data from the import dialog
func (mod *LessonDialogsModule) getImportData() map[string]interface{} {
	if mod.importDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	if mod.importFileEdit != nil {
		data["file"] = strings.TrimSpace(mod.importFileEdit.Text())
	}

	if mod.encodingCombo != nil {
		data["encoding"] = mod.encodingCombo.CurrentText()
	}

	if mod.separatorCombo != nil {
		data["separator"] = mod.separatorCombo.CurrentText()
	}

	if mod.firstRowCheck != nil {
		data["firstRowHeaders"] = mod.firstRowCheck.IsChecked()
	}

	return data
}

// Enable activates the module
func (mod *LessonDialogsModule) Enable(ctx context.Context) error {
	if err := mod.BaseModule.Enable(ctx); err != nil {
		return err
	}

	fmt.Println("LessonDialogsModule enabled")
	return nil
}

// Disable deactivates the module
func (mod *LessonDialogsModule) Disable(ctx context.Context) error {
	if err := mod.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up dialogs
	if mod.newLessonDialog != nil {
		mod.newLessonDialog.Close()
		mod.newLessonDialog = nil
	}

	if mod.propertiesDialog != nil {
		mod.propertiesDialog.Close()
		mod.propertiesDialog = nil
	}

	if mod.importDialog != nil {
		mod.importDialog.Close()
		mod.importDialog = nil
	}

	fmt.Println("LessonDialogsModule disabled")
	return nil
}

// SetManager sets the module manager
func (mod *LessonDialogsModule) SetManager(manager *core.Manager) {
	mod.manager = manager
}

// InitLessonDialogsModule creates and returns a new LessonDialogsModule instance
func InitLessonDialogsModule() core.Module {
	return NewLessonDialogsModule()
}
