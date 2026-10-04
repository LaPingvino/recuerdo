package datatypeicons

import (
	"bytes"
	"testing"
)

func TestIcon(t *testing.T) {
	for _, typ := range []string{"words", "topo", "media"} {
		if b := Icon(typ); !bytes.HasPrefix(b, []byte("\x89PNG")) {
			t.Errorf("%s: no PNG icon", typ)
		}
	}
	if Icon("cooking") != nil {
		t.Error("an icon for an unknown lesson type")
	}
}
