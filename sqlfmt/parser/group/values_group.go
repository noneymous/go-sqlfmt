package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Values clause
type Values struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (val *Values) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(val.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, val.IndentLevel)
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
func (val *Values) IncrementIndentLevel(lev int) {
	val.IndentLevel += lev
}
