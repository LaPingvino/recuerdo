// Package typing is Recuerdo's touch typing course, after OpenTeacher's
// typing tutor: letter exercises row by row (home row first), then words,
// with a speed to reach and no mistakes allowed; an on-screen keyboard
// shows which finger types each key.
package typing

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaPingvino/recuerdo/internal/i18n"
)

// The course: OpenTeacher's 37 letter exercises, then 20 word levels.
const (
	LetterTarget = 20 // words per minute for the letter exercises
	WordLevels   = 20
	rowLength    = 59 // characters of a letter exercise
)

// LetterExercises are the keys each letter exercise practises: per row
// (home, top, bottom, numbers) the index fingers' keys first, then
// outwards in pairs, then in groups; finally every letter at once.
func LetterExercises(l Layout) [][]string {
	label := func(row, i int) string { return l.Rows[row][i].Label }
	span := func(row, from, to int) []string {
		var keys []string
		for i := from; i < to; i++ {
			keys = append(keys, label(row, i))
		}
		return keys
	}
	var ex [][]string
	var all []string
	for _, row := range []int{2, 1, 3, 0} {
		s := 0
		if row == 3 {
			s = 1 // the ISO key left of the bottom row's letters
		}
		ex = append(ex,
			[]string{label(row, s+4), label(row, s+7)},
			[]string{label(row, s+3), label(row, s+8)},
			[]string{label(row, s+2), label(row, s+9)},
			[]string{label(row, s+1), label(row, s+10)},
			span(row, s+1, s+5), span(row, s+7, s+11), span(row, s+5, s+7), span(row, s+4, s+8),
			span(row, s+1, s+11))
		all = append(all, span(row, s+1, s+11)...)
	}
	return append(ex, all)
}

// Levels is the number of levels of the course.
func Levels(l Layout) int { return len(LetterExercises(l)) + WordLevels }

// TargetSpeed is the words per minute a level asks for: 20 for the
// letters, then rising to 80 over the word levels.
func TargetSpeed(l Layout, level int) int {
	letters := len(LetterExercises(l))
	if level < letters {
		return LetterTarget
	}
	return int(math.Round(LetterTarget + float64(level-letters)/float64(WordLevels-1)*60))
}

// WordsPerMinute is a typing speed: five characters (spaces included)
// make a word.
func WordsPerMinute(text string, seconds float64) int {
	if seconds <= 0 {
		return 0
	}
	return int(math.Round(float64(len([]rune(text))) / 5 / (seconds / 60)))
}

// Exercise makes the text of a level: a letter level's keys shuffled in
// groups of five, or eight words of the word list that can be typed on
// the layout.
func Exercise(l Layout, level int, words []string, rng *rand.Rand) string {
	letters := LetterExercises(l)
	if level < len(letters) {
		keys := letters[level]
		var pool []string
		for len(pool) < 80 {
			pool = append(pool, keys...)
		}
		rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		var b strings.Builder
		for i, k := range pool {
			if i > 0 && i%5 == 0 {
				b.WriteString(" ")
			}
			b.WriteString(k)
		}
		r := []rune(b.String())
		return strings.TrimSpace(string(r[:min(len(r), rowLength)]))
	}
	var typable []string
	for _, w := range words {
		if l.typable(w) {
			typable = append(typable, w)
		}
	}
	if len(typable) == 0 {
		typable = []string{"the"}
	}
	picked := make([]string, 8)
	for i := range picked {
		picked[i] = typable[rng.Intn(len(typable))]
	}
	return strings.Join(picked, " ")
}

// ---- profiles ----

// Statuses of a profile after an exercise.
const (
	Start    = "start"
	Mistakes = "mistakes"
	Slow     = "slow"
	Next     = "next"
	Done     = "done"
)

// Result is a finished exercise.
type Result struct {
	Level    int       `json:"level"`
	Exercise string    `json:"exercise"`
	Seconds  float64   `json:"seconds"`
	Mistakes int       `json:"mistakes"`
	At       time.Time `json:"at"`
}

// Speed is the result's words per minute.
func (r Result) Speed() int { return WordsPerMinute(r.Exercise, r.Seconds) }

// Profile is someone taking the course.
type Profile struct {
	Name     string   `json:"name"`
	Layout   string   `json:"layout"`
	Language string   `json:"language"` // of the words (a code: "nl")
	Level    int      `json:"level"`
	Status   string   `json:"status"`
	Current  string   `json:"exercise"`
	Results  []Result `json:"results"`
}

// KeyboardLayout is the profile's layout.
func (p *Profile) KeyboardLayout() Layout { return LayoutByID(p.Layout) }

// Last is the last result, or false.
func (p *Profile) Last() (Result, bool) {
	if len(p.Results) == 0 {
		return Result{}, false
	}
	return p.Results[len(p.Results)-1], true
}

// Finish records an exercise: with no mistakes at the target speed the
// next level comes; otherwise a new exercise of the same level.
func (p *Profile) Finish(seconds float64, mistakes int, rng *rand.Rand) {
	l := p.KeyboardLayout()
	r := Result{Level: p.Level, Exercise: p.Current, Seconds: seconds, Mistakes: mistakes, At: time.Now()}
	p.Results = append(p.Results, r)
	switch {
	case mistakes > 0:
		p.Status = Mistakes
	case r.Speed() < TargetSpeed(l, p.Level):
		p.Status = Slow
	case p.Level+1 >= Levels(l):
		p.Status = Done // the last level again, as long as one likes
	default:
		p.Level++
		p.Status = Next
	}
	p.Current = Exercise(l, p.Level, Words(p.Language), rng)
}

// Instruction is what the tutor says before the next exercise.
func (p *Profile) Instruction(rng *rand.Rand) string {
	l := p.KeyboardLayout()
	pick := func(texts ...string) string { return texts[rng.Intn(len(texts))] }
	last, _ := p.Last()
	letters := p.Level < len(LetterExercises(l))
	var parts []string
	if len(p.Results) == 1 && p.Status != Start {
		parts = append(parts, i18n.T("Congratulations, you finished your first exercise!"))
	}
	switch p.Status {
	case Start:
		home := l.Rows[2]
		keys := []string{home[1].Label, home[2].Label, home[3].Label, home[4].Label, i18n.T("space"), i18n.T("space"),
			home[7].Label, home[8].Label, home[9].Label, home[10].Label}
		for i, k := range keys {
			keys[i] = "'" + k + "'"
		}
		return i18n.T("Welcome to the typing course. You'll improve your typing by doing simple exercises; between them, you get instructions. Let's get started:") +
			"\n\n" + i18n.Tf("First place your fingers on the home row: from left to right, on the keys %s. Keep them there when you're not typing another key. When your fingers are in position, start the first exercise. Work for accuracy at first, not speed.", strings.Join(keys, ", "))
	case Done:
		return i18n.T("Congratulations, you finished this typing course! If you want to continue, you can, but this is the end of the instructions. You did a great job!")
	case Mistakes:
		parts = append(parts, pick(
			i18n.Tf("Mistakes: %d. Please keep trying until you can do it without any.", last.Mistakes),
			i18n.Tf("Too bad: %d mistakes. Keep practising to get better!", last.Mistakes)))
		if letters && last.Speed() >= 40 {
			parts = append(parts, i18n.T("To achieve that, you might try slowing down a bit."))
		}
	case Slow:
		parts = append(parts, i18n.T("You made zero mistakes. Now try to improve your typing speed a bit."))
	case Next:
		if letters {
			parts = append(parts, pick(
				i18n.T("You made zero mistakes and are typing fast enough, so you can continue practising some new letter combinations. Keep up the good work!"),
				i18n.T("You did it flawlessly and fast! Continue practising some new letters combinations to get even better!")))
			if p.Level == 9 {
				parts = append(parts, i18n.T("You're now going to learn letters that aren't on the home row. To see which fingers you need to use, see the keyboard image on your screen. When you're not using a finger to type a letter, put it back on the home row directly."))
			}
			if p.Level%9 == 6 {
				parts = append(parts, i18n.T("The keys you're going to practise now are typed by the left and right index finger and further away from those fingers than the other keys we practised on the current row. Make sure you return your finger to its position on the home row when you're typing another letter."))
			}
		} else {
			parts = append(parts, pick(
				i18n.T("You made zero mistakes and are typing fast enough, so you can continue practising with some new words. Keep up the good work!"),
				i18n.T("You did it flawlessly and fast! Continue practising with new words to get even better!")))
		}
	}
	return strings.Join(parts, "\n\n")
}

// ---- the profiles of this computer ----

// Errors for a new profile's name.
var (
	ErrNameEmpty = errors.New("a name is needed")
	ErrNameTaken = errors.New("that name is taken")
)

// Profiles are the course's profiles on this computer, kept in a file.
type Profiles struct {
	path string
	List []*Profile `json:"profiles"`
}

// DefaultPath is where the profiles are kept.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "recuerdo", "typing.json")
}

// Load reads the profiles (none when the file is not there yet).
func Load(path string) (*Profiles, error) {
	ps := &Profiles{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ps, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// Save writes the profiles (after every exercise: nothing is lost when
// Recuerdo stops).
func (ps *Profiles) Save() error {
	data, err := json.MarshalIndent(ps, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ps.path), 0o755); err != nil {
		return err
	}
	tmp := ps.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, ps.path)
}

// Add makes a profile with its first exercise.
func (ps *Profiles) Add(name, layoutID, language string, rng *rand.Rand) (*Profile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameEmpty
	}
	if ps.Get(name) != nil {
		return nil, ErrNameTaken
	}
	p := &Profile{Name: name, Layout: LayoutByID(layoutID).ID, Language: language, Status: Start}
	p.Current = Exercise(p.KeyboardLayout(), 0, Words(language), rng)
	ps.List = append(ps.List, p)
	sort.Slice(ps.List, func(i, j int) bool { return strings.ToLower(ps.List[i].Name) < strings.ToLower(ps.List[j].Name) })
	return p, nil
}

// Get is a profile by name, or nil.
func (ps *Profiles) Get(name string) *Profile {
	for _, p := range ps.List {
		if strings.EqualFold(p.Name, strings.TrimSpace(name)) {
			return p
		}
	}
	return nil
}

// Remove deletes a profile.
func (ps *Profiles) Remove(name string) {
	for i, p := range ps.List {
		if strings.EqualFold(p.Name, name) {
			ps.List = append(ps.List[:i], ps.List[i+1:]...)
			return
		}
	}
}
