// Package richtext handles words that contain a little HTML: furigana
// (<ruby>), chemical formulas (H<sub>2</sub>O), formatting, pictures and
// sounds. Sanitize keeps only harmless markup for showing them; Plain is
// the text an answer is checked against.
package richtext

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// allowed elements, with the attributes they may keep.
var allowed = map[atom.Atom][]string{
	atom.B: nil, atom.I: nil, atom.U: nil, atom.Em: nil, atom.Strong: nil, atom.Small: nil,
	atom.Sub: nil, atom.Sup: nil, atom.Br: nil, atom.S: nil, atom.Mark: nil,
	atom.Ruby: nil, atom.Rt: nil, atom.Rp: nil,
	atom.Span:  {"lang"},
	atom.Img:   {"src", "alt", "width", "height"},
	atom.Audio: {"src"}, atom.Video: {"src", "width", "height"}, atom.Source: {"src", "type"},
	atom.A: {"href"},
}

// IsRich reports whether s looks like it contains markup.
func IsRich(s string) bool {
	i := strings.IndexByte(s, '<')
	return i >= 0 && i+1 < len(s) && (s[i+1] == '/' || s[i+1] >= 'a' && s[i+1] <= 'z' || s[i+1] >= 'A' && s[i+1] <= 'Z')
}

// safeURL allows embedded pictures and sounds (data:), relative paths and
// web addresses, never javascript: and the like.
func safeURL(u string, link bool) bool {
	l := strings.ToLower(strings.TrimSpace(u))
	switch {
	case strings.HasPrefix(l, "https://"), strings.HasPrefix(l, "http://"):
		return true
	case link:
		return false
	case strings.HasPrefix(l, "data:image/"), strings.HasPrefix(l, "data:audio/"), strings.HasPrefix(l, "data:video/"):
		return true
	}
	return !strings.Contains(l, ":") && !strings.HasPrefix(l, "//")
}

// Sanitize is s as safe HTML: plain text is escaped, and of markup only
// the allowed elements and attributes are kept (others keep their text).
func Sanitize(s string) string {
	if !IsRich(s) {
		return html.EscapeString(s)
	}
	nodes, err := html.ParseFragment(strings.NewReader(s), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return html.EscapeString(s)
	}
	var b bytes.Buffer
	for _, n := range nodes {
		clean(&b, n)
	}
	return b.String()
}

func clean(b *bytes.Buffer, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		b.WriteString(html.EscapeString(n.Data))
		return
	case html.ElementNode:
	default:
		return // comments, doctypes
	}
	attrs, ok := allowed[n.DataAtom]
	if n.DataAtom == atom.Script || n.DataAtom == atom.Style {
		return // their text is code, not words
	}
	if ok {
		b.WriteString("<" + n.Data)
		for _, a := range n.Attr {
			if !contains(attrs, a.Key) {
				continue
			}
			if (a.Key == "src" || a.Key == "href") && !safeURL(a.Val, a.Key == "href") {
				continue
			}
			b.WriteString(" " + a.Key + `="` + html.EscapeString(a.Val) + `"`)
		}
		switch n.DataAtom {
		case atom.Audio, atom.Video:
			b.WriteString(" controls")
		case atom.A:
			b.WriteString(` target="_blank" rel="noopener noreferrer"`)
		}
		b.WriteString(">")
		if n.DataAtom == atom.Br || n.DataAtom == atom.Img || n.DataAtom == atom.Source {
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		clean(b, c)
	}
	if ok {
		b.WriteString("</" + n.Data + ">")
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Plain is the text of s without markup, as typed: H<sub>2</sub>O gives
// H2O, and furigana (<rt>, <rp>) are left out; a picture counts as its
// alt text. Text without markup is returned as it is.
func Plain(s string) string {
	if !IsRich(s) {
		return s
	}
	nodes, err := html.ParseFragment(strings.NewReader(s), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return s
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch {
		case n.Type == html.TextNode:
			b.WriteString(n.Data)
		case n.Type == html.ElementNode:
			switch n.DataAtom {
			case atom.Rt, atom.Rp, atom.Script, atom.Style:
				return
			case atom.Br:
				b.WriteString(" ")
			case atom.Img:
				for _, a := range n.Attr {
					if a.Key == "alt" {
						b.WriteString(a.Val)
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
