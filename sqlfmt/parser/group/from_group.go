package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// From clause
type From struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (f *From) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(f.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, f.IndentLevel)
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

// IncrementIndentLevel indents by its specified indent level
func (f *From) IncrementIndentLevel(lev int) {
	f.IndentLevel += lev
}
