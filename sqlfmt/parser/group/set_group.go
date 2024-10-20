package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Set clause
type Set struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (s *Set) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	columnCount = 0

	src, err := processPunctuation(s.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range separate(src) {
		switch v := el.(type) {
		case lexer.Token, string:
			if err := writeWithComma(buf, v, s.IndentLevel); err != nil {
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
func (s *Set) IncrementIndentLevel(lev int) {
	s.IndentLevel += lev
}
