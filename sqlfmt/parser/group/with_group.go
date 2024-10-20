package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// With clause
type With struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (w *With) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(w.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, w.IndentLevel)
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
func (w *With) IncrementIndentLevel(lev int) {
	w.IndentLevel += lev
}
