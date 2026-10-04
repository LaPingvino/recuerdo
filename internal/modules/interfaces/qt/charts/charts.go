// Package charts draws a lesson's results: the percentage of each test as
// a bar (OpenTeacher's percentNotesViewer) and the latest test's answers
// over time (OpenTeacher's progressViewer).
package charts

import (
	"fmt"

	"github.com/LaPingvino/recuerdo/internal/lesson"
	qt "github.com/mappu/miqt/qt6"
)

// Percentages is the percentage of right answers in each test (tests
// without answers left out).
func Percentages(tests []lesson.Test) []int {
	var out []int
	for _, t := range tests {
		if len(t.Results) == 0 {
			continue
		}
		right := 0
		for _, r := range t.Results {
			if r.Result == "right" {
				right++
			}
		}
		out = append(out, (right*100+len(t.Results)/2)/len(t.Results))
	}
	return out
}

// Bar is a bar of the grades chart, in widget coordinates.
type Bar struct {
	X, Y, W, H float64
	Label      string
}

const barWidth, barSpacing = 36.0, 12.0

// Bars lays out a bar per percentage in a chart of the given height,
// showing the last ones that fit in width.
func Bars(percentages []int, width, height float64) []Bar {
	fit := int((width + barSpacing) / (barWidth + barSpacing))
	if len(percentages) > fit && fit > 0 {
		percentages = percentages[len(percentages)-fit:]
	}
	var bars []Bar
	for i, p := range percentages {
		h := float64(p) / 100 * (height - 20)
		bars = append(bars, Bar{
			X: barSpacing/2 + float64(i)*(barWidth+barSpacing), Y: height - h,
			W: barWidth, H: h, Label: fmt.Sprintf("%d%%", p),
		})
	}
	return bars
}

// Block is an answer on the timeline: its share of the test's time.
type Block struct {
	X, W  float64
	Right bool
}

// Timeline lays out the answers of a test along width by the time each
// took (from the previous answer); answers without a time share the rest
// equally.
func Timeline(t lesson.Test, width float64) []Block {
	n := len(t.Results)
	if n == 0 {
		return nil
	}
	durations := make([]float64, n)
	prev := t.Date
	total := 0.0
	for i, r := range t.Results {
		if r.Time != nil && prev != nil && r.Time.After(*prev) {
			durations[i] = r.Time.Sub(*prev).Seconds()
		}
		if r.Time != nil {
			prev = r.Time
		}
		total += durations[i]
	}
	if total == 0 {
		for i := range durations {
			durations[i] = 1
		}
		total = float64(n)
	}
	var blocks []Block
	x := 0.0
	for i, r := range t.Results {
		w := durations[i] / total * width
		blocks = append(blocks, Block{X: x, W: w, Right: r.Result == "right"})
		x += w
	}
	return blocks
}

// GradesChart shows the percentage of each test as a bar.
type GradesChart struct {
	*qt.QWidget
	percentages []int
}

// NewGradesChart creates the chart.
func NewGradesChart(parent *qt.QWidget) *GradesChart {
	c := &GradesChart{QWidget: qt.NewQWidget(parent)}
	c.SetMinimumHeight(120)
	c.OnPaintEvent(func(super func(*qt.QPaintEvent), e *qt.QPaintEvent) {
		p := qt.NewQPainter2(c.QPaintDevice)
		defer p.End()
		pal := c.Palette()
		highlight := pal.Highlight()
		for _, b := range Bars(c.percentages, float64(c.Width()), float64(c.Height())) {
			p.FillRect(qt.NewQRectF4(b.X, b.Y, b.W, b.H), highlight)
			p.SetPen(pal.Text().Color())
			p.DrawText(qt.NewQPointF3(b.X+4, b.Y-4), b.Label)
		}
	})
	return c
}

// SetTests shows tests' percentages.
func (c *GradesChart) SetTests(tests []lesson.Test) {
	c.percentages = Percentages(tests)
	c.Update()
}

// TimelineChart shows the latest test's answers over time.
type TimelineChart struct {
	*qt.QWidget
	test lesson.Test
}

// NewTimelineChart creates the chart.
func NewTimelineChart(parent *qt.QWidget) *TimelineChart {
	c := &TimelineChart{QWidget: qt.NewQWidget(parent)}
	c.SetMinimumHeight(28)
	c.SetToolTip("Your last session over time: green right, red wrong; wider blocks took longer")
	c.OnPaintEvent(func(super func(*qt.QPaintEvent), e *qt.QPaintEvent) {
		p := qt.NewQPainter2(c.QPaintDevice)
		defer p.End()
		green := qt.NewQBrush3(qt.NewQColor3(76, 175, 80))
		red := qt.NewQBrush3(qt.NewQColor3(229, 115, 115))
		for _, b := range Timeline(c.test, float64(c.Width())) {
			brush := red
			if b.Right {
				brush = green
			}
			p.FillRect(qt.NewQRectF4(b.X, 0, max(b.W-1, 1), float64(c.Height())), brush)
		}
	})
	return c
}

// SetTest shows a test.
func (c *TimelineChart) SetTest(t lesson.Test) {
	c.test = t
	c.Update()
}
