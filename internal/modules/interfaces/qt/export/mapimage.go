package export

import (
	"fmt"
	"os"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/printsupport"
)

// IsTopo reports whether data is a topography lesson: one with a map.
func IsTopo(data *lesson.LessonData) bool {
	img, ok := data.Resources[lesson.MapImageResource].([]byte)
	return ok && len(img) > 0
}

// MapPicture is a topography lesson's map at full size with its places as
// dots and their names, as OpenTeacher saves and prints it.
func MapPicture(data *lesson.LessonData) (*qt.QImage, error) {
	b, _ := data.Resources[lesson.MapImageResource].([]byte)
	src := qt.QImage_FromDataWithData(b)
	if src.IsNull() {
		return nil, fmt.Errorf("the lesson has no map")
	}
	img := src.ConvertToFormat(qt.QImage__Format_ARGB32)
	p := qt.NewQPainter2(img.QPaintDevice)
	p.SetRenderHint(qt.QPainter__Antialiasing)
	font := p.Font()
	font.SetPointSizeF(max(9, float64(img.Width())/90))
	p.SetFont(font)
	r := max(5, float64(img.Width())/160)
	outline := qt.NewQPen4(qt.NewQBrush3(qt.NewQColor3(255, 255, 255)), r/3)
	dot := qt.NewQBrush3(qt.NewQColor3(33, 102, 172))
	for _, it := range data.List.Items {
		x, y, ok := it.GetTopoCoordinates()
		if !ok {
			continue
		}
		fx, fy := float64(x), float64(y)
		p.SetPenWithPen(outline)
		p.SetBrush(dot)
		p.DrawEllipse3(qt.NewQPointF3(fx, fy), r, r)
		name := it.Name
		if name == "" && len(it.Questions) > 0 {
			name = it.Questions[0]
		}
		p.SetPen(qt.NewQColor3(255, 255, 255))
		for _, d := range [][2]float64{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			p.DrawText(qt.NewQPointF3(fx+r*1.6+d[0], fy+r/2+d[1]), name)
		}
		p.SetPen(qt.NewQColor3(20, 20, 20))
		p.DrawText(qt.NewQPointF3(fx+r*1.6, fy+r/2), name)
	}
	p.End()
	return img, nil
}

// paintMap draws the map picture as large as fits on a page device, centred.
func paintMap(data *lesson.LessonData, device *qt.QPaintDevice) error {
	img, err := MapPicture(data)
	if err != nil {
		return err
	}
	w, h := float64(device.Width()), float64(device.Height())
	s := min(w/float64(img.Width()), h/float64(img.Height()))
	iw, ih := float64(img.Width())*s, float64(img.Height())*s
	p := qt.NewQPainter2(device)
	p.SetRenderHint(qt.QPainter__SmoothPixmapTransform)
	p.DrawImage5(qt.NewQRectF4((w-iw)/2, (h-ih)/2, iw, ih), img)
	p.End()
	return nil
}

func saveMapPNG(data *lesson.LessonData, path string) error {
	img, err := MapPicture(data)
	if err != nil {
		return err
	}
	if !img.Save(path) {
		return fmt.Errorf("could not write %s", path)
	}
	return nil
}

func saveMapPDF(data *lesson.LessonData, path string) error {
	img, err := MapPicture(data)
	if err != nil {
		return err
	}
	w := qt.NewQPdfWriter(path)
	w.QPagedPaintDevice.SetPageSize(qt.NewQPageSize2(qt.QPageSize__A4))
	if img.Width() > img.Height() {
		w.QPagedPaintDevice.SetPageOrientation(qt.QPageLayout__Landscape)
	}
	w.QPagedPaintDevice.SetPageMargins(qt.NewQMarginsF2(15, 15, 15, 15), qt.QPageLayout__Millimeter)
	w.SetTitle(data.List.Title)
	w.SetCreator("Recuerdo")
	err = paintMap(data, w.QPagedPaintDevice.QPaintDevice)
	w.Delete()
	if err != nil {
		return err
	}
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		return fmt.Errorf("could not write PDF %s", path)
	}
	return nil
}

func printMap(data *lesson.LessonData, printer *printsupport.QPrinter) error {
	if b, _ := data.Resources[lesson.MapImageResource].([]byte); len(b) > 0 {
		if img := qt.QImage_FromDataWithData(b); !img.IsNull() && img.Width() > img.Height() {
			printer.QPagedPaintDevice.SetPageOrientation(qt.QPageLayout__Landscape)
		}
	}
	return paintMap(data, printer.QPagedPaintDevice.QPaintDevice)
}
