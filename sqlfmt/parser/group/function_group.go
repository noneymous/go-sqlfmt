package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Function clause
type Function struct {
	Element      []lexer.Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (f *Function) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(f.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeFunction(buf, token, previousToken, f.IndentLevel, f.ColumnCount, f.IsColumnArea)
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
func (f *Function) IncrementIndentLevel(lev int) {
	f.IndentLevel += lev
}
