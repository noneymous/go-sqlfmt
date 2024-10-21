package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// LimitClause such as LIMIT, OFFSET, FETCH FIRST
type LimitClause struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (l *LimitClause) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	elements, err := processPunctuation(l.Element)
	if err != nil {
		return err
	}

	var previousToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, l.IndentLevel)
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
func (l *LimitClause) IncrementIndentLevel(lev int) {
	l.IndentLevel += lev
}
