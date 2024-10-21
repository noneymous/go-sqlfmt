package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// OrGroup clause
type OrGroup struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (o *OrGroup) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(o.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, o.IndentLevel)
		} else {
			_ = el.Reindent(buf, previousToken)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified increment level
func (o *OrGroup) IncrementIndentLevel(lev int) {
	o.IndentLevel += lev
}
