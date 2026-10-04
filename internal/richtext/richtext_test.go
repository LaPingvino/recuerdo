package richtext

import "testing"

func TestSanitize(t *testing.T) {
	for in, want := range map[string]string{
		"dog":                                     "dog",
		"a < b & c":                               "a &lt; b &amp; c",
		"H<sub>2</sub>O":                          "H<sub>2</sub>O",
		"<ruby>漢<rt>かん</rt></ruby>":               "<ruby>漢<rt>かん</rt></ruby>",
		`<b onclick="x()">bold</b>`:               "<b>bold</b>",
		`<script>alert(1)</script>hi`:             "hi",
		`<img src="javascript:alert(1)" alt="x">`: `<img alt="x">`,
		`<img src="data:image/png;base64,AAAA" alt="dog">`:         `<img src="data:image/png;base64,AAAA" alt="dog">`,
		`<a href="javascript:x">link</a>`:                          `<a target="_blank" rel="noopener noreferrer">link</a>`,
		`<a href="https://openteacher.org">site</a>`:               `<a href="https://openteacher.org" target="_blank" rel="noopener noreferrer">site</a>`,
		`<audio src="bark.mp3"></audio>`:                           `<audio src="bark.mp3" controls></audio>`,
		`<div class="x"><span lang="ja" style="c">日本</span></div>`: `<span lang="ja">日本</span>`,
		`<iframe src="https://evil"></iframe>text`:                 "text",
	} {
		if got := Sanitize(in); got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlain(t *testing.T) {
	for in, want := range map[string]string{
		"dog":            "dog",
		"a < b":          "a < b",
		"H<sub>2</sub>O": "H2O",
		"<ruby>漢<rp>(</rp><rt>かん</rt><rp>)</rp>字<rt>じ</rt></ruby>": "漢字",
		"x<sup>2</sup> + 1":             "x2 + 1",
		`<img src="dog.png" alt="dog">`: "dog",
		"one<br>two":                    "one two",
	} {
		if got := Plain(in); got != want {
			t.Errorf("Plain(%q) = %q, want %q", in, got, want)
		}
	}
}
