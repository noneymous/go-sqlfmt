package group

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// TieClause such as UNION, EXCEPT, INTERSECT
type TieClause struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (tie *TieClause) Reindent(buf *bytes.Buffer, prev lexer.Token) error {
	elements, err := processPunctuation(tie.Element)
	if err != nil {
		return err
	}

	var lastToken lexer.Token
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			write(buf, token, tie.IndentLevel)
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

// IncrementIndentLevel increments by its specified indent level
func (tie *TieClause) IncrementIndentLevel(lev int) {
	tie.IndentLevel += lev
}
