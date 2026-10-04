package media

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestReverse(t *testing.T) {
	list := lesson.WordList{Items: []lesson.WordItem{
		lesson.MediaItem(0, "dog", "resources/dog.png", false, "Which animal?", "dog"),
		lesson.MediaItem(1, "site", "https://openteacher.org", true, "", "OpenTeacher"),
	}}
	Reverse(&list)
	a, b := list.Items[0], list.Items[1]
	if a.Questions[0] != "dog" || a.Answers[0] != "Which animal?" || *a.Filename != "resources/dog.png" ||
		b.Questions[0] != "OpenTeacher" || b.Answers != nil || !*b.Remote || a.Name != "dog" {
		t.Errorf("reversed: %+v", list.Items)
	}
	if m := NewMediaReverserModule(); m.DataType() != "media" || m.GetType() != "reverser" {
		t.Errorf("module %q %q", m.DataType(), m.GetType())
	}
}
