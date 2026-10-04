package lesson

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// sampleFiles are OpenTeacher's loader test files, one or more for each
// format it reads.
const sampleFiles = "../../legacy/modules/org/openteacher/logic/loaders/test/testFiles"

// knownBroken lists sample files Recuerdo does not load correctly yet,
// with what goes wrong. TestOpenTeacherSampleFiles fails when one starts
// working, so the list only shrinks.
var knownBroken = map[string]string{
	"application_x-oriente-voca.voca3.0.wdl":     "Voca format not supported",
	"application_x-oriente-voca.voca4.0.wdl":     "Voca format not supported",
	"application_x-oriente-voca.vocatude1.0.wdl": "Voca format not supported",
}

// firstQuestion is the first question of sample files whose content
// would otherwise pass the generic checks while being read wrongly.
var firstQuestion = map[string]string{
	"text_csv.openteacher3x.csv": "een", // its first row names the languages
	"text_csv.teach2000.csv":     "een", // its first row is a header in Dutch
}

// notLessons are sample files that are not word lists.
var notLessons = map[string]bool{"COPYING": true, "netherlands.png": true}

// xmlMarkup matches XML tags, declarations and comments.
var xmlMarkup = regexp.MustCompile(`<[?!/]?[A-Za-z-]|-->`)

// garbage reports text that cannot come from a word list: invalid UTF-8,
// control characters or XML markup (a file read as text).
func garbage(s string) bool {
	if !utf8.ValidString(s) {
		return true
	}
	if xmlMarkup.MatchString(s) {
		return true
	}
	return strings.IndexFunc(s, func(r rune) bool { return r < 0x20 && r != '\t' && r != '\n' }) >= 0
}

func checkSample(path string) string {
	data, err := NewFileLoader().LoadFile(path)
	if err != nil {
		return "error: " + err.Error()
	}
	if len(data.List.Items) == 0 {
		return "no items"
	}
	if want, ok := firstQuestion[filepath.Base(path)]; ok {
		if q := data.List.Items[0].Questions; len(q) == 0 || q[0] != want {
			return fmt.Sprintf("first question %q, want %q", q, want)
		}
	}
	ids := map[int]bool{}
	for _, item := range data.List.Items {
		if ids[item.ID] {
			return "duplicate item ID"
		}
		ids[item.ID] = true
		if len(item.Questions) == 0 && len(item.Answers) == 0 {
			return "empty item"
		}
		for _, s := range append(append([]string{}, item.Questions...), item.Answers...) {
			if garbage(s) {
				return "garbage text " + strings.ToValidUTF8(s[:min(len(s), 20)], "?")
			}
		}
	}
	return ""
}

// TestOpenTeacherSampleFiles loads each of OpenTeacher's sample files, as
// OpenTeacher's own loader test does: each must give items with unique
// IDs, here also without garbage text.
func TestOpenTeacherSampleFiles(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(sampleFiles, "*"))
	if err != nil || len(files) == 0 {
		t.Skip("OpenTeacher sample files not found")
	}
	for _, f := range files {
		name := filepath.Base(f)
		if notLessons[name] {
			continue
		}
		t.Run(name, func(t *testing.T) {
			problem := checkSample(f)
			reason, broken := knownBroken[name]
			switch {
			case broken && problem == "":
				t.Errorf("loads correctly now: remove it from knownBroken (%s)", reason)
			case broken:
				t.Skipf("known broken (%s): %s", reason, problem)
			case problem != "":
				t.Error(problem)
			}
		})
	}
}

func TestNonLessonFilesAreRefused(t *testing.T) {
	png := filepath.Join(sampleFiles, "netherlands.png")
	if _, err := os.Stat(png); err != nil {
		t.Skip("sample files not found")
	}
	if _, err := NewFileLoader().LoadFile(png); err == nil {
		t.Error("a PNG image loaded as a word list")
	}
}
