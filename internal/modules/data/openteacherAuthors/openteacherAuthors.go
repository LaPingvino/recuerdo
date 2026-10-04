// Package openteacherauthors lists the people who made OpenTeacher, which
// Recuerdo is based on, for the credits. Port of OpenTeacher's
// data/openteacherAuthors.
package openteacherauthors

import (
	"github.com/LaPingvino/recuerdo/internal/core"
)

// Author is someone who contributed to OpenTeacher, and how.
type Author struct {
	Role, Name string
}

// Authors are OpenTeacher's contributors, in OpenTeacher's order.
var Authors = []Author{
	{"Core developer", "Milan Boers"},
	{"Core developer", "Cas Widdershoven"},
	{"Core developer", "Marten de Vries"},
	{"Patches contributor", "Roel Huybrechts"},
	{"Patches contributor", "David D Lowe"},
	{"Debian/Ubuntu packager", "Charlie Smotherman"},
	{"Artwork", "Yordi de Graaf"},
	{"Artwork", "Oxygen icon theme"},
	{"Topography maps", "Wikimedia Commons"},
	{"Chat channel spammer", "Stefan de Vries"},
}

// ByRole groups the authors by role, roles in order of first appearance.
func ByRole() (roles []string, names map[string][]string) {
	names = map[string][]string{}
	for _, a := range Authors {
		if _, ok := names[a.Role]; !ok {
			roles = append(roles, a.Role)
		}
		names[a.Role] = append(names[a.Role], a.Name)
	}
	return roles, names
}

// OpenTeacherAuthorsModule offers the authors as an OpenTeacher
// "openteacherAuthors" module.
type OpenTeacherAuthorsModule struct {
	*core.BaseModule
}

// NewOpenTeacherAuthorsModule creates the module.
func NewOpenTeacherAuthorsModule() *OpenTeacherAuthorsModule {
	return &OpenTeacherAuthorsModule{BaseModule: core.NewBaseModule("data", "openteacherauthors-module")}
}

// Authors returns OpenTeacher's contributors.
func (mod *OpenTeacherAuthorsModule) Authors() []Author { return Authors }

// InitOpenTeacherAuthorsModule creates the module.
func InitOpenTeacherAuthorsModule() core.Module { return NewOpenTeacherAuthorsModule() }
