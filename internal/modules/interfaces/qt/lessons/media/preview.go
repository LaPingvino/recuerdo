package media

import (
	"fmt"
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"os"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
)

// Preview shows a media item: pictures and text in place, other media
// with a button that opens them in the system's player or web browser.
type Preview struct {
	*qt.QWidget
	picture *qt.QLabel
	text    *qt.QTextBrowser
	other   *qt.QLabel
	open    *qt.QPushButton
	image   *qt.QImage
	target  string // what Open opens: a web address or a file
	files   map[string][]byte
	// Opened, if set, replaces opening (for tests).
	Opened func(target string)
}

// NewPreview creates an empty preview.
func NewPreview(parent *qt.QWidget) *Preview {
	p := &Preview{QWidget: qt.NewQWidget(parent)}
	layout := qt.NewQVBoxLayout(p.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	p.picture = qt.NewQLabel(p.QWidget)
	p.picture.SetAlignment(qt.AlignCenter)
	p.picture.SetMinimumSize2(160, 120)
	p.picture.SetSizePolicy2(qt.QSizePolicy__Ignored, qt.QSizePolicy__Ignored)
	layout.AddWidget2(p.picture.QWidget, 1)
	p.text = qt.NewQTextBrowser(p.QWidget)
	layout.AddWidget2(p.text.QWidget, 1)
	p.other = qt.NewQLabel(p.QWidget)
	p.other.SetAlignment(qt.AlignCenter)
	p.other.SetWordWrap(true)
	layout.AddWidget2(p.other.QWidget, 1)
	p.open = qt.NewQPushButton3(i18n.T("Open"))
	p.open.OnClicked(p.Open)
	row := qt.NewQHBoxLayout2()
	row.AddStretch()
	row.AddWidget(p.open.QWidget)
	row.AddStretch()
	layout.AddLayout(row.QLayout)
	p.OnResizeEvent(func(super func(*qt.QResizeEvent), e *qt.QResizeEvent) {
		super(e)
		p.scalePicture()
	})
	p.Show(nil, nil)
	return p
}

// Show shows item, whose embedded files are in files (nil: nothing).
func (p *Preview) Show(item *lesson.WordItem, files map[string][]byte) {
	p.picture.SetVisible(false)
	p.text.SetVisible(false)
	p.other.SetVisible(true)
	p.open.SetVisible(false)
	p.image, p.target = nil, ""
	if item == nil || item.Filename == nil || *item.Filename == "" {
		p.other.SetText("")
		return
	}
	name, remote := *item.Filename, item.Remote != nil && *item.Remote
	kind := Kind(name, remote)
	data := files[name]
	if remote {
		p.target = name
	} else if data != nil {
		p.target = "file:" + name // written out when opened
	}
	switch {
	case kind == Image && data != nil:
		if img := qt.QImage_FromDataWithData(data); !img.IsNull() {
			p.image = img
			p.picture.SetVisible(true)
			p.other.SetVisible(false)
			p.scalePicture()
			return
		}
	case kind == Text && data != nil:
		p.text.SetPlainText(string(data))
		p.text.SetVisible(true)
		p.other.SetVisible(false)
		return
	}
	switch {
	case p.target == "":
		p.other.SetText(fmt.Sprintf(i18n.T("The %s %s is missing from the lesson."), kind, filepath.Base(name)))
	case remote:
		p.other.SetText(fmt.Sprintf("A %s:\n%s", kind, name))
	default:
		p.other.SetText(fmt.Sprintf("A %s: %s", kind, filepath.Base(name)))
	}
	p.open.SetText(map[string]string{Audio: "▶ Play", Video: "▶ Play", YouTube: "▶ Play", Vimeo: "▶ Play",
		Dailymotion: "▶ Play"}[kind])
	if p.open.Text() == "" {
		p.open.SetText(i18n.T("Open"))
	}
	p.open.SetVisible(p.target != "")
	p.files = files
}

func (p *Preview) scalePicture() {
	if p.image == nil {
		return
	}
	w, h := max(p.picture.Width(), 160), max(p.picture.Height(), 120)
	scaled := p.image.Scaled3(min(w, p.image.Width()*2), min(h, p.image.Height()*2), qt.KeepAspectRatio, qt.SmoothTransformation)
	p.picture.SetPixmap(qt.QPixmap_FromImage(scaled))
}

// Open opens the shown medium outside Recuerdo: a web address in the
// browser, an embedded file (written to a temporary file) in the
// system's program for it.
func (p *Preview) Open() {
	target := p.target
	if target == "" {
		return
	}
	if name, ok := cutPrefix(target, "file:"); ok {
		dir := filepath.Join(os.TempDir(), "recuerdo-media")
		path := filepath.Join(dir, filepath.Base(name))
		if os.MkdirAll(dir, 0o755) != nil || os.WriteFile(path, p.files[name], 0o644) != nil {
			return
		}
		target = path
	}
	if p.Opened != nil {
		p.Opened(target)
		return
	}
	if AddressOK(target) {
		qt.QDesktopServices_OpenUrl(qt.NewQUrl3(target))
	} else {
		qt.QDesktopServices_OpenUrl(qt.QUrl_FromLocalFile(target))
	}
}

func cutPrefix(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):], true
	}
	return s, false
}
