package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Insert clause
type Insert struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (insert *Insert) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(insert.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, insert.IndentLevel)
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
func (insert *Insert) IncrementIndentLevel(lev int) {
	insert.IndentLevel += lev
}
