package lesson

// Merge adds other's words and test results to base, as OpenTeacher's
// words merger does. The added words get IDs after base's, and their
// results point to the new IDs; base's title and languages stay.
func Merge(base *WordList, other WordList) {
	next := 0
	for _, it := range base.Items {
		next = max(next, it.ID+1)
	}
	newID := map[int]int{}
	for _, it := range other.Items {
		newID[it.ID] = next
		it.ID = next
		base.Items = append(base.Items, it)
		next++
	}
	for _, t := range other.Tests {
		mapped := Test{Date: t.Date}
		for _, r := range t.Results {
			if id, ok := newID[r.ItemID]; ok {
				r.ItemID = id
				mapped.Results = append(mapped.Results, r)
			}
		}
		base.Tests = append(base.Tests, mapped)
	}
}
