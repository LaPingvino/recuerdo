package topo

import (
	qt "github.com/mappu/miqt/qt6"
)

// Dot styles of places on a map.
const (
	dotRadius = 6.0
	clickNear = 18.0 // view pixels: how near a click must be to a place
)

// MapView shows a map with its places as dots, scaled to fit, and reports
// clicks in picture coordinates.
type MapView struct {
	*qt.QWidget
	image     *qt.QImage
	places    []Place
	showNames bool
	selected  int // place ID drawn as selected, -1 none
	hidden    bool
	onClick   func(x, y int)

	// editing: dragging a place moves it, right-clicking one or pressing
	// Delete acts on it (only where the callbacks are set)
	onMove    func(id, x, y int)
	onContext func(id int, at *qt.QPoint)
	onDelete  func()
	dragging  int // the ID of the place being dragged, -1 none
	dragged   bool
}

// NewMapView creates an empty map view.
func NewMapView(parent *qt.QWidget) *MapView {
	v := &MapView{QWidget: qt.NewQWidget(parent), selected: -1, showNames: true, dragging: -1}
	v.SetMinimumSize2(320, 240)
	v.SetSizePolicy2(qt.QSizePolicy__Expanding, qt.QSizePolicy__Expanding)
	v.OnPaintEvent(func(super func(*qt.QPaintEvent), e *qt.QPaintEvent) { v.paint() })
	v.OnMousePressEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		if v.image == nil {
			return
		}
		x, y := v.fit().ToPicture(float64(e.X()), float64(e.Y()))
		place, onPlace := v.Near(x, y)
		if e.Button() == qt.RightButton {
			if onPlace && v.onContext != nil {
				v.onContext(place.ID, e.GlobalPos())
			}
			return
		}
		if v.onClick != nil && x >= 0 && y >= 0 && x <= v.image.Width() && y <= v.image.Height() {
			v.onClick(x, y)
		}
		if onPlace && v.onMove != nil {
			v.dragging, v.dragged = place.ID, false
			v.SetCursor(qt.NewQCursor2(qt.ClosedHandCursor))
		}
	})
	v.OnMouseMoveEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		if v.image == nil {
			return
		}
		x, y := v.fit().ToPicture(float64(e.X()), float64(e.Y()))
		if v.dragging < 0 {
			// a hand over places that can be dragged
			if _, on := v.Near(x, y); on && v.onMove != nil {
				v.SetCursor(qt.NewQCursor2(qt.OpenHandCursor))
			} else {
				v.UnsetCursor()
			}
			return
		}
		x = min(max(x, 0), v.image.Width())
		y = min(max(y, 0), v.image.Height())
		for i := range v.places {
			if v.places[i].ID == v.dragging {
				v.places[i].X, v.places[i].Y = x, y
			}
		}
		v.dragged = true
		v.Update()
	})
	v.OnMouseReleaseEvent(func(super func(*qt.QMouseEvent), e *qt.QMouseEvent) {
		if v.dragging < 0 {
			return
		}
		id, moved := v.dragging, v.dragged
		v.dragging, v.dragged = -1, false
		v.UnsetCursor()
		if !moved {
			return
		}
		for _, p := range v.places {
			if p.ID == id {
				v.onMove(id, p.X, p.Y)
			}
		}
	})
	v.SetFocusPolicy(qt.StrongFocus)
	v.OnKeyPressEvent(func(super func(*qt.QKeyEvent), e *qt.QKeyEvent) {
		if k := qt.Key(e.Key()); (k == qt.Key_Delete || k == qt.Key_Backspace) && v.onDelete != nil {
			v.onDelete()
			return
		}
		super(e)
	})
	return v
}

// SetImage shows a picture (nil: none).
func (v *MapView) SetImage(img *qt.QImage) {
	if img != nil && img.IsNull() {
		img = nil
	}
	v.image = img
	v.Update()
}

// HasImage tells whether a map is shown.
func (v *MapView) HasImage() bool { return v.image != nil }

// SetPlaces sets the places drawn; names are written beside them when
// showNames is set.
func (v *MapView) SetPlaces(places []Place, showNames bool) {
	v.places, v.showNames = places, showNames
	v.Update()
}

// Select draws the place with id as selected (-1: none).
func (v *MapView) Select(id int) {
	v.selected = id
	v.Update()
}

// OnClick calls f with the picture coordinates of each click on the map.
func (v *MapView) OnClick(f func(x, y int)) { v.onClick = f }

// OnMove makes places draggable: f gets a moved place's ID and its new
// picture coordinates when it is let go.
func (v *MapView) OnMove(f func(id, x, y int)) {
	v.onMove = f
	v.SetMouseTracking(f != nil) // for the hand over places
}

// OnContext calls f when a place is right-clicked, with where (on the
// screen) to show a menu.
func (v *MapView) OnContext(f func(id int, at *qt.QPoint)) { v.onContext = f }

// OnDelete calls f when Delete (or Backspace) is pressed on the map.
func (v *MapView) OnDelete(f func()) { v.onDelete = f }

// Near is the place near a click at picture coordinates (x, y).
func (v *MapView) Near(x, y int) (Place, bool) {
	// (a view not laid out yet has a tiny scale: keep the radius sane)
	return Nearest(v.places, x, y, clickNear/max(v.fit().Scale, 0.2))
}

func (v *MapView) fit() Fit {
	if v.image == nil {
		return Fit{Scale: 1}
	}
	return NewFit(v.image.Width(), v.image.Height(), v.Width(), v.Height())
}

func (v *MapView) paint() {
	p := qt.NewQPainter2(v.QPaintDevice)
	defer p.End()
	pal := v.Palette()
	p.FillRect(qt.NewQRectF4(0, 0, float64(v.Width()), float64(v.Height())), pal.Base())
	if v.image == nil {
		p.SetPen(pal.Text().Color())
		p.DrawText5(qt.NewQRectF4(0, 0, float64(v.Width()), float64(v.Height())), int(qt.AlignCenter),
			"Choose a map above to place the places on.")
		return
	}
	p.SetRenderHint(qt.QPainter__SmoothPixmapTransform)
	f := v.fit()
	p.DrawImage5(qt.NewQRectF4(f.DX, f.DY, float64(v.image.Width())*f.Scale, float64(v.image.Height())*f.Scale), v.image)
	p.SetRenderHint(qt.QPainter__Antialiasing)

	outline := qt.NewQPen4(qt.NewQBrush3(qt.NewQColor3(255, 255, 255)), 2)
	normal := qt.NewQBrush3(qt.NewQColor3(33, 102, 172))
	chosen := qt.NewQBrush3(qt.NewQColor3(214, 39, 40))
	for _, pl := range v.places {
		x, y := f.ToView(pl.X, pl.Y)
		r, brush := dotRadius, normal
		if pl.ID == v.selected {
			r, brush = dotRadius*1.6, chosen
		}
		p.SetPenWithPen(outline)
		p.SetBrush(brush)
		p.DrawEllipse3(qt.NewQPointF3(x, y), r, r)
		if v.showNames && pl.Name != "" {
			label(p, x+r+3, y+4, pl.Name)
		}
	}
}

// label writes text with a light halo, readable on any map.
func label(p *qt.QPainter, x, y float64, text string) {
	p.SetPen(qt.NewQColor3(255, 255, 255))
	for _, d := range [][2]float64{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		p.DrawText(qt.NewQPointF3(x+d[0], y+d[1]), text)
	}
	p.SetPen(qt.NewQColor3(20, 20, 20))
	p.DrawText(qt.NewQPointF3(x, y), text)
}
