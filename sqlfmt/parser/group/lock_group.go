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
func (l *Lock) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	var lastToken lexer.Token
	for _, el := range l.Element {
		if token, ok := el.(lexer.Token); ok {
			writeLock(buf, token)
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
func (l *Lock) IncrementIndentLevel(lev int) {
	l.IndentLevel += lev
}
