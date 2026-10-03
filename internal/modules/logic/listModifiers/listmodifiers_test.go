package listmodifiers_test

import (
	"reflect"
	gosort "sort"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/lesson"
	random "github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/random_"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/reverse"
	"github.com/LaPingvino/recuerdo/internal/modules/logic/listModifiers/sort"
)

// the list OpenTeacher's randomTest and sortTest use
var items = []lesson.WordItem{
	{ID: 1, Questions: []string{"b"}},
	{ID: 2, Questions: []string{"c"}},
	{ID: 2, Questions: []string{"a"}},
}

func TestRandom(t *testing.T) {
	got := random.ModifyList([]int{0, 2}, nil)
	gosort.Ints(got)
	if !reflect.DeepEqual(got, []int{0, 2}) {
		t.Errorf("shuffled indexes %v are not a permutation of [0 2]", got)
	}
	swapAll := func(n int, swap func(i, j int)) { swap(0, n-1) }
	if got := random.ModifyList([]int{0, 1, 2}, swapAll); !reflect.DeepEqual(got, []int{2, 1, 0}) {
		t.Errorf("with a fixed shuffle: %v", got)
	}
	in := []int{0, 1, 2}
	random.ModifyList(in, swapAll)
	if !reflect.DeepEqual(in, []int{0, 1, 2}) {
		t.Error("the input was changed")
	}
}

func TestReverse(t *testing.T) {
	if got := reverse.ModifyList([]int{0, 2}); !reflect.DeepEqual(got, []int{2, 0}) {
		t.Errorf("reverse = %v, want [2 0]", got)
	}
}

func TestSort(t *testing.T) {
	if got := sort.ModifyList([]int{0, 2}, items); !reflect.DeepEqual(got, []int{2, 0}) {
		t.Errorf("sort = %v, want [2 0]", got)
	}
	withEmpty := append(items, lesson.WordItem{ID: 4})
	if got := sort.ModifyList([]int{0, 1, 2, 3}, withEmpty); !reflect.DeepEqual(got, []int{3, 2, 0, 1}) {
		t.Errorf("sort with an item without questions = %v", got)
	}
}

func TestModules(t *testing.T) {
	for _, m := range []interface {
		core.Module
		DisplayName() string
	}{random.NewRandomModule(), reverse.NewReverseModule(), sort.NewSortModule()} {
		if m.Type() != "listModifier" || m.Name() == m.DisplayName() {
			t.Errorf("%T: type %q, name %q, display name %q", m, m.Type(), m.Name(), m.DisplayName())
		}
	}
}
