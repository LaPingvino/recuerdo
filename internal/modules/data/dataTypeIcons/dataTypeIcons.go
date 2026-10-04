// Package datatypeicons has an icon for each type of lesson (words,
// topography, media), from KDE's Oxygen icon theme (see icons/COPYING).
// Port of OpenTeacher's data/dataTypeIcons.
package datatypeicons

import (
	"embed"

	"github.com/LaPingvino/recuerdo/internal/core"
)

//go:embed icons/*.png icons/COPYING
var icons embed.FS

// Icon returns the PNG icon for a lesson type ("words", "topo", "media"),
// or nil for other types.
func Icon(dataType string) []byte {
	b, err := icons.ReadFile("icons/" + dataType + ".png")
	if err != nil {
		return nil
	}
	return b
}

// DataTypeIconsModule offers the icons as an OpenTeacher "dataTypeIcons"
// module.
type DataTypeIconsModule struct {
	*core.BaseModule
}

// NewDataTypeIconsModule creates the module.
func NewDataTypeIconsModule() *DataTypeIconsModule {
	return &DataTypeIconsModule{BaseModule: core.NewBaseModule("data", "datatypeicons-module")}
}

// FindIcon returns the icon for a lesson type, or nil.
func (mod *DataTypeIconsModule) FindIcon(dataType string) []byte { return Icon(dataType) }

// InitDataTypeIconsModule creates the module.
func InitDataTypeIconsModule() core.Module { return NewDataTypeIconsModule() }
