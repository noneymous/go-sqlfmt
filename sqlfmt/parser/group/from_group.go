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
func (f *From) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(f.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, f.IndentLevel)
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

// IncrementIndentLevel indents by its specified indent level
func (f *From) IncrementIndentLevel(lev int) {
	f.IndentLevel += lev
}
