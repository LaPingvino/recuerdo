package openteacherauthors

import "testing"

func TestByRole(t *testing.T) {
	roles, names := ByRole()
	if len(roles) != 6 || roles[0] != "Core developer" {
		t.Errorf("roles %q", roles)
	}
	if got := names["Core developer"]; len(got) != 3 || got[2] != "Marten de Vries" {
		t.Errorf("core developers %q", got)
	}
}
