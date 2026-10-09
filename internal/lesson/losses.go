package lesson

import (
	"fmt"
	"path/filepath"
	"strings"
)

// keeps is what a lesson format stores besides the words themselves.
type keeps struct{ tests, comments, title, languages bool }

// formatKeeps are the formats Recuerdo saves lessons in that leave
// something out; those not listed (.otwd, .json, .kvtml, .ottp, .otmd)
// keep everything.
var formatKeeps = map[string]keeps{
	".csv":  {comments: true, languages: true},
	".txt":  {title: true, languages: true},
	".ot":   {title: true, languages: true},
	".wrts": {title: true, languages: true},
	".t2k":  {tests: true, comments: true},
	".slk":  {comments: true, title: true},
}

// Losses is what saving data to path would leave out, for the user to
// hear about before it happens: "the results of 3 tests", "the comments".
func Losses(data *LessonData, path string) []string {
	k, ok := formatKeeps[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return nil
	}
	var out []string
	if n := len(data.List.Tests); n > 0 && !k.tests {
		if n == 1 {
			out = append(out, "the results of 1 test")
		} else {
			out = append(out, fmt.Sprintf("the results of %d tests", n))
		}
	}
	if !k.comments {
		for _, it := range data.List.Items {
			if strings.TrimSpace(it.Comment) != "" {
				out = append(out, "the comments")
				break
			}
		}
	}
	// a title is lost only when it is not the file name
	if !k.title && data.List.Title != "" && data.List.Title != titleFromPath(path) {
		out = append(out, fmt.Sprintf("the title %q", data.List.Title))
	}
	if !k.languages && (data.List.QuestionLanguage != "" || data.List.AnswerLanguage != "") {
		out = append(out, "the languages")
	}
	return out
}
