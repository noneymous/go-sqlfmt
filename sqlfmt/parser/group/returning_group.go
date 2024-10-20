package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Returning clause
type Returning struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (r *Returning) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	columnCount = 0

	src, err := processPunctuation(r.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range separate(src) {
		switch v := el.(type) {
		case lexer.Token, string:
			if err := writeWithComma(buf, v, r.IndentLevel); err != nil {
				return err
			}
		case lexer.Reindenter:
			_ = v.Reindent(buf, lastToken)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			lastToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (r *Returning) IncrementIndentLevel(lev int) {
	r.IndentLevel += lev
}
