package cli

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/testserver"
)

func TestMakeAdmin(t *testing.T) {
	store, err := testserver.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var out bytes.Buffer
	c := &ctx{in: bufio.NewReader(strings.NewReader("")), out: &out, err: &out}
	if err := makeAdmin(c, store, "beheerder"); err != nil {
		t.Fatal(err)
	}
	_, password, ok := strings.Cut(strings.TrimSpace(out.String()), "(shown only now): ")
	if !ok || len(password) != 14 {
		t.Fatalf("output %q", out.String())
	}
	if _, u, err := store.Login("beheerder", password); err != nil || u.Role != testserver.Admin {
		t.Errorf("login with the printed password: %v %+v", err, u)
	}
	out.Reset()
	makeAdmin(c, store, "beheerder") // again: kept, not replaced
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("second time: %q", out.String())
	}
	if a, b := randomPassword(14), randomPassword(14); a == b || strings.ContainsAny(a, "0O1lI") {
		t.Errorf("passwords %q %q", a, b)
	}
}
