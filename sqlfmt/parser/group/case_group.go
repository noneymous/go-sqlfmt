package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Case Clause
type Case struct {
	Element        []lexer.Reindenter
	IndentLevel    int
	hasCommaBefore bool
}

// Reindent reindents its elements
func (c *Case) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(c.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeCase(buf, token, c.IndentLevel, c.hasCommaBefore)
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

// IncrementIndentLevel increments by its specified increment level
func (c *Case) IncrementIndentLevel(lev int) {
	c.IndentLevel += lev
}
