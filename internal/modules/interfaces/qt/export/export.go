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
	w := qt.NewQPdfWriter(path)
	w.QPagedPaintDevice.SetPageSize(qt.NewQPageSize2(qt.QPageSize__A4))
	w.QPagedPaintDevice.SetPageMargins2(qt.NewQMarginsF2(25, 25, 25, 25), qt.QPageLayout__Millimeter)
	doc.Print(w.QPagedPaintDevice)
	w.Delete()
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
