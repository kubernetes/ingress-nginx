package html

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// EscapeHTMLCallouts writes html-escaped d to w. It escapes &, <, > and " characters, *but*
// expands callouts <<N>> with the callout HTML, i.e. by calling r.callout() with a newly created
// ast.Callout node.
func (r *Renderer) EscapeHTMLCallouts(w io.Writer, d []byte) {
	ld := len(d)
Parse:
	for i := 0; i < ld; i++ {
		for _, comment := range r.opts.Comments {
			if !bytes.HasPrefix(d[i:], comment) {
				break
			}

			lc := len(comment)
			if i+lc < ld {
				if id, consumed := parser.IsCallout(d[i+lc:]); consumed > 0 {
					// We have seen a callout
					callout := &ast.Callout{ID: id}
					r.callout(w, callout)
					i += consumed + lc - 1
					continue Parse
				}
			}
		}

		escSeq := Escaper[d[i]]
		if escSeq != nil {
			w.Write(escSeq)
		} else {
			w.Write([]byte{d[i]})
		}
	}
}
