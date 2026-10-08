package topo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/osm"
)

// The map of a topography lesson from OpenStreetMap: search an area by
// name, choose how it looks (without names by default: they would give
// the answers away), and Recuerdo draws or downloads it. Places typed by
// name are then put where they really are.

// mapSide is the longest side of a map made from OpenStreetMap, in pixels.
const mapSide = 1400

// chooseOSMMap shows the dialog to make the lesson's map from
// OpenStreetMap.
func (w *TopoLessonWidget) chooseOSMMap() {
	d := qt.NewQDialog(w.QWidget)
	d.SetWindowTitle(i18n.T("Map from OpenStreetMap"))
	d.Resize(520, 420)
	layout := qt.NewQVBoxLayout(d.QWidget)

	row := qt.NewQHBoxLayout2()
	query := qt.NewQLineEdit(d.QWidget)
	query.SetPlaceholderText(i18n.T("A country, a province, a region: Netherlands, Bavaria, Andes"))
	row.AddWidget(query.QWidget)
	searchBtn := qt.NewQPushButton3(i18n.T("Search"))
	row.AddWidget(searchBtn.QWidget)
	layout.AddLayout(row.QLayout)

	results := qt.NewQListWidget(d.QWidget)
	layout.AddWidget(results.QWidget)

	styleRow := qt.NewQHBoxLayout2()
	styleRow.AddWidget(qt.NewQLabel3(i18n.T("Map:")).QWidget)
	style := qt.NewQComboBox(d.QWidget)
	for _, s := range osm.Styles {
		style.AddItem(i18n.T(s.Name))
	}
	styleRow.AddWidget(style.QWidget)
	styleRow.AddStretch()
	layout.AddLayout(styleRow.QLayout)

	status := qt.NewQLabel3(i18n.T("Map data © OpenStreetMap contributors."))
	status.SetWordWrap(true)
	layout.AddWidget(status.QWidget)

	buttons := qt.NewQDialogButtonBox(d.QWidget)
	useBtn := buttons.AddButton2(i18n.T("Use this map"), qt.QDialogButtonBox__AcceptRole)
	useBtn.SetEnabled(false)
	buttons.AddButtonWithButton(qt.QDialogButtonBox__Cancel)
	buttons.OnRejected(func() { d.Reject() })
	layout.AddWidget(buttons.QWidget)

	var areas []osm.Area
	busy := func(on bool) {
		searchBtn.SetEnabled(!on)
		useBtn.SetEnabled(!on && results.CurrentRow() >= 0)
	}
	search := func() {
		q := query.Text()
		if q == "" {
			return
		}
		busy(true)
		status.SetText(i18n.T("Searching…"))
		go func() {
			found, err := osm.Search(context.Background(), q)
			mainthread.Start(func() {
				busy(false)
				results.Clear()
				areas = nil
				if err != nil {
					status.SetText(i18n.Tf("Could not search: %v", err))
					return
				}
				for _, a := range found {
					if len(a.Border) > 0 { // areas, not points
						areas = append(areas, a)
						results.AddItem(a.Name)
					}
				}
				if len(areas) == 0 {
					status.SetText(i18n.T("No area of that name: try a country, province or region."))
					return
				}
				results.SetCurrentRow(0)
				status.SetText(i18n.T("Map data © OpenStreetMap contributors."))
			})
		}()
	}
	searchBtn.OnClicked(search)
	query.OnReturnPressed(search)
	results.OnCurrentRowChanged(func(int) { busy(false) })
	useBtn.OnClicked(func() {
		row := results.CurrentRow()
		if row < 0 || row >= len(areas) {
			return
		}
		a, s := areas[row].Main(), osm.Styles[style.CurrentIndex()]
		g := osm.Fit(a, mapSide)
		busy(true)
		go func() {
			img, err := osm.Map(context.Background(), g, a, s, func(done, all int) {
				mainthread.Start(func() { status.SetText(i18n.Tf("Making the map: %d of %d", done, all)) })
			})
			var data []byte
			if err == nil {
				var b bytes.Buffer
				if err = png.Encode(&b, img); err == nil {
					data = b.Bytes()
				}
			}
			mainthread.Start(func() {
				busy(false)
				if err != nil {
					status.SetText(i18n.Tf("Could not make the map: %v", err))
					return
				}
				w.useOSMMap(data, g)
				d.Accept()
			})
		}()
	})
	d.Exec()
}

// useOSMMap makes a map made from OpenStreetMap the lesson's, with where
// it lies on Earth.
func (w *TopoLessonWidget) useOSMMap(img []byte, g osm.Geo) {
	geo, _ := json.Marshal(g)
	w.lesson.Data.Resources[lesson.MapImageResource] = img
	w.lesson.Data.Resources[lesson.MapGeoResource] = geo
	w.loadMap()
	w.modified()
	w.refresh()
	w.enterHint.SetText(i18n.T("Type the name of a place on this map: Recuerdo puts it where it is."))
}

// geo is where the lesson's map lies on Earth, if it was made from
// OpenStreetMap.
func (w *TopoLessonWidget) geo() (osm.Geo, bool) {
	b, ok := w.lesson.Data.Resources[lesson.MapGeoResource].([]byte)
	if !ok {
		return osm.Geo{}, false
	}
	var g osm.Geo
	return g, json.Unmarshal(b, &g) == nil && g.Width > 0
}

// addFromOSM looks a place up within the map's area and adds it where it
// is; false if the map is not from OpenStreetMap.
func (w *TopoLessonWidget) addFromOSM(name string) bool {
	g, ok := w.geo()
	if !ok {
		return false
	}
	w.enterHint.SetText(i18n.Tf("Looking for %s…", name))
	go func() {
		found, err := osm.SearchIn(context.Background(), name, g)
		mainthread.Start(func() {
			if err != nil {
				w.enterHint.SetText(fmt.Sprintf(i18n.T("Could not look it up (%v): click where %s is on the map."), err, name))
				return
			}
			for _, a := range found {
				if x, y, on := g.ToPicture(a.Lat, a.Lon); on {
					w.add(name, x, y)
					w.nameEdit.Clear()
					return
				}
			}
			w.enterHint.SetText(fmt.Sprintf(i18n.T("Now click where %s is on the map."), name))
		})
	}()
	return true
}
