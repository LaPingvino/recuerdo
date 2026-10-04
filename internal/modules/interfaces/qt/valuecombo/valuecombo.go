// Package valuecombo fills combo boxes whose items are values (lesson
// types, orders, notations, ...): each shows its translation, and the
// value is kept as the item's data, so code and settings keep working
// with the English values in any language.
package valuecombo

import (
	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/mappu/miqt/qt"
)

// Fill adds values to c, each shown translated.
func Fill(c *qt.QComboBox, values []string) {
	for _, v := range values {
		c.AddItem3(i18n.T(v), qt.NewQVariant14(v))
	}
}

// Value is the selected item's value ("" when none is).
func Value(c *qt.QComboBox) string {
	if c.CurrentIndex() < 0 {
		return ""
	}
	return c.CurrentData().ToString()
}

// Set selects the item with value v (nothing changes if there is none).
func Set(c *qt.QComboBox, v string) {
	if i := c.FindData(qt.NewQVariant14(v)); i >= 0 {
		c.SetCurrentIndex(i)
	}
}
