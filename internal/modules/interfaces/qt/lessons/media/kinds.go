package media

import (
	"fmt"
	"path"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

// The kinds of media, after OpenTeacher's mediaTypes.
const (
	Image       = "picture"
	Text        = "text"
	Audio       = "sound"
	Video       = "video"
	YouTube     = "YouTube video"
	Vimeo       = "Vimeo video"
	Dailymotion = "Dailymotion video"
	Website     = "website"
	Unknown     = "file"
)

var extensions = map[string]string{
	".jpg": Image, ".jpeg": Image, ".png": Image, ".bmp": Image, ".gif": Image, ".tif": Image, ".tiff": Image,
	".webp": Image, ".svg": Image,
	".txt": Text,
	".mp3": Audio, ".wav": Audio, ".wma": Audio, ".aif": Audio, ".aiff": Audio, ".mid": Audio, ".midi": Audio,
	".aac": Audio, ".flac": Audio, ".ogg": Audio, ".oga": Audio, ".opus": Audio, ".m4a": Audio,
	".avi": Video, ".wmv": Video, ".flv": Video, ".mp4": Video, ".mpg": Video, ".mpeg": Video, ".mov": Video,
	".mkv": Video, ".webm": Video, ".ogv": Video,
}

// Kind is the kind of media at name: a file name or, when remote, a web
// address.
func Kind(name string, remote bool) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	if remote || strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		switch {
		case strings.Contains(lower, "youtube.") && strings.Contains(lower, "/watch?"),
			strings.Contains(lower, "youtu.be/"):
			return YouTube
		case strings.Contains(lower, "vimeo.com/"):
			return Vimeo
		case strings.Contains(lower, "dailymotion.com/video/"):
			return Dailymotion
		}
		// a web address of a file is that file (http://example.org/dog.jpg)
		if k, ok := extensions[path.Ext(strings.SplitN(lower, "?", 2)[0])]; ok && k != Text {
			return k
		}
		return Website
	}
	if k, ok := extensions[path.Ext(lower)]; ok {
		return k
	}
	return Unknown
}

// Shown reports whether a kind is shown in Recuerdo itself; the others
// open in the system's player or web browser.
func Shown(kind string) bool { return kind == Image || kind == Text }

// AddressOK reports whether text looks like a web address.
func AddressOK(text string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	return (strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://")) && strings.Contains(t[8:], ".")
}

// ResourceName is the name to embed a file under in a lesson
// ("resources/dog.jpg"), unique among files.
func ResourceName(base string, files map[string][]byte) string {
	base = path.Base(strings.ReplaceAll(base, "\\", "/"))
	name := "resources/" + base
	ext := path.Ext(base)
	for i := 2; files[name] != nil; i++ {
		name = fmt.Sprintf("resources/%s-%d%s", strings.TrimSuffix(base, ext), i, ext)
	}
	return name
}

// NextID is an ID not used by any of items.
func NextID(items []lesson.WordItem) int {
	id := 0
	for _, it := range items {
		if it.ID >= id {
			id = it.ID + 1
		}
	}
	return id
}

// Teachable are the items that have an answer to ask for.
func Teachable(list lesson.WordList) lesson.WordList {
	out := list
	out.Items = nil
	for _, it := range list.Items {
		if len(it.Answers) > 0 {
			out.Items = append(out.Items, it)
		}
	}
	return out
}
