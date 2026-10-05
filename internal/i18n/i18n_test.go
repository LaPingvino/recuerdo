package i18n

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

func TestReadPO(t *testing.T) {
	dir := t.TempDir()
	po := `# a comment
msgid ""
msgstr "Content-Type: text/plain; charset=UTF-8\n"

msgid "&File"
msgstr "&Bestand"

#, fuzzy
msgid "Unsure"
msgstr "Onzeker"

msgid ""
"Two "
"lines"
msgstr "Twee "
"regels"

msgid "Untranslated"
msgstr ""
`
	os.WriteFile(filepath.Join(dir, "nl.po"), []byte(po), 0o644)
	got, err := ReadPO(filepath.Join(dir, "nl.po"))
	want := map[string]string{"&File": "&Bestand", "Two lines": "Twee regels"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("%v %v", got, err)
	}
}

func TestResolveAndUse(t *testing.T) {
	dir := t.TempDir()
	for _, l := range []string{"nl", "pt_BR", "en_GB"} {
		os.WriteFile(filepath.Join(dir, l+".po"), []byte("msgid \"&File\"\nmsgstr \""+l+"\"\n"), 0o644)
	}
	for lang, want := range map[string]string{
		"nl": "nl", "nl_BE.UTF-8": "nl", "pt": "pt_BR", "pt-BR": "pt_BR", "en_GB": "en_GB",
		"en": "", "en_US.UTF-8": "", "C": "", "": "", "fr": "", "de@euro": "",
	} {
		if got := Resolve(dir, lang); got != want {
			t.Errorf("Resolve(%q) = %q, want %q", lang, got, want)
		}
	}
	if err := Use(dir, "nl_NL"); err != nil || T("&File") != "nl" || T("Unknown") != "Unknown" || Current() != "nl" {
		t.Errorf("nl: %q %v", T("&File"), err)
	}
	Use(dir, "")
	if T("&File") != "&File" || Current() != "" {
		t.Error("English should not translate")
	}
}

func TestSystemLanguage(t *testing.T) {
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "nl_NL.UTF-8")
	if l := SystemLanguage(); l != "nl_NL" {
		t.Errorf("%q", l)
	}
	t.Setenv("LANGUAGE", "fy:nl")
	if l := SystemLanguage(); l != "fy" {
		t.Errorf("LANGUAGE: %q", l)
	}
}

// The merged translations in data/translations all read.
func TestShippedTranslations(t *testing.T) {
	dir := filepath.Join("..", "..", "data", "translations")
	langs := Available(dir)
	if len(langs) < 25 {
		t.Fatalf("%d languages", len(langs))
	}
	for _, l := range langs {
		c, err := ReadPO(filepath.Join(dir, l+".po"))
		if err != nil || len(c) < 10 {
			t.Errorf("%s: %d texts, %v", l, len(c), err)
		}
	}
	if err := Use(dir, "nl"); err != nil || T("&File") != "&Bestand" || T("Question language:") != "Taal van de vragen:" {
		t.Errorf("Dutch: %q %q %v", T("&File"), T("Question language:"), err)
	}
	// "..." added back; Recuerdo's wording through an alias, colon dropped
	for msgid, want := range map[string]string{"&Open...": "&Open...", "&Settings...": "&Instellingen...",
		"E&xit": "A&fsluiten", "Open Recent": "Recent geopend", "Question": "Vraag", "No such text": "No such text"} {
		if got := T(msgid); got != want {
			t.Errorf("T(%q) = %q, want %q", msgid, got, want)
		}
	}
	// a full-width colon is dropped too
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "zh_CN.po"), []byte("msgid \"Recently opened:\"\nmsgstr \"最近打开的：\"\n"), 0o644)
	if err := Use(tmp, "zh_CN"); err != nil || T("Open Recent") != "最近打开的" {
		t.Errorf("Chinese: %q %v", T("Open Recent"), err)
	}
	Use(dir, "")
}

func TestRecuerdoOverOpenTeacher(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "nl.po"), []byte("msgid \"&File\"\nmsgstr \"&Bestand\"\n\nmsgid \"Teach\"\nmsgstr \"Leer\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "recuerdo-nl.po"), []byte("msgid \"Teach\"\nmsgstr \"Oefenen\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "recuerdo-eo.po"), []byte("msgid \"Teach\"\nmsgstr \"Ekzerci\"\n"), 0o644)
	if got := Available(dir); !reflect.DeepEqual(got, []string{"eo", "nl"}) {
		t.Errorf("available %v", got)
	}
	Use(dir, "nl")
	if T("&File") != "&Bestand" || T("Teach") != "Oefenen" {
		t.Errorf("nl: %q %q", T("&File"), T("Teach"))
	}
	Use(dir, "eo")
	if T("Teach") != "Ekzerci" {
		t.Errorf("eo (own file only): %q", T("Teach"))
	}
	Use(dir, "")
}

// Recuerdo's own translations belong to texts in recuerdo.pot and keep
// their placeholders (%s, %d, %[1]s ...), so fmt never garbles them.
func TestRecuerdoTranslations(t *testing.T) {
	dir := filepath.Join("..", "..", "data", "translations")
	pot, err := ReadPOT(filepath.Join(dir, "recuerdo.pot"))
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "recuerdo-*.po"))
	if len(files) == 0 {
		t.Fatal("no recuerdo-*.po")
	}
	verbs := regexp.MustCompile(`%(\[\d+\])?[-+# 0]*\d*(\.\d+)?[a-zA-Z%]`)
	kinds := func(s string) []string {
		var out []string
		for _, v := range verbs.FindAllString(s, -1) {
			out = append(out, v[len(v)-1:])
		}
		sort.Strings(out)
		return out
	}
	for _, f := range files {
		c, err := ReadPO(f)
		if err != nil {
			t.Fatal(err)
		}
		for msgid, msgstr := range c {
			if !pot[msgid] {
				t.Errorf("%s: %q is not in recuerdo.pot (stale: rerun scripts/extract_strings.py)", filepath.Base(f), msgid)
			}
			if !reflect.DeepEqual(kinds(msgid), kinds(msgstr)) {
				t.Errorf("%s: %q -> %q: placeholders differ", filepath.Base(f), msgid, msgstr)
			}
		}
	}
}

func TestRightToLeft(t *testing.T) {
	for lang, want := range map[string]bool{"ar": true, "ur": true, "ur_PK": true, "he": true, "nl": false, "en_GB": false, "": false} {
		if got := RightToLeft(lang); got != want {
			t.Errorf("RightToLeft(%q) = %v", lang, got)
		}
	}
}
