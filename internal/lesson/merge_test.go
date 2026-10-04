package lesson

import "testing"

func TestMerge(t *testing.T) {
	base := WordList{Title: "A", Items: []WordItem{{ID: 0, Questions: []string{"een"}}, {ID: 5, Questions: []string{"twee"}}}}
	other := WordList{Title: "B",
		Items: []WordItem{{ID: 0, Questions: []string{"drie"}}, {ID: 1, Questions: []string{"vier"}}},
		Tests: []Test{{Results: []TestResult{{ItemID: 1, Result: "wrong"}, {ItemID: 0, Result: "right"}}}},
	}
	Merge(&base, other)
	if base.Title != "A" || len(base.Items) != 4 || base.Items[2].ID != 6 || base.Items[3].ID != 7 {
		t.Fatalf("merged items %+v", base.Items)
	}
	if len(base.Tests) != 1 || base.Tests[0].Results[0].ItemID != 7 || base.Tests[0].Results[1].ItemID != 6 {
		t.Errorf("results not mapped to the new IDs: %+v", base.Tests)
	}
	if other.Items[0].ID != 0 {
		t.Error("the other list was changed")
	}
}
