package recentlyopened

import (
	"fmt"
	"reflect"
	"testing"
)

type mapSettings map[string]interface{}

func (m mapSettings) GetSettingWithDefault(k string, d interface{}) interface{} {
	if v, ok := m[k]; ok {
		return v
	}
	return d
}
func (m mapSettings) SetSetting(k string, v interface{}) error { m[k] = v; return nil }

func TestRecentlyOpened(t *testing.T) {
	s := mapSettings{}
	Add(s, "a.otwd")
	Add(s, "b.csv")
	Add(s, "a.otwd") // again: moves to the top, once
	if got := List(s); !reflect.DeepEqual(got, []string{"a.otwd", "b.csv"}) {
		t.Errorf("list %q", got)
	}
	for i := 0; i < 20; i++ {
		Add(s, fmt.Sprintf("%d.ot", i))
	}
	if got := List(s); len(got) != Max || got[0] != "19.ot" {
		t.Errorf("long list %q", got)
	}
	// as loaded from the JSON settings file
	s[SettingKey] = []interface{}{"x.ot", 3, "y.ot"}
	if got := List(s); !reflect.DeepEqual(got, []string{"x.ot", "y.ot"}) {
		t.Errorf("from JSON %q", got)
	}
	Clear(s)
	if len(List(s)) != 0 {
		t.Error("not cleared")
	}
}
