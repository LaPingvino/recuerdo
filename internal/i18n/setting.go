package i18n

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/resources"
	"github.com/LaPingvino/recuerdo/internal/settingsdefs"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Dir is where the translations are.
func Dir() string { return filepath.Join(resources.Dir(), "data", "translations") }

// Name is a language's name in that language ("nl" -> "Nederlands").
func Name(code string) string {
	tag, err := language.Parse(code)
	if err != nil {
		return code
	}
	if n := display.Self.Name(tag); n != "" {
		return n
	}
	return code
}

// The interface language, in the settings dialog.
func init() {
	settingsdefs.Register(settingsdefs.Def{
		Key: LanguageSetting, Category: "Interface", Name: "Language",
		Help: "The language of Recuerdo's menus and screens (after restarting Recuerdo). " +
			"Texts without a translation stay English.",
		Kind: settingsdefs.Choice, Default: "",
		// the installed languages, when the setting is shown
		Fill: func(d *settingsdefs.Def) {
			langs := Available(Dir())
			sort.Slice(langs, func(i, j int) bool { return Name(langs[i]) < Name(langs[j]) })
			d.Choices, d.Labels = []string{""}, []string{"System language"}
			for _, l := range append([]string{"en"}, langs...) {
				d.Choices = append(d.Choices, l)
				d.Labels = append(d.Labels, Name(l))
			}
		},
	})
}

// Start loads the language chosen in settings (or the system's).
func Start(store settingsdefs.Store) {
	lang, _ := store.GetSettingWithDefault(LanguageSetting, "").(string)
	if lang == "" {
		lang = SystemLanguage()
	}
	Use(Dir(), lang)
}

// RightToLeft reports whether lang (a code such as "ar" or "ur_PK") is
// written from right to left, so the interface is laid out mirrored.
func RightToLeft(lang string) bool {
	switch strings.SplitN(lang, "_", 2)[0] {
	case "ar", "ur", "fa", "he", "ps", "sd", "ug", "yi", "ckb", "dv":
		return true
	}
	return false
}
