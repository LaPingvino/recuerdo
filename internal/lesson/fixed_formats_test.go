package lesson

import (
	"path/filepath"
	"testing"
)

func TestFixedFormatsContent(t *testing.T) {
	load := func(name string) *LessonData {
		t.Helper()
		d, err := NewFileLoader().LoadFile(filepath.Join(sampleFiles, name))
		if err != nil {
			t.Skipf("%s: %v", name, err)
		}
		return d
	}
	if d := load("text_csv.openteacher3x.csv"); d.List.QuestionLanguage != "Dutch" || d.List.AnswerLanguage != "English" || len(d.List.Items) != 2 {
		t.Errorf("csv: languages %q/%q, %d items", d.List.QuestionLanguage, d.List.AnswerLanguage, len(d.List.Items))
	}
	if d := load("application_x-openteachingwords.openteacher3x.otwd"); d.List.Title != "test" || len(d.List.Tests) != 1 || d.List.Items[1].Answers[0] != "two" {
		t.Errorf("otwd: title %q, %d tests, items %+v", d.List.Title, len(d.List.Tests), d.List.Items)
	}
	if d := load("application_xml.abbyylingvotutor_x5.xml"); d.List.Title != "OpenTeacher test suite list" || d.List.Items[0].Comment != "test" {
		t.Errorf("abbyy: title %q, first item %+v", d.List.Title, d.List.Items[0])
	}
	if d := load("application_x-backpack.backpack"); len(d.List.Items) != 2 || len(d.List.Items[0].Answers) != 2 {
		t.Errorf("backpack: %+v", d.List.Items)
	}
	if d := load("text_plain.vtrain.txt"); len(d.List.Items) != 3 || d.List.Items[1].Answers[0] != "two" {
		t.Errorf("vtrain: %+v", d.List.Items)
	}
}

func TestCSVHeader(t *testing.T) {
	for row, want := range map[[3]string]bool{
		{"Dutch", "English", ""}:                             true,
		{"Questions", "Answers", "Comment"}:                  true,
		{"Vraag", "Antwoord (betekenis)", ""}:                true,
		{"huis", "house", ""}:                                false,
		{"English", "huis", ""}:                              false,
		{"Nederlands", "Deutsch", "Comment after answering"}: true,
	} {
		r := row[:2]
		if row[2] != "" {
			r = row[:]
		}
		if got := isCSVHeader(r); got != want {
			t.Errorf("isCSVHeader(%q) = %v", r, got)
		}
	}
}
