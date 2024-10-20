package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Parenthesis clause
type Parenthesis struct {
	Element      []lexer.Reindenter
	IndentLevel  int
	InColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (p *Parenthesis) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	var hasStartBefore bool

	elements, err := processPunctuation(p.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for i, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			hasStartBefore = i == 1
			writeParenthesis(buf, token, p.IndentLevel, p.ColumnCount, p.InColumnArea, hasStartBefore)
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
func (p *Parenthesis) IncrementIndentLevel(lev int) {
	p.IndentLevel += lev
}
