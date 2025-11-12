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
	"strings"

	"github.com/LaPingvino/openteacher/internal/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// LessonDialogsModule is a Go port of the Python LessonDialogsModule class
type LessonDialogsModule struct {
	*core.BaseModule
	manager          *core.Manager
	newLessonDialog  *widgets.QDialog
	propertiesDialog *widgets.QDialog
	importDialog     *widgets.QDialog
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
func (mod *LessonDialogsModule) ShowNewLessonDialog(parent *widgets.QWidget) map[string]interface{} {
	if mod.newLessonDialog == nil {
		mod.createNewLessonDialog(parent)
	}

	// Reset form
	mod.resetNewLessonForm()

	if mod.newLessonDialog.Exec() == int(widgets.QDialog__Accepted) {
		return mod.getNewLessonData()
	}

	return nil
}

// ShowPropertiesDialog displays the lesson properties dialog
func (mod *LessonDialogsModule) ShowPropertiesDialog(parent *widgets.QWidget, lessonData map[string]interface{}) map[string]interface{} {
	if mod.propertiesDialog == nil {
		mod.createPropertiesDialog(parent)
	}

	// Load current lesson data
	mod.loadPropertiesData(lessonData)

	if mod.propertiesDialog.Exec() == int(widgets.QDialog__Accepted) {
		return mod.getPropertiesData()
	}

	return nil
}

// ShowImportDialog displays the lesson import dialog
func (mod *LessonDialogsModule) ShowImportDialog(parent *widgets.QWidget) map[string]interface{} {
	if mod.importDialog == nil {
		mod.createImportDialog(parent)
	}

	if mod.importDialog.Exec() == int(widgets.QDialog__Accepted) {
		return mod.getImportData()
	}

	return nil
}

// createNewLessonDialog creates the new lesson dialog
func (mod *LessonDialogsModule) createNewLessonDialog(parent *widgets.QWidget) {
	mod.newLessonDialog = widgets.NewQDialog(parent, 0)
	mod.newLessonDialog.SetWindowTitle("Create New Lesson")
	mod.newLessonDialog.SetFixedSize2(400, 350)
	mod.newLessonDialog.SetWindowModality(qtcore.Qt__ApplicationModal)

	layout := widgets.NewQVBoxLayout()
	mod.newLessonDialog.SetLayout(layout)

	// Lesson name
	nameGroup := widgets.NewQGroupBox2("Lesson Information", nil)
	nameLayout := widgets.NewQFormLayout(nil)
	nameGroup.SetLayout(nameLayout)

	nameEdit := widgets.NewQLineEdit(nil)
	nameEdit.SetObjectName("lessonName")
	nameEdit.SetPlaceholderText("Enter lesson name...")
	nameLayout.AddRow3("Name:", nameEdit)

	descEdit := widgets.NewQTextEdit(nil)
	descEdit.SetObjectName("lessonDescription")
	descEdit.SetPlaceholderText("Enter lesson description...")
	descEdit.SetMaximumHeight(80)
	nameLayout.AddRow3("Description:", descEdit)

	layout.AddWidget(nameGroup, 0, 0)

	// Lesson type
	typeGroup := widgets.NewQGroupBox2("Lesson Type", nil)
	typeLayout := widgets.NewQVBoxLayout()
	typeGroup.SetLayout(typeLayout)

	wordsRadio := widgets.NewQRadioButton2("Words List", nil)
	wordsRadio.SetObjectName("wordsRadio")
	wordsRadio.SetChecked(true)
	typeLayout.AddWidget(wordsRadio, 0, 0)

	topoRadio := widgets.NewQRadioButton2("Topology", nil)
	topoRadio.SetObjectName("topoRadio")
	typeLayout.AddWidget(topoRadio, 0, 0)

	mediaRadio := widgets.NewQRadioButton2("Media", nil)
	mediaRadio.SetObjectName("mediaRadio")
	typeLayout.AddWidget(mediaRadio, 0, 0)

	layout.AddWidget(typeGroup, 0, 0)

	// Language settings
	langGroup := widgets.NewQGroupBox2("Languages", nil)
	langLayout := widgets.NewQFormLayout(nil)
	langGroup.SetLayout(langLayout)

	questionLangCombo := widgets.NewQComboBox(nil)
	questionLangCombo.SetObjectName("questionLanguage")
	questionLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	langLayout.AddRow3("Question language:", questionLangCombo)

	answerLangCombo := widgets.NewQComboBox(nil)
	answerLangCombo.SetObjectName("answerLanguage")
	answerLangCombo.AddItems([]string{"English", "Dutch", "French", "German", "Spanish", "Italian"})
	answerLangCombo.SetCurrentIndex(1) // Default to Dutch
	langLayout.AddRow3("Answer language:", answerLangCombo)

	layout.AddWidget(langGroup, 0, 0)

	// Buttons
	buttonBox := widgets.NewQDialogButtonBox3(
		widgets.QDialogButtonBox__Ok|widgets.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.newLessonDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.newLessonDialog.Reject()
	})
}

// createPropertiesDialog creates the lesson properties dialog
func (mod *LessonDialogsModule) createPropertiesDialog(parent *widgets.QWidget) {
	mod.propertiesDialog = widgets.NewQDialog(parent, 0)
	mod.propertiesDialog.SetWindowTitle("Lesson Properties")
	mod.propertiesDialog.SetFixedSize2(450, 400)
	mod.propertiesDialog.SetWindowModality(qtcore.Qt__ApplicationModal)

	layout := widgets.NewQVBoxLayout()
	mod.propertiesDialog.SetLayout(layout)

	// Create tab widget
	tabWidget := widgets.NewQTabWidget(nil)
	layout.AddWidget(tabWidget, 1, 0)

	// General tab
	generalTab := widgets.NewQWidget(nil, 0)
	generalLayout := widgets.NewQFormLayout(nil)
	generalTab.SetLayout(generalLayout)

	nameEdit := widgets.NewQLineEdit(nil)
	nameEdit.SetObjectName("propLessonName")
	generalLayout.AddRow3("Name:", nameEdit)

	descEdit := widgets.NewQTextEdit(nil)
	descEdit.SetObjectName("propLessonDescription")
	descEdit.SetMaximumHeight(100)
	generalLayout.AddRow3("Description:", descEdit)

	authorEdit := widgets.NewQLineEdit(nil)
	authorEdit.SetObjectName("propAuthor")
	generalLayout.AddRow3("Author:", authorEdit)

	versionEdit := widgets.NewQLineEdit(nil)
	versionEdit.SetObjectName("propVersion")
	generalLayout.AddRow3("Version:", versionEdit)

	tabWidget.AddTab(generalTab, "General")

	// Statistics tab
	statsTab := widgets.NewQWidget(nil, 0)
	statsLayout := widgets.NewQFormLayout(nil)
	statsTab.SetLayout(statsLayout)

	itemCountLabel := widgets.NewQLabel2("0", nil, 0)
	itemCountLabel.SetObjectName("itemCount")
	statsLayout.AddRow3("Number of items:", itemCountLabel)

	createdLabel := widgets.NewQLabel2("Unknown", nil, 0)
	createdLabel.SetObjectName("createdDate")
	statsLayout.AddRow3("Created:", createdLabel)

	modifiedLabel := widgets.NewQLabel2("Unknown", nil, 0)
	modifiedLabel.SetObjectName("modifiedDate")
	statsLayout.AddRow3("Last modified:", modifiedLabel)

	fileSizeLabel := widgets.NewQLabel2("0 KB", nil, 0)
	fileSizeLabel.SetObjectName("fileSize")
	statsLayout.AddRow3("File size:", fileSizeLabel)

	tabWidget.AddTab(statsTab, "Statistics")

	// Buttons
	buttonBox := widgets.NewQDialogButtonBox3(
		widgets.QDialogButtonBox__Ok|widgets.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.propertiesDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.propertiesDialog.Reject()
	})
}

// createImportDialog creates the lesson import dialog
func (mod *LessonDialogsModule) createImportDialog(parent *widgets.QWidget) {
	mod.importDialog = widgets.NewQDialog(parent, 0)
	mod.importDialog.SetWindowTitle("Import Lesson")
	mod.importDialog.SetFixedSize2(500, 300)
	mod.importDialog.SetWindowModality(qtcore.Qt__ApplicationModal)

	layout := widgets.NewQVBoxLayout()
	mod.importDialog.SetLayout(layout)

	// File selection
	fileGroup := widgets.NewQGroupBox2("Source File", nil)
	fileLayout := widgets.NewQHBoxLayout()
	fileGroup.SetLayout(fileLayout)

	fileEdit := widgets.NewQLineEdit(nil)
	fileEdit.SetObjectName("importFile")
	fileEdit.SetPlaceholderText("Select file to import...")
	fileLayout.AddWidget(fileEdit, 1, 0)

	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectClicked(func(checked bool) {
		fileName := widgets.QFileDialog_GetOpenFileName(mod.importDialog,
			"Select lesson file",
			"",
			"All supported files (*.ot *.txt *.csv);;OpenTeacher files (*.ot);;Text files (*.txt);;CSV files (*.csv);;All files (*.*)",
			"",
			0)
		if fileName != "" {
			fileEdit.SetText(fileName)
		}
	})
	fileLayout.AddWidget(browseButton, 0, 0)

	layout.AddWidget(fileGroup, 0, 0)

	// Import options
	optionsGroup := widgets.NewQGroupBox2("Import Options", nil)
	optionsLayout := widgets.NewQVBoxLayout()
	optionsGroup.SetLayout(optionsLayout)

	encodingLayout := widgets.NewQHBoxLayout()
	encodingLabel := widgets.NewQLabel2("File encoding:", nil, 0)
	encodingLayout.AddWidget(encodingLabel, 0, 0)

	encodingCombo := widgets.NewQComboBox(nil)
	encodingCombo.SetObjectName("encoding")
	encodingCombo.AddItems([]string{"UTF-8", "UTF-16", "ISO-8859-1", "ASCII"})
	encodingLayout.AddWidget(encodingCombo, 1, 0)
	optionsLayout.AddLayout(encodingLayout, 0)

	separatorLayout := widgets.NewQHBoxLayout()
	separatorLabel := widgets.NewQLabel2("Field separator:", nil, 0)
	separatorLayout.AddWidget(separatorLabel, 0, 0)

	separatorCombo := widgets.NewQComboBox(nil)
	separatorCombo.SetObjectName("separator")
	separatorCombo.AddItems([]string{"Tab", "Comma", "Semicolon", "Space"})
	separatorLayout.AddWidget(separatorCombo, 1, 0)
	optionsLayout.AddLayout(separatorLayout, 0)

	firstRowCheck := widgets.NewQCheckBox2("First row contains headers", nil)
	firstRowCheck.SetObjectName("firstRowHeaders")
	firstRowCheck.SetChecked(true)
	optionsLayout.AddWidget(firstRowCheck, 0, 0)

	layout.AddWidget(optionsGroup, 0, 0)

	// Preview area
	previewGroup := widgets.NewQGroupBox2("Preview", nil)
	previewLayout := widgets.NewQVBoxLayout()
	previewGroup.SetLayout(previewLayout)

	previewText := widgets.NewQTextEdit(nil)
	previewText.SetObjectName("preview")
	previewText.SetReadOnly(true)
	previewText.SetPlainText("Select a file to see preview...")
	previewLayout.AddWidget(previewText, 1, 0)

	layout.AddWidget(previewGroup, 1, 0)

	// Buttons
	buttonBox := widgets.NewQDialogButtonBox3(
		widgets.QDialogButtonBox__Ok|widgets.QDialogButtonBox__Cancel,
		nil)
	layout.AddWidget(buttonBox, 0, 0)

	buttonBox.ConnectAccepted(func() {
		mod.importDialog.Accept()
	})

	buttonBox.ConnectRejected(func() {
		mod.importDialog.Reject()
	})
}

// resetNewLessonForm resets the new lesson dialog form
func (mod *LessonDialogsModule) resetNewLessonForm() {
	if mod.newLessonDialog == nil {
		return
	}

	nameEdit := widgets.NewQLineEditFromPointer(mod.newLessonDialog.FindChild("lessonName", qtcore.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		nameEdit.Clear()
	}

	descEdit := widgets.NewQTextEditFromPointer(mod.newLessonDialog.FindChild("lessonDescription", qtcore.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		descEdit.Clear()
	}

	wordsRadio := widgets.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("wordsRadio", qtcore.Qt__FindChildrenRecursively).Pointer())
	if wordsRadio != nil {
		wordsRadio.SetChecked(true)
	}
}

// getNewLessonData extracts data from the new lesson dialog
func (mod *LessonDialogsModule) getNewLessonData() map[string]interface{} {
	if mod.newLessonDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	nameEdit := widgets.NewQLineEditFromPointer(mod.newLessonDialog.FindChild("lessonName", qtcore.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		data["name"] = strings.TrimSpace(nameEdit.Text())
	}

	descEdit := widgets.NewQTextEditFromPointer(mod.newLessonDialog.FindChild("lessonDescription", qtcore.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		data["description"] = strings.TrimSpace(descEdit.ToPlainText())
	}

	// Determine lesson type
	wordsRadio := widgets.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("wordsRadio", qtcore.Qt__FindChildrenRecursively).Pointer())
	topoRadio := widgets.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("topoRadio", qtcore.Qt__FindChildrenRecursively).Pointer())
	mediaRadio := widgets.NewQRadioButtonFromPointer(mod.newLessonDialog.FindChild("mediaRadio", qtcore.Qt__FindChildrenRecursively).Pointer())

	if wordsRadio != nil && wordsRadio.IsChecked() {
		data["type"] = "words"
	} else if topoRadio != nil && topoRadio.IsChecked() {
		data["type"] = "topology"
	} else if mediaRadio != nil && mediaRadio.IsChecked() {
		data["type"] = "media"
	} else {
		data["type"] = "words" // default
	}

	// Get languages
	questionLang := widgets.NewQComboBoxFromPointer(mod.newLessonDialog.FindChild("questionLanguage", qtcore.Qt__FindChildrenRecursively).Pointer())
	if questionLang != nil {
		data["questionLanguage"] = questionLang.CurrentText()
	}

	answerLang := widgets.NewQComboBoxFromPointer(mod.newLessonDialog.FindChild("answerLanguage", qtcore.Qt__FindChildrenRecursively).Pointer())
	if answerLang != nil {
		data["answerLanguage"] = answerLang.CurrentText()
	}

	return data
}

// loadPropertiesData loads lesson data into the properties dialog
func (mod *LessonDialogsModule) loadPropertiesData(lessonData map[string]interface{}) {
	if mod.propertiesDialog == nil || lessonData == nil {
		return
	}

	if name, ok := lessonData["name"].(string); ok {
		nameEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propLessonName", qtcore.Qt__FindChildrenRecursively).Pointer())
		if nameEdit != nil {
			nameEdit.SetText(name)
		}
	}

	if desc, ok := lessonData["description"].(string); ok {
		descEdit := widgets.NewQTextEditFromPointer(mod.propertiesDialog.FindChild("propLessonDescription", qtcore.Qt__FindChildrenRecursively).Pointer())
		if descEdit != nil {
			descEdit.SetPlainText(desc)
		}
	}

	if author, ok := lessonData["author"].(string); ok {
		authorEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propAuthor", qtcore.Qt__FindChildrenRecursively).Pointer())
		if authorEdit != nil {
			authorEdit.SetText(author)
		}
	}

	if version, ok := lessonData["version"].(string); ok {
		versionEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propVersion", qtcore.Qt__FindChildrenRecursively).Pointer())
		if versionEdit != nil {
			versionEdit.SetText(version)
		}
	}

	// Update statistics
	if itemCount, ok := lessonData["itemCount"].(int); ok {
		itemLabel := widgets.NewQLabelFromPointer(mod.propertiesDialog.FindChild("itemCount", qtcore.Qt__FindChildrenRecursively).Pointer())
		if itemLabel != nil {
			itemLabel.SetText(fmt.Sprintf("%d", itemCount))
		}
	}
}

// getPropertiesData extracts data from the properties dialog
func (mod *LessonDialogsModule) getPropertiesData() map[string]interface{} {
	if mod.propertiesDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	nameEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propLessonName", qtcore.Qt__FindChildrenRecursively).Pointer())
	if nameEdit != nil {
		data["name"] = strings.TrimSpace(nameEdit.Text())
	}

	descEdit := widgets.NewQTextEditFromPointer(mod.propertiesDialog.FindChild("propLessonDescription", qtcore.Qt__FindChildrenRecursively).Pointer())
	if descEdit != nil {
		data["description"] = strings.TrimSpace(descEdit.ToPlainText())
	}

	authorEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propAuthor", qtcore.Qt__FindChildrenRecursively).Pointer())
	if authorEdit != nil {
		data["author"] = strings.TrimSpace(authorEdit.Text())
	}

	versionEdit := widgets.NewQLineEditFromPointer(mod.propertiesDialog.FindChild("propVersion", qtcore.Qt__FindChildrenRecursively).Pointer())
	if versionEdit != nil {
		data["version"] = strings.TrimSpace(versionEdit.Text())
	}

	return data
}

// getImportData extracts data from the import dialog
func (mod *LessonDialogsModule) getImportData() map[string]interface{} {
	if mod.importDialog == nil {
		return nil
	}

	data := make(map[string]interface{})

	fileEdit := widgets.NewQLineEditFromPointer(mod.importDialog.FindChild("importFile", qtcore.Qt__FindChildrenRecursively).Pointer())
	if fileEdit != nil {
		data["file"] = strings.TrimSpace(fileEdit.Text())
	}

	encodingCombo := widgets.NewQComboBoxFromPointer(mod.importDialog.FindChild("encoding", qtcore.Qt__FindChildrenRecursively).Pointer())
	if encodingCombo != nil {
		data["encoding"] = encodingCombo.CurrentText()
	}

	separatorCombo := widgets.NewQComboBoxFromPointer(mod.importDialog.FindChild("separator", qtcore.Qt__FindChildrenRecursively).Pointer())
	if separatorCombo != nil {
		data["separator"] = separatorCombo.CurrentText()
	}

	firstRowCheck := widgets.NewQCheckBoxFromPointer(mod.importDialog.FindChild("firstRowHeaders", qtcore.Qt__FindChildrenRecursively).Pointer())
	if firstRowCheck != nil {
		data["firstRowHeaders"] = firstRowCheck.IsChecked()
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
