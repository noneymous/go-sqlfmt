package group

import (
	"bytes"
	"fmt"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/pkg/errors"
)

// Select clause
type Select struct {
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (s *Select) Reindent(buf *bytes.Buffer, lastParentToken lexer.Token) error {
	columnCount = 0

	src, err := processPunctuation(s.Element)
	if err != nil {
		return err
	}
	elements := separate(src)

	var previousToken lexer.Token
	for _, el := range elements {
		switch v := el.(type) {
		case lexer.Token, string:
			if errWrite := writeSelect(buf, el, s.IndentLevel); errWrite != nil {
				return errors.Wrap(errWrite, "writeSelect failed")
			}
		case *Case:
			if previousToken.Type == lexer.COMMA {
				v.hasCommaBefore = true
			}
			_ = v.Reindent(buf, previousToken)
			// Case group in Select clause must be in column area
			columnCount++
		case *Parenthesis:
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, previousToken)
			columnCount++
		case *Subquery:
			if previousToken.Type == lexer.EXISTS {
				_ = v.Reindent(buf, previousToken)
				continue
			}
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, previousToken)
		case *Function:
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, previousToken)
			columnCount++
		case lexer.Reindenter:
			_ = v.Reindent(buf, previousToken)
			columnCount++
		default:
			return fmt.Errorf("can not reindent %#v", v)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (s *Select) IncrementIndentLevel(lev int) {
	s.IndentLevel += lev
}
