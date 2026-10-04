package charts

import (
	"reflect"
	"testing"
	"time"

	"github.com/LaPingvino/recuerdo/internal/lesson"
)

func test(results ...string) lesson.Test {
	var t lesson.Test
	for i, r := range results {
		t.Results = append(t.Results, lesson.TestResult{ItemID: i, Result: r})
	}
	return t
}

func TestPercentages(t *testing.T) {
	got := Percentages([]lesson.Test{test("right", "wrong"), {}, test("right", "right", "wrong")})
	if !reflect.DeepEqual(got, []int{50, 67}) {
		t.Errorf("%v", got)
	}
}

func TestBars(t *testing.T) {
	bars := Bars([]int{100, 50}, 400, 120)
	if len(bars) != 2 || bars[0].H != 100 || bars[1].H != 50 || bars[1].Y != 70 || bars[0].Label != "100%" {
		t.Errorf("%+v", bars)
	}
	if bars[1].X <= bars[0].X+bars[0].W {
		t.Error("bars overlap")
	}
	// only the latest that fit
	many := Bars([]int{10, 20, 30, 40, 50}, 100, 120)
	if len(many) != 2 || many[1].Label != "50%" {
		t.Errorf("narrow: %+v", many)
	}
}

func TestTimeline(t *testing.T) {
	start := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	at := func(s int) *time.Time { x := start.Add(time.Duration(s) * time.Second); return &x }
	tt := lesson.Test{Date: &start, Results: []lesson.TestResult{
		{Result: "right", Time: at(1)}, {Result: "wrong", Time: at(4)},
	}}
	b := Timeline(tt, 400)
	if len(b) != 2 || b[0].W != 100 || b[1].W != 300 || b[1].X != 100 || !b[0].Right || b[1].Right {
		t.Errorf("%+v", b)
	}
	// without times: equal blocks
	if b := Timeline(test("right", "wrong"), 100); b[0].W != 50 || b[1].W != 50 {
		t.Errorf("no times: %+v", b)
	}
}
