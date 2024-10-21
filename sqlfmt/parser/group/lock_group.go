package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Lock clause
type Lock struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindent its elements
func (l *Lock) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	var previousToken lexer.Token
	for _, el := range l.Element {
		if token, ok := el.(lexer.Token); ok {
			writeLock(buf, token)
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

// IncrementIndentLevel increments by its specified increment level
func (l *Lock) IncrementIndentLevel(lev int) {
	l.IndentLevel += lev
}
