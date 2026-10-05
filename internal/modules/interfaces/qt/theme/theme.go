// Package theme sets the desktop's colours: the system's (Qt's choice),
// light, or dark (OpenTeacher had a dark theme profile). Charts, maps and
// previews follow the palette.
package theme

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
)

// Setting is the theme setting's key; its values are the theme names.
const Setting = "org.recuerdo.theme"

// The themes.
// i18n:values
const (
	System = "System"
	Light  = "Light"
	Dark   = "Dark"
)

func init() {
	settingsdefs.Register(settingsdefs.Def{
		Key: Setting, Category: "Interface", Name: "Theme",
		Help: "The colours of Recuerdo: as the system has them, light or dark", Kind: settingsdefs.Choice,
		Choices: []string{System, Light, Dark}, Default: System,
	})
}

var systemStyle, systemPalette = "", (*qt.QPalette)(nil)

// Apply sets a theme for the whole application.
func Apply(name string) {
	if systemPalette == nil { // remember the system's look, to go back to it
		systemPalette = qt.NewQPalette7(qt.QGuiApplication_Palette()) // a copy
		systemStyle = qt.QApplication_Style().Name()
	}
	switch name {
	case Dark:
		qt.QApplication_SetStyleWithStyle("Fusion")
		qt.QApplication_SetPalette(dark())
	case Light:
		style := qt.QApplication_SetStyleWithStyle("Fusion")
		qt.QApplication_SetPalette(style.StandardPalette())
	default:
		if systemStyle != "" {
			qt.QApplication_SetStyleWithStyle(systemStyle)
		}
		qt.QApplication_SetPalette(systemPalette)
	}
}

func dark() *qt.QPalette {
	p := qt.NewQPalette()
	c := func(r, g, b int) *qt.QColor { return qt.NewQColor3(r, g, b) }
	set := func(role qt.QPalette__ColorRole, color *qt.QColor) { p.SetColor2(role, color) }
	set(qt.QPalette__Window, c(37, 40, 45))
	set(qt.QPalette__WindowText, c(230, 233, 236))
	set(qt.QPalette__Base, c(27, 30, 34))
	set(qt.QPalette__AlternateBase, c(45, 49, 55))
	set(qt.QPalette__ToolTipBase, c(45, 49, 55))
	set(qt.QPalette__ToolTipText, c(230, 233, 236))
	set(qt.QPalette__Text, c(230, 233, 236))
	set(qt.QPalette__PlaceholderText, c(140, 148, 158))
	set(qt.QPalette__Button, c(52, 57, 64))
	set(qt.QPalette__ButtonText, c(230, 233, 236))
	set(qt.QPalette__BrightText, c(255, 110, 100))
	set(qt.QPalette__Link, c(100, 175, 210))
	set(qt.QPalette__Highlight, c(75, 163, 195))
	set(qt.QPalette__HighlightedText, c(15, 18, 22))
	set(qt.QPalette__Mid, c(70, 76, 84))
	set(qt.QPalette__Dark, c(20, 22, 25))
	set(qt.QPalette__Light, c(80, 86, 95))
	for _, role := range []qt.QPalette__ColorRole{qt.QPalette__WindowText, qt.QPalette__Text, qt.QPalette__ButtonText} {
		p.SetColor(qt.QPalette__Disabled, role, c(120, 126, 134))
	}
	return p
}
