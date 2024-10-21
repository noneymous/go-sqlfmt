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
func (j *Join) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(j.Element)
	if err != nil {
		return err
	}

	var isFirst = true
	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeJoin(buf, token, j.IndentLevel, isFirst)
		} else {
			_ = el.Reindent(buf, previousToken)
		}
		isFirst = false

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified increment level
func (j *Join) IncrementIndentLevel(lev int) {
	j.IndentLevel += lev
}
