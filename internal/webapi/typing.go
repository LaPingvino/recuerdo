package webapi

import (
	"math/rand"
	"time"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/langcode"
	"github.com/LaPingvino/recuerdo/internal/typing"
)

// The typing course for the web version: the page keeps the profiles (in
// the browser's storage) and hands one to these functions, which give it
// back updated (internal/typing does the course).

var typingRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// TypingLayout is a keyboard layout with its translated name.
type TypingLayout struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Rows [][]typing.Key `json:"rows"`
}

// TypingChoices are the layouts and the languages with words.
func TypingChoices() map[string]any {
	var layouts []TypingLayout
	for _, l := range typing.Layouts {
		layouts = append(layouts, TypingLayout{ID: l.ID, Name: i18n.T(l.Name), Rows: l.Rows})
	}
	var langs [][2]string
	for _, code := range typing.WordLanguages() {
		langs = append(langs, [2]string{code, i18n.T(langcode.Name(code))})
	}
	return map[string]any{"layouts": layouts, "languages": langs}
}

// TypingState is a profile with what the page shows of it.
type TypingState struct {
	Profile     typing.Profile `json:"profile"`
	Instruction string         `json:"instruction"`
	Level       int            `json:"level"` // from 1
	Levels      int            `json:"levels"`
	Target      int            `json:"target"`
	Speed       int            `json:"speed"`
	Mistakes    int            `json:"mistakes"`
	Done        bool           `json:"done"` // at least one exercise finished
}

func typingState(p typing.Profile) TypingState {
	l := p.KeyboardLayout()
	st := TypingState{Profile: p, Instruction: p.Instruction(typingRand), Level: p.Level + 1,
		Levels: typing.Levels(l), Target: typing.TargetSpeed(l, p.Level)}
	if last, ok := p.Last(); ok {
		st.Speed, st.Mistakes, st.Done = last.Speed(), last.Mistakes, true
	}
	return st
}

// TypingNew makes a profile with its first exercise.
func TypingNew(name, layoutID, language string) (TypingState, error) {
	ps := &typing.Profiles{} // the page keeps the profiles; this only makes one
	p, err := ps.Add(name, layoutID, language, typingRand)
	if err != nil {
		return TypingState{}, err
	}
	return typingState(*p), nil
}

// TypingShow is a profile's state.
func TypingShow(p typing.Profile) TypingState { return typingState(p) }

// TypingFinish records a finished exercise.
func TypingFinish(p typing.Profile, seconds float64, mistakes int) TypingState {
	p.Finish(seconds, mistakes, typingRand)
	return typingState(p)
}
