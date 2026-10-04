package lesson

import "time"

// otTest is a test (a practice session) as OpenTeacher's .otwd and .ottp
// files store it: a result per answer, with when it was "active".
type otTest struct {
	Results  []otResult `json:"results"`
	Finished bool       `json:"finished"`
}

type otResult struct {
	ItemID int       `json:"itemId"`
	Result string    `json:"result"`
	Active *otActive `json:"active,omitempty"`
}

type otActive struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end"`
}

// otTime is how OpenTeacher writes times in its files.
const otTime = "2006-01-02T15:04:05.999999"

// fromOTTests converts tests as read from a file.
func fromOTTests(tests []otTest) []Test {
	var out []Test
	for _, t := range tests {
		var test Test
		for i, r := range t.Results {
			res := TestResult{ItemID: r.ItemID, Result: r.Result}
			if r.Active != nil {
				if end, err := time.Parse(otTime, r.Active.End); err == nil {
					res.Time = &end
				}
				if start, err := time.Parse(otTime, r.Active.Start); err == nil && i == 0 {
					test.Date = &start
				}
			}
			test.Results = append(test.Results, res)
		}
		out = append(out, test)
	}
	return out
}

// toOTTests converts tests for writing: each answer is active from the
// previous answer (or the test's start) to when it was given.
func toOTTests(tests []Test) []otTest {
	out := []otTest{}
	for _, t := range tests {
		ot := otTest{Finished: true, Results: []otResult{}}
		prev := t.Date
		for _, r := range t.Results {
			res := otResult{ItemID: r.ItemID, Result: r.Result}
			if r.Time != nil {
				res.Active = &otActive{End: r.Time.Format(otTime)}
				if prev != nil {
					res.Active.Start = prev.Format(otTime)
				}
				prev = r.Time
			}
			ot.Results = append(ot.Results, res)
		}
		out = append(out, ot)
	}
	return out
}
