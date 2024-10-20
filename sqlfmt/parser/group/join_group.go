package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Join clause
type Join struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindent its elements
func (j *Join) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(j.Element)
	if err != nil {
		return err
	}

	var isFirst = true
	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeJoin(buf, token, j.IndentLevel, isFirst)
		} else {
			_ = el.Reindent(buf, lastToken)
		}
		isFirst = false

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			lastToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified increment level
func (j *Join) IncrementIndentLevel(lev int) {
	j.IndentLevel += lev
}
