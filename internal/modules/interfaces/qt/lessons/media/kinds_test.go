package media

import (
	"testing"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func TestKind(t *testing.T) {
	for _, c := range []struct {
		name   string
		remote bool
		want   string
	}{
		{"resources/Dog.JPG", false, Image},
		{"notes.txt", false, Text},
		{"bark.mp3", false, Audio},
		{"run.mp4", false, Video},
		{"data.xyz", false, Unknown},
		{"https://www.youtube.com/watch?v=abc", true, YouTube},
		{"https://youtu.be/abc", true, YouTube},
		{"https://vimeo.com/123", true, Vimeo},
		{"http://www.dailymotion.com/video/x7", true, Dailymotion},
		{"http://openteacher.org/", true, Website},
		{"https://example.org/dog.png?size=2", true, Image},
		{"https://example.org/page.txt", true, Website},
	} {
		if got := Kind(c.name, c.remote); got != c.want {
			t.Errorf("Kind(%q) = %q, want %q", c.name, got, c.want)
		}
	}
	if !Shown(Image) || Shown(Video) || !AddressOK("https://a.org") || AddressOK("a.org") {
		t.Error("Shown/AddressOK")
	}
}

func TestResourceNameAndTeachable(t *testing.T) {
	files := map[string][]byte{"resources/dog.jpg": {1}, "resources/dog-2.jpg": {1}}
	if n := ResourceName("/home/x/dog.jpg", files); n != "resources/dog-3.jpg" {
		t.Errorf("%q", n)
	}
	if n := ResourceName(`C:\pics\cat.png`, files); n != "resources/cat.png" {
		t.Errorf("%q", n)
	}
	list := lesson.WordList{Items: []lesson.WordItem{
		lesson.MediaItem(0, "a", "a.png", false, "q", "ans"), lesson.MediaItem(4, "b", "b.png", false, "", ""),
	}}
	if got := Teachable(list); len(got.Items) != 1 || NextID(list.Items) != 5 {
		t.Errorf("teachable %+v", got.Items)
	}
}
