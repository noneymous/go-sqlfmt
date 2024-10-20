package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// GroupBy clause
type GroupBy struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (g *GroupBy) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	columnCount = 0

	elements, err := processPunctuation(g.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range separate(elements) {
		switch v := el.(type) {
		case lexer.Token, string:
			if err := writeWithComma(buf, v, g.IndentLevel); err != nil {
				return err
			}
		case lexer.Reindenter:
			_ = v.Reindent(buf, lastToken)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			lastToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (g *GroupBy) IncrementIndentLevel(lev int) {
	g.IndentLevel += lev
}
