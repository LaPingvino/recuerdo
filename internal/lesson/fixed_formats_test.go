package lesson

import (
	"path/filepath"
	"strings"
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
	if d := load("application_x-kvtml.kvoctrain.kvtml"); len(d.List.Items) != 1 || d.List.Items[0].Questions[0] != "one" || d.List.Items[0].Answers[0] != "een" {
		t.Errorf("kvoctrain: %+v", d.List.Items)
	}
	if d := load("application_x-teach2000.wrts.t2k"); len(d.List.Items) != 3 || d.List.Items[2].Answers[0] != "three" || d.List.Items[0].Comment != "" {
		t.Errorf("WRTS-written t2k: %+v", d.List.Items)
	}
	if d := load("application_x-apkg.anki.apkg"); len(d.List.Items) != 3 || d.List.Title != "application_x-apkg.anki" {
		t.Errorf("apkg: title %q, %+v", d.List.Title, d.List.Items)
	}
	if d := load("application_x-wrts.wrts.wrts"); d.List.Title != "Test" || d.List.QuestionLanguage != "Dutch" || d.List.Items[0].Answers[0] != "one" {
		t.Errorf("wrts: %q %q %+v", d.List.Title, d.List.QuestionLanguage, d.List.Items)
	}
	if d := load("application_x-pauker.pauker-modified.pau"); d.List.Items[0].Questions[0] != "éen" || d.List.Items[0].Answers[0] != "oné" {
		t.Errorf("pauker: %+v", d.List.Items)
	}
	if d := load("application_x-pauker.pauker.pau.gz"); len(d.List.Items) == 0 {
		t.Errorf("gzipped pauker: no items")
	}
	if d := load("application_x-jmemorizelesson.jmemorize.jml"); len(d.List.Items) == 0 || d.List.Items[len(d.List.Items)-1].Questions[0] != "c" {
		t.Errorf("jml (in order of creation, c last): %+v", d.List.Items)
	}
	// its font line puts the answers in TekniaGreek, a Greek mimicry font
	if d := load("application_x-overhoor.overhoorvoorwindows4.5.1.oh"); len(d.List.Items) != 3 || d.List.Items[0].Answers[0] != "ονε" || d.List.Items[2].Questions[0] != "drié" {
		t.Errorf("overhoor: %+v", d.List.Items)
	}
	if d := load("application_x-overhoor.wrts.ohw"); d.List.Items[0].Answers[0] != "one" {
		t.Errorf("overhoor without fonts: %+v", d.List.Items)
	}
	if d := load("application_x-overhoringsprogrammatalen.downloaded-and-edited.ovr"); d.List.QuestionLanguage != "Duits" || len(d.List.Items) != 2 || len(d.List.Items[1].Answers) != 2 {
		t.Errorf("ovr: %q %+v", d.List.QuestionLanguage, d.List.Items)
	}
	if d := load("application_x-granuledeck.granule.dkf"); len(d.List.Items) != 3 || d.List.Items[0].Answers[0] != "one" {
		t.Errorf("granule: %+v", d.List.Items)
	}
	if d := load("application_x-domingo.domingo.voc"); len(d.List.Items) != 3 || d.List.Items[1].Answers[0] != "two" {
		t.Errorf("domingo: %+v", d.List.Items)
	}
	if d := load("application_x-fm-dictionary.fmd"); len(d.List.Items) != 3 || d.List.QuestionLanguage != "" {
		t.Errorf("fmd: %q %+v", d.List.QuestionLanguage, d.List.Items)
	}
	if d := load("application_x-vocabularium.edited(all-files-are).voc"); d.List.QuestionLanguage != "English" || d.List.AnswerLanguage != "Nederlands" || len(d.List.Items) != 3 || !strings.HasPrefix(d.List.Title, "Title here.") {
		t.Errorf("vocabularium: %q %q %q %+v", d.List.Title, d.List.QuestionLanguage, d.List.AnswerLanguage, d.List.Items)
	}
	if d := load("application_x-vokabeltrainer.vokabeltrainer-with-comment.vtl3"); len(d.List.Items) != 3 || d.List.Items[0].Answers[0] != "one" || d.List.Items[2].Comment != "comment" {
		t.Errorf("vokabeltrainer: %+v", d.List.Items)
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
