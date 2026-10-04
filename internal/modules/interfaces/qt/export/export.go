// Package export saves lessons in the formats that need Qt or another
// program: PDF and ODT (made with Qt from the HTML export, as OpenTeacher
// does) and, through LibreOffice when it is installed, word processor and
// spreadsheet formats. Other formats go to internal/lesson's FileSaver.
package export

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/mappu/miqt/qt"
	"github.com/mappu/miqt/qt/printsupport"
)

// libreOfficeText and libreOfficeSheets map the formats LibreOffice
// converts to its filter names, from an ODT and a SYLK file respectively
// (OpenTeacher's libreofficeFormats saver).
var (
	libreOfficeText = map[string]string{
		".doc": "MS Word 97", ".rtf": "Rich Text Format",
		".docx": "Office Open XML Text", ".uot": "UOF text",
	}
	libreOfficeSheets = map[string]string{
		".ods": "calc8", ".xls": "MS Excel 97", ".xlsx": "Calc Office Open XML",
		".dif": "DIF", ".uos": "UOF spreadsheet",
	}
)

// ErrNoLibreOffice is returned when a format needs LibreOffice and it is
// not installed.
var ErrNoLibreOffice = errors.New("this format needs LibreOffice (soffice), which was not found")

// Extensions are the formats Save handles itself, beyond FileSaver's.
func Extensions() []string {
	exts := []string{".pdf", ".odt"}
	for e := range libreOfficeText {
		exts = append(exts, e)
	}
	for e := range libreOfficeSheets {
		exts = append(exts, e)
	}
	return exts
}

// Save writes a lesson to path in the format its extension names. PDF and
// ODT need a QApplication.
func Save(data *lesson.LessonData, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch {
	case IsTopo(data) && ext == ".png":
		return saveMapPNG(data, path)
	case IsTopo(data) && ext == ".pdf":
		return saveMapPDF(data, path)
	case ext == ".pdf":
		return savePDF(data, path)
	case ext == ".odt":
		return saveODT(data, path)
	case libreOfficeText[ext] != "":
		return viaLibreOffice(data, path, ".odt", libreOfficeText[ext])
	case libreOfficeSheets[ext] != "":
		return viaLibreOffice(data, path, ".slk", libreOfficeSheets[ext])
	}
	return lesson.NewFileSaver().SaveFile(data, path)
}

// document is the lesson's HTML export as a Qt text document.
func document(data *lesson.LessonData) (*qt.QTextDocument, error) {
	if IsMedia(data) {
		return mediaDocument(data), nil
	}
	tmp, err := os.MkdirTemp("", "recuerdo-export-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	htmlPath := filepath.Join(tmp, "lesson.html")
	if err := lesson.NewFileSaver().SaveFile(data, htmlPath); err != nil {
		return nil, err
	}
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, err
	}
	doc := qt.NewQTextDocument()
	doc.SetHtml(string(html))
	return doc, nil
}

// savePDF prints the HTML export to an A4 PDF with 25 mm margins.
func savePDF(data *lesson.LessonData, path string) error {
	doc, err := document(data)
	if err != nil {
		return err
	}
	// through a PDF printer with its own margins, as printing does: with
	// margins set (on a QPdfWriter or here) QTextDocument.Print lays the
	// document out at about half the page's width
	printer := printsupport.NewQPrinter()
	defer printer.Delete()
	printer.SetOutputFormat(printsupport.QPrinter__PdfFormat)
	printer.SetOutputFileName(path)
	printer.QPagedPaintDevice.SetPageSize(qt.NewQPageSize2(qt.QPageSize__A4))
	printer.SetDocName(data.List.Title)
	printer.SetCreator("Recuerdo")
	doc.Print(printer.QPagedPaintDevice)
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		return fmt.Errorf("could not write PDF %s", path)
	}
	return nil
}

// saveODT writes the HTML export as an OpenDocument text file.
func saveODT(data *lesson.LessonData, path string) error {
	doc, err := document(data)
	if err != nil {
		return err
	}
	if !qt.NewQTextDocumentWriter4(path, []byte("odf")).Write(doc) {
		return fmt.Errorf("could not write ODT %s", path)
	}
	return nil
}

// viaLibreOffice saves an interim file (ODT for text, SYLK for
// spreadsheets) and has LibreOffice convert it.
func viaLibreOffice(data *lesson.LessonData, path, interimExt, filter string) error {
	soffice, err := exec.LookPath("soffice")
	if err != nil {
		if soffice, err = exec.LookPath("libreoffice"); err != nil {
			return ErrNoLibreOffice
		}
	}
	tmp, err := os.MkdirTemp("", "recuerdo-lo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	interim := filepath.Join(tmp, "document"+interimExt)
	if err := Save(data, interim); err != nil {
		return err
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	cmd := exec.Command(soffice,
		"-env:UserInstallation=file://"+filepath.ToSlash(filepath.Join(tmp, "profile")),
		"--headless", "--convert-to", ext+":"+filter, "--outdir", tmp, interim)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("LibreOffice could not convert to %s: %v: %s", ext, err, out)
	}
	converted := filepath.Join(tmp, "document."+ext)
	b, err := os.ReadFile(converted)
	if err != nil {
		return fmt.Errorf("LibreOffice did not write %s", ext)
	}
	return os.WriteFile(path, b, 0o644)
}

// formatNames name the formats Save handles itself.
var formatNames = map[string]string{
	".pdf": "PDF", ".odt": "OpenDocument Text",
	".doc": "Word 97", ".docx": "Word", ".rtf": "Rich Text", ".uot": "Uniform Office text",
	".ods": "OpenDocument Spreadsheet", ".xls": "Excel 97", ".xlsx": "Excel",
	".dif": "Data Interchange Format", ".uos": "Uniform Office spreadsheet",
}

// SaveFilterFor is the file dialog filter for saving data: for a
// topography lesson its own format, the map as a picture and as PDF; for a
// media lesson its own format and PDF; for others SaveFilter.
func SaveFilterFor(data *lesson.LessonData) string {
	switch {
	case IsTopo(data):
		return "OpenTeaching Topography (*.ottp);;Map picture (*.png);;Map as PDF (*.pdf)"
	case IsMedia(data):
		return "OpenTeaching Media (*.otmd);;PDF (*.pdf)"
	}
	return SaveFilter()
}

// DefaultExtension is the extension a lesson is saved with by default.
func DefaultExtension(data *lesson.LessonData) string {
	switch {
	case IsTopo(data):
		return ".ottp"
	case IsMedia(data):
		return ".otmd"
	}
	return ".otwd"
}

// SaveFilter is a file dialog filter with every format Save can write,
// the lesson formats first (OpenTeaching Words, the default, at the top).
func SaveFilter() string {
	fs := lesson.NewFileSaver()
	exts := append([]string{".otwd"}, fs.GetSupportedSaveExtensions()...)
	var parts []string
	seen := map[string]bool{}
	add := func(name, ext string) {
		if !seen[ext] {
			seen[ext] = true
			parts = append(parts, fmt.Sprintf("%s (*%s)", name, ext))
		}
	}
	for _, e := range exts {
		add(fs.GetSaveFormatName(e), e)
	}
	for _, e := range []string{".pdf", ".odt", ".docx", ".doc", ".rtf", ".uot", ".xlsx", ".xls", ".ods", ".dif", ".uos"} {
		add(formatNames[e], e)
	}
	return strings.Join(parts, ";;")
}

// CanSave reports whether Save can write a file with path's extension.
func CanSave(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if formatNames[ext] != "" {
		return true
	}
	for _, e := range lesson.NewFileSaver().GetSupportedSaveExtensions() {
		if e == ext {
			return true
		}
	}
	return false
}

// WithExtension adds the extension the chosen filter names when path has
// none of its own (file dialogs on some systems do not).
func WithExtension(path, filter string) string {
	if filepath.Ext(path) != "" {
		return path
	}
	if i := strings.Index(filter, "(*."); i >= 0 {
		if j := strings.Index(filter[i:], ")"); j > 0 {
			return path + filter[i+2:i+j]
		}
	}
	return path + ".otwd"
}

// Print prints a lesson's HTML export on printer, as OpenTeacher's word
// list printing does: the document is named after the lesson.
func Print(data *lesson.LessonData, printer *printsupport.QPrinter) error {
	name := data.List.Title
	if name == "" {
		name = "Untitled word list"
	}
	printer.SetDocName(name)
	printer.SetCreator("Recuerdo")
	if IsTopo(data) {
		// a topography lesson prints its map, as OpenTeacher's print/topo
		return printMap(data, printer)
	}
	doc, err := document(data)
	if err != nil {
		return err
	}
	doc.Print(printer.QPagedPaintDevice)
	return nil
}
