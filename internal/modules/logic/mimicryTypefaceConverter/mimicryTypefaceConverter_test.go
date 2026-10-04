package mimicrytypefaceconverter

import "testing"

// The cases of OpenTeacher's mimicryTypefaceConverterTest, and a few more.
func TestConvert(t *testing.T) {
	for _, c := range []struct{ font, in, want string }{
		{"d m 9038m4edreolwi4i832 disfdhoiw", "tést", "tést"},
		{"Greek", "a b", "α β"},
		{"Symbol", "a b D", "α β Δ"},
		{"TekniaGreek", "logoV", "λογοῃ"},
		{"Arial", "abc", "abc"},
		{"symbol", "1 + a = é", "1 + α = é"},
	} {
		if got := Convert(c.font, c.in); got != c.want {
			t.Errorf("Convert(%q, %q) = %q, want %q", c.font, c.in, got, c.want)
		}
	}
}
