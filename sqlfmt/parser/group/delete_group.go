package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Delete clause
type Delete struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (d *Delete) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(d.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, d.IndentLevel)
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
func (d *Delete) IncrementIndentLevel(lev int) {
	d.IndentLevel += lev
}
