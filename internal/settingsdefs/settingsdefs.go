// Package settingsdefs lists the settings the user can change: modules
// register theirs (as OpenTeacher's modules registered settings with the
// settings module) and the settings dialog shows them, grouped by
// category.
package settingsdefs

import (
	"sort"
	"sync"
)

// Kinds of settings.
const (
	Bool    = "bool"    // a checkbox
	Choice  = "choice"  // one of Choices
	Seconds = "seconds" // a duration, stored in milliseconds
	Action  = "action"  // a button that runs Run
)

// Store is where settings are kept.
type Store interface {
	GetSettingWithDefault(key string, defaultValue interface{}) interface{}
	SetSetting(key string, value interface{}) error
}

// Def is a setting the user can change.
type Def struct {
	Key      string
	Category string // the dialog's tab
	Name     string
	Help     string
	Kind     string
	Choices  []string    // Choice: the values stored
	Labels   []string    // Choice: what the user sees (default: Choices)
	Default  interface{} // bool, string (Choice) or milliseconds (Seconds)
	Min, Max float64     // Seconds
	Run      func(Store) // Action
	order    int
}

var (
	mu   sync.Mutex
	defs = map[string]Def{}
)

// Register adds a setting (replacing one with the same key).
func Register(d Def) {
	mu.Lock()
	defer mu.Unlock()
	if old, ok := defs[d.Key]; ok {
		d.order = old.order
	} else {
		d.order = len(defs)
	}
	defs[d.Key] = d
}

// CategoryOrder is the order of the dialog's tabs; other categories come
// after these, alphabetically.
var CategoryOrder = []string{"Practice", "Results", "Files", "Interface"}

func rank(category string) int {
	for i, c := range CategoryOrder {
		if c == category {
			return i
		}
	}
	return len(CategoryOrder)
}

// All are the registered settings by category (CategoryOrder), in the
// order they were registered within a category.
func All() []Def {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Def, 0, len(defs))
	for _, d := range defs {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if rank(a.Category) != rank(b.Category) {
			return rank(a.Category) < rank(b.Category)
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.order < b.order
	})
	return out
}

// Categories are the categories of the registered settings, in order.
func Categories() []string {
	var cats []string
	seen := map[string]bool{}
	for _, d := range All() {
		if !seen[d.Category] {
			seen[d.Category] = true
			cats = append(cats, d.Category)
		}
	}
	return cats
}

// Value is d's current value in s, as the kind's Go type (bool, string,
// float64 milliseconds); the default when unset or of the wrong type.
func (d Def) Value(s Store) interface{} {
	v := s.GetSettingWithDefault(d.Key, d.Default)
	switch d.Kind {
	case Bool:
		if b, ok := v.(bool); ok {
			return b
		}
	case Choice:
		if str, ok := v.(string); ok {
			for _, c := range d.Choices {
				if c == str {
					return str
				}
			}
		}
	case Seconds:
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case int64:
			return float64(n)
		}
	}
	return d.Default
}
