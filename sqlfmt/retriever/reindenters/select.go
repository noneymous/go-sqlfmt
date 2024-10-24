package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Select group reindenter
type Select struct {
	Options     *lexer.Options // Options used later to format element
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (group *Select) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Reset column count
	columnCount = 0

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken lexer.Token
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case lexer.Token, string:
			if errWrite := group.writeSelect(buf, el, previousToken, group.IndentLevel); errWrite != nil {
				return fmt.Errorf("failed writing select: %s", errWrite)
			}
		case *Case:
			if previousToken.Type == lexer.COMMA {
				v.hasCommaBefore = true
			}
			_ = v.Reindent(buf, elements, i)
			// Case group in Select clause must be in column area
			columnCount++
		case *Parenthesis:
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, elements, i)
			columnCount++
		case *Subquery:
			if previousToken.Type == lexer.EXISTS {
				_ = v.Reindent(buf, elements, i)
				continue
			}
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, elements, i)
		case *Function:
			v.IsColumnArea = true
			v.ColumnCount = columnCount
			_ = v.Reindent(buf, elements, i)
			columnCount++
		case lexer.Reindenter:
			_ = v.Reindent(buf, elements, i)
			columnCount++
		default:
			return fmt.Errorf("can not reindent %#v", v)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (group *Select) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Select) writeSelect(buf *bytes.Buffer, el interface{}, previousToken lexer.Token, indent int) error {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	if token, ok := el.(lexer.Token); ok {
		switch token.Type {
		case lexer.SELECT, lexer.INTO:
			buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		case lexer.EXISTS, lexer.ANY:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			columnCount++
		case lexer.AS, lexer.DISTINCT, lexer.DISTINCTROW, lexer.GROUP, lexer.ON:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		case lexer.COMMA:
			//buf.WriteString(fmt.Sprintf("%s%s%s%s", token.Value, NewLine, strings.Repeat(DoubleWhiteSpace, indent), WhiteSpace))
			buf.WriteString(fmt.Sprintf("%s", token.Value))
		default:
			return fmt.Errorf("can not reindent %#v", token.Value)
		}
	} else if str, ok := el.(string); ok {
		str = strings.Trim(str, WHITESPACE)

		// Add newline and spacing for first columns or if previous token ended with comma
		if columnCount == 0 || previousToken.Type == lexer.COMMA {
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, str))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, str))
		}

		// Increment column count
		columnCount++
	}
	return nil
}
