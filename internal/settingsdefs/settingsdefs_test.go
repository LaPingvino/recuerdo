package settingsdefs

import (
	"reflect"
	"testing"
)

type mapStore map[string]interface{}

func (m mapStore) GetSettingWithDefault(key string, def interface{}) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}
func (m mapStore) SetSetting(key string, v interface{}) error { m[key] = v; return nil }

func TestRegistry(t *testing.T) {
	Register(Def{Key: "a", Category: "Practice", Kind: Bool, Default: false})
	Register(Def{Key: "b", Category: "Files", Kind: Action})
	Register(Def{Key: "c", Category: "Practice", Kind: Choice, Choices: []string{"x", "y"}, Default: "x"})
	Register(Def{Key: "a", Category: "Practice", Kind: Bool, Default: true}) // replaced, keeps its place
	var keys []string
	for _, d := range All() {
		keys = append(keys, d.Key)
	}
	// by category (Practice before Files), then in registration order
	if !reflect.DeepEqual(keys, []string{"a", "c", "b"}) || !reflect.DeepEqual(Categories(), []string{"Practice", "Files"}) {
		t.Errorf("keys %v, categories %v", keys, Categories())
	}

	s := mapStore{"c": "nonsense", "d": 1500}
	all := map[string]Def{}
	for _, d := range All() {
		all[d.Key] = d
	}
	if all["a"].Value(s) != true || all["c"].Value(s) != "x" {
		t.Errorf("defaults: %v %v", all["a"].Value(s), all["c"].Value(s))
	}
	s["c"] = "y"
	if all["c"].Value(s) != "y" {
		t.Error("stored choice")
	}
	sec := Def{Key: "d", Kind: Seconds, Default: 1000.0}
	if sec.Value(s) != 1500.0 {
		t.Errorf("seconds from an int: %v", sec.Value(s))
	}
}
