package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Update clause
type Update struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (u *Update) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	columnCount = 0

	src, err := processPunctuation(u.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range separate(src) {
		switch v := el.(type) {
		case lexer.Token, string:
			if err := writeWithComma(buf, v, u.IndentLevel); err != nil {
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
func (u *Update) IncrementIndentLevel(lev int) {
	u.IndentLevel += lev
}
