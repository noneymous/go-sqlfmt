package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Subquery group
type Subquery struct {
	Element      []lexer.Reindenter
	IndentLevel  int
	InColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (s *Subquery) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(s.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			writeSubquery(buf, token, s.IndentLevel, s.ColumnCount, s.InColumnArea)
		} else {
			if s.InColumnArea {
				el.IncrementIndentLevel(1)
				_ = el.Reindent(buf, lastToken)
			} else {
				_ = el.Reindent(buf, lastToken)
			}
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			lastToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (s *Subquery) IncrementIndentLevel(lev int) {
	s.IndentLevel += lev
}
