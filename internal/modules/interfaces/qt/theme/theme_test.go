package theme

import (
	"os"
	"testing"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessons/words"
)

func TestMain(m *testing.M) {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")
	qt.NewQApplication([]string{"theme-test"})
	os.Exit(m.Run())
}

func TestThemes(t *testing.T) {
	window := func() int { return qt.QGuiApplication_Palette().Window().Color().Lightness() }
	before := window()
	Apply(Dark)
	if window() > 80 {
		t.Errorf("dark: window lightness %d", window())
	}
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		l := &lesson.Lesson{DataType: "words", Data: lesson.LessonData{List: lesson.WordList{Title: "Animals",
			QuestionLanguage: "Dutch", AnswerLanguage: "English", Items: []lesson.WordItem{
				{ID: 0, Questions: []string{"hond"}, Answers: []string{"dog"}},
				{ID: 1, Questions: []string{"the area of a circle"}, Answers: []string{"$\\pi r^2$"}},
			}}}}
		w := words.NewWordsLessonWidget(l, nil)
		w.Resize(900, 420)
		w.Show()
		qt.QCoreApplication_ProcessEvents()
		w.Grab().Save(dir + "/dark.png")
		w.Hide() // not Close: that asks whether to save the lesson
	}
	Apply(Light)
	if window() < 180 {
		t.Errorf("light: window lightness %d", window())
	}
	Apply(System)
	if window() != before {
		t.Errorf("system: %d, was %d", window(), before)
	}
}
