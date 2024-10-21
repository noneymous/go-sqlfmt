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
func (r *Returning) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	columnCount = 0

	src, err := processPunctuation(r.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range separate(src) {
		switch v := el.(type) {
		case lexer.Token, string:
			if err := writeWithComma(buf, v, r.IndentLevel); err != nil {
				return err
			}
		case lexer.Reindenter:
			_ = v.Reindent(buf, previousToken)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (r *Returning) IncrementIndentLevel(lev int) {
	r.IndentLevel += lev
}
