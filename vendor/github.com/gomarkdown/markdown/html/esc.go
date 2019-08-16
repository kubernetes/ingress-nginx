package html

import (
	"html"
	"io"
)

var Escaper = [256][]byte{
	'&': []byte("&amp;"),
	'<': []byte("&lt;"),
	'>': []byte("&gt;"),
	'"': []byte("&quot;"),
}

// EscapeHTML writes html-escaped d to w. It escapes &, <, > and " characters.
func EscapeHTML(w io.Writer, d []byte) {
	var start, end int
	n := len(d)
	for end < n {
		escSeq := Escaper[d[end]]
		if escSeq != nil {
			w.Write(d[start:end])
			w.Write(escSeq)
			start = end + 1
		}
		end++
	}
	if start < n && end <= n {
		w.Write(d[start:end])
	}
}

func escLink(w io.Writer, text []byte) {
	unesc := html.UnescapeString(string(text))
	EscapeHTML(w, []byte(unesc))
}

// Escape writes the text to w, but skips the escape character.
func Escape(w io.Writer, text []byte) {
	esc := false
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' {
			esc = !esc
		}
		if esc && text[i] == '\\' {
			continue
		}
		w.Write([]byte{text[i]})
	}
}
