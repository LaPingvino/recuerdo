package percentscalculator

import (
	"testing"

	lessontypes "github.com/LaPingvino/recuerdo/internal/modules/logic/lessonTypes"
)

func test(right ...bool) lessontypes.Test {
	var t lessontypes.Test
	for i, r := range right {
		t.Results = append(t.Results, lessontypes.Result{ItemID: i, Right: r})
	}
	return t
}

func TestPercents(t *testing.T) {
	for _, c := range []struct {
		test lessontypes.Test
		want int
	}{
		{test(), 0},
		{test(true), 100},
		{test(true, false), 50},
		{test(true, true, false), 67}, // 66.67 rounds up
		{test(true, false, false, false, false, false, false, false), 13}, // 12.5 rounds half up
	} {
		if got := Percents(c.test); got != c.want {
			t.Errorf("Percents(%v) = %d, want %d", c.test.Results, got, c.want)
		}
	}
	if got := AveragePercents([]lessontypes.Test{test(true), test(true, false), test(false)}); got != 50 {
		t.Errorf("average of 100, 50, 0 = %d", got)
	}
	if AveragePercents(nil) != 0 {
		t.Error("average of no tests")
	}
	m := NewPercentsCalculatorModule()
	if m.CalculatePercents(test(true, false)) != 50 || m.CalculateAveragePercents([]lessontypes.Test{test(true)}) != 100 {
		t.Error("module methods")
	}
}
