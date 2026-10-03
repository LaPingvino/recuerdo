// Package icon gives Recuerdo's icon as Qt pixmaps.
package icon

import (
	"github.com/LaPingvino/recuerdo/assets"
	"github.com/mappu/miqt/qt"
)

// Pixmap returns the icon scaled to size×size pixels (needs a
// QApplication).
func Pixmap(size int) *qt.QPixmap {
	pm := qt.NewQPixmap()
	pm.LoadFromDataWithData(assets.Icon)
	return pm.Scaled3(size, size, qt.KeepAspectRatio, qt.SmoothTransformation)
}

// Icon returns the icon for windows and the task bar.
func Icon() *qt.QIcon {
	pm := qt.NewQPixmap()
	pm.LoadFromDataWithData(assets.Icon)
	return qt.NewQIcon2(pm)
}
