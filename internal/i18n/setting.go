package i18n

import (
	"path/filepath"
	"sort"

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
	langs := Available(Dir())
	choices, labels := []string{""}, []string{"System language"}
	sort.Slice(langs, func(i, j int) bool { return Name(langs[i]) < Name(langs[j]) })
	for _, l := range append([]string{"en"}, langs...) {
		choices = append(choices, l)
		labels = append(labels, Name(l))
	}
	settingsdefs.Register(settingsdefs.Def{
		Key: LanguageSetting, Category: "Interface", Name: "Language",
		Help: "The language of Recuerdo's menus and screens (after restarting Recuerdo). " +
			"Translations come from OpenTeacher; texts it did not have stay English.",
		Kind: settingsdefs.Choice, Choices: choices, Labels: labels, Default: "",
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
