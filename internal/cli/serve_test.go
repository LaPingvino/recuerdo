package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWebHandler(t *testing.T) {
	dir := t.TempDir()
	if _, err := webHandler(dir); err == nil || !strings.Contains(err.Error(), "build-web.sh") {
		t.Errorf("a folder without the web version: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "recuerdo.wasm"), []byte("\x00asm"), 0o644)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<title>Recuerdo</title>"), 0o644)
	h, err := webHandler(dir)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()
	for path, want := range map[string]string{"/recuerdo.wasm": "application/wasm", "/": "text/html"} {
		r, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if r.StatusCode != 200 || !strings.HasPrefix(r.Header.Get("Content-Type"), want) {
			t.Errorf("%s: %d %q", path, r.StatusCode, r.Header.Get("Content-Type"))
		}
	}
}
