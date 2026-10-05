// Package openteacherauthors lists the people who made OpenTeacher, which
// Recuerdo is based on, for the credits. Port of OpenTeacher's
// data/openteacherAuthors.
package openteacherauthors

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/langcode"
	"github.com/LaPingvino/recuerdo/internal/resources"
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
	// the translators, per language (data/translators.txt, made by
	// recuerdo-dev translators from OpenTeacher's .po files)
	for _, t := range Translators() {
		role := "Translator (" + t.Language + ")"
		roles = append(roles, role)
		names[role] = t.Names
	}
	return roles, names
}

// Translation is who translated OpenTeacher into a language.
type Translation struct {
	Language string // its English name, with the code for a variant: "English (en_GB)"
	Names    []string
}

// Translators reads data/translators.txt ("nl: Marten de Vries, ...").
func Translators() []Translation {
	data, err := os.ReadFile(filepath.Join(resources.Dir(), "data", "translators.txt"))
	if err != nil {
		return nil
	}
	var out []Translation
	for _, line := range strings.Split(string(data), "\n") {
		code, list, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		code = strings.TrimSpace(code)
		lang := langcode.Name(strings.SplitN(code, "_", 2)[0])
		if lang == "" {
			lang = code
		} else if strings.Contains(code, "_") {
			lang += " (" + code + ")"
		}
		var names []string
		for _, n := range strings.Split(list, ",") {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
		out = append(out, Translation{Language: lang, Names: names})
	}
	return out
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
