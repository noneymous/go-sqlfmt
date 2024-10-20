package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Having clause
type Having struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (h *Having) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(h.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, h.IndentLevel)
		} else {
			_ = el.Reindent(buf, lastToken)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			lastToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (h *Having) IncrementIndentLevel(lev int) {
	h.IndentLevel += lev
}
