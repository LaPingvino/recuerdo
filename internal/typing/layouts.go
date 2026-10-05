package typing

// Key is a key of an on-screen keyboard: its label (the unshifted
// character, or a name such as "Shift"), where it is (in key widths from
// the left of its row) and which finger types it: 1-4 the left hand
// (little finger to index), 7-10 the right (index to little finger), 5
// the thumbs (the space bar).
type Key struct {
	Label  string  `json:"label"`
	X      float64 `json:"x"`
	Width  float64 `json:"width"`
	Finger int     `json:"finger"`
}

// Layout is a keyboard layout: five rows (numbers, top, home, bottom,
// space).
type Layout struct {
	ID   string  `json:"id"`
	Name string  `json:"name"` // shown with i18n.T (see the list at the end)
	Rows [][]Key `json:"rows"`
}

// geometry is OpenTeacher's keyboard (typingTutor/keyboard SIZE_MAP):
// per row (x, width, finger) of each key, matched by position with the
// layouts' labels.
var geometry = [][][3]float64{
	{{0, 1, 1}, {1, 1, 1}, {2, 1, 2}, {3, 1, 3}, {4, 1, 4}, {5, 1, 4}, {6, 1, 7}, {7, 1, 7}, {8, 1, 8}, {9, 1, 9}, {10, 1, 10}, {11, 1, 10}, {12, 1, 10}, {13, 2, 10}},
	{{0, 1.5, 1}, {1.5, 1, 1}, {2.5, 1, 2}, {3.5, 1, 3}, {4.5, 1, 4}, {5.5, 1, 4}, {6.5, 1, 7}, {7.5, 1, 7}, {8.5, 1, 8}, {9.5, 1, 9}, {10.5, 1, 10}, {11.5, 1, 10}, {12.5, 1, 10}, {13.5, 1.5, 10}},
	{{0, 2, 1}, {2, 1, 1}, {3, 1, 2}, {4, 1, 3}, {5, 1, 4}, {6, 1, 4}, {7, 1, 7}, {8, 1, 7}, {9, 1, 8}, {10, 1, 9}, {11, 1, 10}, {12, 1, 10}, {13, 1, 10}, {14, 1, 10}},
	{{0, 1.5, 1}, {1.5, 1, 1}, {2.5, 1, 1}, {3.5, 1, 2}, {4.5, 1, 3}, {5.5, 1, 4}, {6.5, 1, 4}, {7.5, 1, 7}, {8.5, 1, 7}, {9.5, 1, 8}, {10.5, 1, 9}, {11.5, 1, 10}, {12.5, 2.5, 10}},
	{{1.5, 12, 5}},
}

// Key names (not characters); Enter's lower half is "".
const (
	Backspace = "Backspace"
	Tab       = "Tab"
	Enter     = "Enter"
	CapsLock  = "Caps Lock"
	Shift     = "Shift"
	Space     = "Space"
)

func layout(id, name string, labels [4][]string) Layout {
	l := Layout{ID: id, Name: name}
	rows := [5][]string{labels[0], labels[1], labels[2], labels[3], {Space}}
	for r, row := range rows {
		var keys []Key
		for i, label := range row {
			g := geometry[r][i]
			keys = append(keys, Key{Label: label, X: g[0], Width: g[1], Finger: int(g[2])})
		}
		l.Rows = append(l.Rows, keys)
	}
	return l
}

// Layouts are the keyboard layouts (OpenTeacher's six; its AZERTY
// layouts had "x" twice where "w" belongs).
var Layouts = []Layout{
	layout("qwerty", "QWERTY", [4][]string{
		{"`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", Backspace},
		{Tab, "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", Enter},
		{CapsLock, "a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "'", "\\", ""},
		{Shift, "\\", "z", "x", "c", "v", "b", "n", "m", ",", ".", "/", Shift},
	}),
	layout("azerty-be", "Belgian AZERTY", [4][]string{
		{"²", "&", "é", "\"", "'", "(", "§", "è", "!", "ç", "à", ")", "-", Backspace},
		{Tab, "a", "z", "e", "r", "t", "y", "u", "i", "o", "p", "^", "$", Enter},
		{CapsLock, "q", "s", "d", "f", "g", "h", "j", "k", "l", "m", "ù", "µ", ""},
		{Shift, "<", "w", "x", "c", "v", "b", "n", ",", ";", ":", "=", Shift},
	}),
	layout("azerty-fr", "French AZERTY", [4][]string{
		{"²", "&", "é", "\"", "'", "(", "-", "è", "_", "ç", "à", ")", "=", Backspace},
		{Tab, "a", "z", "e", "r", "t", "y", "u", "i", "o", "p", "^", "$", Enter},
		{CapsLock, "q", "s", "d", "f", "g", "h", "j", "k", "l", "m", "ù", "*", ""},
		{Shift, "<", "w", "x", "c", "v", "b", "n", ",", ";", ":", "!", Shift},
	}),
	layout("colemak", "Colemak", [4][]string{
		{"`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", Backspace},
		{Tab, "q", "w", "f", "p", "g", "j", "l", "u", "y", ";", "[", "]", Enter},
		{Backspace, "a", "r", "s", "t", "d", "h", "n", "e", "i", "o", "'", "\\", ""},
		{Shift, "\\", "z", "x", "c", "v", "b", "k", "m", ",", ".", "/", Shift},
	}),
	layout("dvorak", "Dvorak", [4][]string{
		{"`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "[", "]", Backspace},
		{Tab, "'", ",", ".", "p", "y", "f", "g", "c", "r", "l", "/", "=", Enter},
		{CapsLock, "a", "o", "e", "u", "i", "d", "h", "t", "n", "s", "-", "\\", ""},
		{Shift, "\\", ";", "q", "j", "k", "x", "b", "m", "w", "v", "z", Shift},
	}),
	layout("qwertz", "QWERTZ", [4][]string{
		{"^", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "ß", "´", Backspace},
		{Tab, "q", "w", "e", "r", "t", "z", "u", "i", "o", "p", "ü", "+", Enter},
		{CapsLock, "a", "s", "d", "f", "g", "h", "j", "k", "l", "ö", "ä", "#", ""},
		{Shift, "<", "y", "x", "c", "v", "b", "n", "m", ",", ".", "-", Shift},
	}),
}

// LayoutByID finds a layout (QWERTY when the ID is unknown).
func LayoutByID(id string) Layout {
	for _, l := range Layouts {
		if l.ID == id {
			return l
		}
	}
	return Layouts[0]
}

// KeyFor is the key that types character c (lower case), or false.
func (l Layout) KeyFor(c string) (Key, bool) {
	if c == " " {
		return l.Rows[4][0], true
	}
	for _, row := range l.Rows {
		for _, k := range row {
			if k.Label == c && len([]rune(c)) == 1 {
				return k, true
			}
		}
	}
	return Key{}, false
}

// typable is whether every character of word has a key on the layout.
func (l Layout) typable(word string) bool {
	for _, r := range word {
		if _, ok := l.KeyFor(string(r)); !ok {
			return false
		}
	}
	return true
}

// The names shown translated (scripts/extract_strings.py finds them
// here): the layouts and the languages with words.
// i18n:values
var (
	_ = "QWERTY"
	_ = "Belgian AZERTY"
	_ = "French AZERTY"
	_ = "Colemak"
	_ = "Dvorak"
	_ = "QWERTZ"
	_ = "English"
	_ = "Dutch"
	_ = "German"
	_ = "French"
	_ = "Spanish"
)
