package export

import (
	"fmt"
	"html"
	"path"
	"strings"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
)

// IsMedia reports whether data is a media lesson: its items have files or
// web addresses.
func IsMedia(data *lesson.LessonData) bool {
	for _, it := range data.List.Items {
		if it.Filename != nil {
			return true
		}
	}
	return false
}

var pictureExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".webp": true, ".tif": true, ".tiff": true}

// mediaDocument is a media lesson as a printable document, after
// OpenTeacher's htmlGenerator/media: a row per item with its medium (a
// picture shown small, anything else by its name or address), its name,
// question and answer.
func mediaDocument(data *lesson.LessonData) *qt.QTextDocument {
	doc := qt.NewQTextDocument()
	files := lesson.MediaFiles(data)
	var b strings.Builder
	title := data.List.Title
	if title == "" {
		title = "Media lesson"
	}
	fmt.Fprintf(&b, "<h1>%s</h1><table border='1' cellspacing='0' cellpadding='6' width='100%%'>", html.EscapeString(title))
	b.WriteString("<tr><th>Medium</th><th>Name</th><th>Question</th><th>Answer</th></tr>")
	for i, it := range data.List.Items {
		medium := ""
		if it.Filename != nil {
			name := *it.Filename
			remote := it.Remote != nil && *it.Remote
			if img := files[name]; !remote && img != nil && pictureExts[strings.ToLower(path.Ext(name))] {
				q := qt.QImage_FromDataWithData(img)
				if !q.IsNull() {
					q = q.Scaled3(120, 90, qt.KeepAspectRatio, qt.SmoothTransformation)
					res := fmt.Sprintf("media-%d.png", i)
					doc.AddResource(int(qt.QTextDocument__ImageResource), qt.NewQUrl3(res), q.ToQVariant())
					medium = fmt.Sprintf("<img src='%s'>", res)
				}
			}
			if medium == "" {
				shown := name
				if !remote {
					shown = path.Base(name)
				}
				medium = html.EscapeString(shown)
			}
		}
		fmt.Fprintf(&b, "<tr><td>%s</td><td><b>%s</b></td><td><i>%s</i></td><td>%s</td></tr>", medium,
			html.EscapeString(it.Name), html.EscapeString(strings.Join(it.Questions, ", ")),
			html.EscapeString(strings.Join(it.Answers, ", ")))
	}
	b.WriteString("</table>")
	doc.SetHtml(b.String())
	return doc
}
