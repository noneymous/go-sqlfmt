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
func (insert *Insert) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(insert.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, insert.IndentLevel)
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

// IncrementIndentLevel increments by its specified indent level
func (insert *Insert) IncrementIndentLevel(lev int) {
	insert.IndentLevel += lev
}
