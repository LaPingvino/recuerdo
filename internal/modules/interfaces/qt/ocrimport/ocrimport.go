// Package ocrimport is the dialog for reading a word list from a picture
// (File > Import from Picture), after OpenTeacher's ocrGui wizard: the
// picture can be straightened and the list cropped out before it is read
// (internal/ocr, with Tesseract).
package ocrimport

import (
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"os"
	"path/filepath"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/ocr"
	qt "github.com/mappu/miqt/qt6"
)

// previewSize is the largest side of the preview.
const previewSize = 560

// Rect is a rectangle in picture coordinates.
type Rect struct{ X, Y, W, H int }

// toImage converts a point on the preview, shown at scale (preview pixels
// per picture pixel), to picture coordinates within w×h.
func toImage(px, py int, scale float64, w, h int) (int, int) {
	if scale <= 0 {
		return 0, 0
	}
	x := min(max(int(float64(px)/scale), 0), w)
	y := min(max(int(float64(py)/scale), 0), h)
	return x, y
}

// spanning is the rectangle between two corners.
func spanning(x1, y1, x2, y2 int) Rect {
	return Rect{min(x1, x2), min(y1, y2), max(x1, x2) - min(x1, x2), max(y1, y2) - min(y1, y2)}
}

// Dialog shows a picture to straighten and crop before reading it.
type Dialog struct {
	*qt.QDialog
	original *qt.QImage
	rotation float64 // degrees
	crop     Rect    // in the rotated picture; empty: all of it
	scale    float64
	preview  *qt.QLabel
	slider   *qt.QSlider
	dragFrom [2]int
	dragging bool
}

// New creates the dialog for the picture at path.
func New(parent *qt.QWidget, path string) (*Dialog, error) {
	img := qt.NewQImage8(path)
	if img.IsNull() {
		return nil, os.ErrInvalid
	}
	d := &Dialog{QDialog: qt.NewQDialog(parent), original: img}
	d.SetWindowTitle(i18n.T("Import from Picture"))
	layout := qt.NewQVBoxLayout(d.QWidget)
	hint := qt.NewQLabel(d.QWidget)
	hint.SetText(i18n.T("Straighten the picture so the lines are level, and drag a rectangle around the word list if the picture shows more."))
	hint.SetWordWrap(true)
	layout.AddWidget(hint.QWidget)

	d.preview = qt.NewQLabel(d.QWidget)
	d.preview.SetAlignment(qt.AlignLeft | qt.AlignTop)
	d.preview.SetMinimumSize2(200, 150)
	d.preview.SetCursor(qt.NewQCursor2(qt.CrossCursor))
	d.preview.OnMousePressEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		d.dragging, d.dragFrom = true, [2]int{e.X(), e.Y()}
	})
	d.preview.OnMouseMoveEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		if d.dragging {
			d.setCropFromPreview(d.dragFrom[0], d.dragFrom[1], e.X(), e.Y())
		}
	})
	d.preview.OnMouseReleaseEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		if d.dragging {
			d.dragging = false
			d.setCropFromPreview(d.dragFrom[0], d.dragFrom[1], e.X(), e.Y())
		}
	})
	layout.AddWidget(d.preview.QWidget)

	rotateRow := qt.NewQHBoxLayout2()
	label := qt.NewQLabel(d.QWidget)
	label.SetText(i18n.T("Straighten:"))
	rotateRow.AddWidget(label.QWidget)
	d.slider = qt.NewQSlider3(qt.Horizontal)
	d.slider.SetRange(-45, 45)
	d.slider.OnValueChanged(func(v int) { d.SetRotation(float64(v)) })
	rotateRow.AddWidget(d.slider.QWidget)
	turn := qt.NewQPushButton3(i18n.T("Rotate 90°"))
	turn.OnClicked(func() { d.SetRotation(d.rotation + 90) })
	rotateRow.AddWidget(turn.QWidget)
	clear := qt.NewQPushButton3(i18n.T("Whole picture"))
	clear.OnClicked(func() { d.crop = Rect{}; d.updatePreview() })
	rotateRow.AddWidget(clear.QWidget)
	layout.AddLayout(rotateRow.QLayout)

	buttons := qt.NewQDialogButtonBox(d.QWidget)
	buttons.SetStandardButtons(qt.QDialogButtonBox__Ok | qt.QDialogButtonBox__Cancel)
	buttons.Button(qt.QDialogButtonBox__Ok).SetText(i18n.T("Import"))
	buttons.OnAccepted(func() { d.Accept() })
	buttons.OnRejected(func() { d.Reject() })
	layout.AddWidget(buttons.QWidget)

	d.updatePreview()
	return d, nil
}

// SetRotation turns the picture by degrees (clockwise); the crop is reset.
func (d *Dialog) SetRotation(degrees float64) {
	d.rotation, d.crop = degrees, Rect{}
	d.updatePreview()
}

// SetCrop limits the reading to r, in the rotated picture.
func (d *Dialog) SetCrop(r Rect) {
	d.crop = r
	d.updatePreview()
}

func (d *Dialog) setCropFromPreview(x1, y1, x2, y2 int) {
	img := d.rotated()
	ax, ay := toImage(x1, y1, d.scale, img.Width(), img.Height())
	bx, by := toImage(x2, y2, d.scale, img.Width(), img.Height())
	d.SetCrop(spanning(ax, ay, bx, by))
}

func (d *Dialog) rotated() *qt.QImage {
	if d.rotation == 0 {
		return d.original
	}
	m := qt.NewQTransform2()
	m.Rotate(d.rotation)
	return d.original.Transformed2(m, qt.SmoothTransformation)
}

// Prepared is the picture as it will be read: rotated, then cropped.
func (d *Dialog) Prepared() *qt.QImage {
	img := d.rotated()
	if d.crop.W > 4 && d.crop.H > 4 {
		return img.Copy2(d.crop.X, d.crop.Y, d.crop.W, d.crop.H)
	}
	return img
}

func (d *Dialog) updatePreview() {
	img := d.rotated()
	d.scale = min(1, float64(previewSize)/float64(max(img.Width(), img.Height(), 1)))
	scaled := img.Scaled3(int(float64(img.Width())*d.scale), int(float64(img.Height())*d.scale), qt.KeepAspectRatio, qt.SmoothTransformation)
	pm := qt.QPixmap_FromImage(scaled)
	if d.crop.W > 0 && d.crop.H > 0 {
		p := qt.NewQPainter2(pm.QPaintDevice)
		p.SetPen(qt.NewQColor3(220, 40, 40))
		p.DrawRect(qt.NewQRectF4(float64(d.crop.X)*d.scale, float64(d.crop.Y)*d.scale, float64(d.crop.W)*d.scale, float64(d.crop.H)*d.scale))
		p.End()
	}
	d.preview.SetPixmap(pm)
	d.preview.SetMinimumSize2(pm.Width(), pm.Height())
}

// Words reads the word list in the prepared picture.
func (d *Dialog) Words() ([]lesson.WordItem, error) {
	dir, err := os.MkdirTemp("", "recuerdo-ocr-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "page.png")
	if !d.Prepared().Save(path) {
		return nil, os.ErrInvalid
	}
	return ocr.LoadWordList(path)
}
