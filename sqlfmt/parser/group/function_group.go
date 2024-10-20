package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Function clause
type Function struct {
	Element      []lexer.Reindenter
	IndentLevel  int
	InColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (f *Function) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(f.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeFunction(buf, token, lastToken, f.IndentLevel, f.ColumnCount, f.InColumnArea)
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
func (f *Function) IncrementIndentLevel(lev int) {
	f.IndentLevel += lev
}
