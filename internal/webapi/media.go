package webapi

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/lesson/mediakind"
)

// Media lessons in the web version: pictures, sounds, videos, texts and
// websites, each with a question and an answer, as the desktop has them
// (lessons/media). The page shows an item's file from MediaFile; a remote
// item (a website or an online video) is its address.

// isMedia reports whether the open lesson is a media lesson.
func (a *App) isMedia() bool {
	if a.data == nil {
		return false
	}
	if _, ok := a.data.Resources[lesson.MediaFilesResource].(map[string][]byte); ok {
		return true
	}
	for _, it := range a.data.List.Items {
		if it.Filename != nil {
			return true
		}
	}
	return false
}

// mediaOf adds an item's media to it as the page has it.
func mediaOf(out *Item, it lesson.WordItem) {
	if it.Filename == nil {
		return
	}
	remote := it.Remote != nil && *it.Remote
	out.Name, out.MediaKind = it.Name, mediakind.Kind(*it.Filename, remote)
	if remote {
		out.Media = *it.Filename
	} else {
		out.File = *it.Filename
	}
}

// NewMedia starts a media lesson without items yet.
func (a *App) NewMedia(title string) (Lesson, error) {
	d := lesson.NewLessonData()
	d.List.Title = title
	lesson.MediaFiles(d)
	a.data, a.name, a.session = d, title+".otmd", nil
	return a.Lesson()
}

// AddMediaFile embeds a file (a picture, sound, video or text) as a new
// item named after it.
func (a *App) AddMediaFile(name string, data []byte) (Item, error) {
	if !a.isMedia() {
		return Item{}, errors.New("this lesson is not a media lesson")
	}
	files := lesson.MediaFiles(a.data)
	file := mediakind.ResourceName(name, files)
	files[file] = data
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	return a.addMedia(lesson.MediaItem(mediakind.NextID(a.data.List.Items), strings.TrimSuffix(base, path.Ext(base)), file, false, "", ""))
}

// AddMediaAddress adds a web address (a website or an online video).
func (a *App) AddMediaAddress(address string) (Item, error) {
	if !a.isMedia() {
		return Item{}, errors.New("this lesson is not a media lesson")
	}
	address = strings.TrimSpace(address)
	if !mediakind.AddressOK(address) {
		return Item{}, fmt.Errorf("%q is not a web address: it should start with https:// or http://", address)
	}
	return a.addMedia(lesson.MediaItem(mediakind.NextID(a.data.List.Items), address, address, true, "", ""))
}

func (a *App) addMedia(it lesson.WordItem) (Item, error) {
	a.data.List.Items = append(a.data.List.Items, it)
	return itemOf(it), nil
}

// SetMedia changes the name, question and answer of a media item.
func (a *App) SetMedia(id int, name, question, answer string) error {
	if a.data == nil {
		return ErrNoLesson
	}
	for i := range a.data.List.Items {
		if it := &a.data.List.Items[i]; it.ID == id {
			words := func(text string) []string {
				if text = strings.TrimSpace(text); text != "" {
					return []string{text}
				}
				return nil
			}
			it.Name, it.Questions, it.Answers = strings.TrimSpace(name), words(question), words(answer)
			return nil
		}
	}
	return fmt.Errorf("no item with id %d", id)
}

// MediaFile is an embedded file's bytes and its type, for the page to show
// (nil if the lesson has no such file).
func (a *App) MediaFile(file string) ([]byte, string) {
	if a.data == nil {
		return nil, ""
	}
	files, _ := a.data.Resources[lesson.MediaFilesResource].(map[string][]byte)
	b := files[file]
	if b == nil {
		return nil, ""
	}
	return b, mediaType(file, b)
}

// mediaType is a file's MIME type: from its extension for the kinds
// http.DetectContentType does not know.
func mediaType(file string, b []byte) string {
	switch strings.ToLower(path.Ext(file)) {
	case ".svg":
		return "image/svg+xml"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a", ".aac":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".opus", ".oga":
		return "audio/ogg"
	case ".mp4", ".mov":
		return "video/mp4"
	case ".mkv", ".webm":
		return "video/webm"
	case ".ogv":
		return "video/ogg"
	case ".txt":
		return "text/plain; charset=utf-8"
	}
	return http.DetectContentType(b)
}
