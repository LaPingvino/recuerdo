package notecalculatorchooser

import (
	"reflect"
	"testing"
)

func TestChoose(t *testing.T) {
	if want := []string{"American", "Dutch", "ECTS", "French", "German", "Percents"}; !reflect.DeepEqual(Names(), want) {
		t.Errorf("Names() = %v", Names())
	}
	for _, name := range Names() {
		if got := Choose(name).DisplayName(); got != name {
			t.Errorf("Choose(%q) = %q", name, got)
		}
	}
	if got := Choose("Klingon").DisplayName(); got != Default {
		t.Errorf("unknown notation gave %q, want the default %q", got, Default)
	}
}
