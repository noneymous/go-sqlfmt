package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Select group reindenter
type Select struct {
	Options     *Options // Options used later to format element
	Element     []Reindenter
	IndentLevel int
	ColumnCount int
}

// Reindent reindents its elements
func (group *Select) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case Token, string:
			if errWrite := group.writeSelect(buf, el, previousToken, group.IndentLevel); errWrite != nil {
				return fmt.Errorf("failed writing select: %s", errWrite)
			}
		case *Case:
			if previousToken.Type == lexer.COMMA {
				v.hasCommaBefore = true
			}
			_ = v.Reindent(buf, elements, i)
			group.ColumnCount++ // Case group in Select clause must be in column area
		case *Parenthesis:
			v.IsColumnArea = true
			v.ColumnCount = group.ColumnCount
			_ = v.Reindent(buf, elements, i)
			group.ColumnCount++
		case *Subquery:
			if previousToken.Type == lexer.EXISTS {
				_ = v.Reindent(buf, elements, i)
				continue
			}
			v.IsColumnArea = true
			v.ColumnCount = group.ColumnCount
			_ = v.Reindent(buf, elements, i)
		case *Function:
			v.IsColumnArea = true
			v.ColumnCount = group.ColumnCount
			_ = v.Reindent(buf, elements, i)
			group.ColumnCount++
		case Reindenter:
			_ = v.Reindent(buf, elements, i)
			group.ColumnCount++
		default:
			return fmt.Errorf("invalid element '%#v'", v)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndent increments by its specified indent level
func (group *Select) IncrementIndent(lev int) {
	group.IndentLevel += lev

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, group.Options.Whitespace)
	if err != nil {
		elements = group.Element
	}

	// Iterate and increase indent of child elements too
	for _, el := range elements {
		el.IncrementIndent(lev)
	}
}

func (group *Select) writeSelect(buf *bytes.Buffer, el interface{}, previousToken Token, indent int) error {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	if token, isToken := el.(Token); isToken {
		switch token.Type {
		case lexer.SELECT, lexer.INTO:
			buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		case lexer.EXISTS, lexer.ANY:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			group.ColumnCount++
		case lexer.AS, lexer.DISTINCT, lexer.DISTINCTROW, lexer.GROUP, lexer.ON:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		case lexer.COMMA:
			buf.WriteString(fmt.Sprintf("%s", token.Value))
		default:
			return fmt.Errorf("unexpected SELECT token '%#v'", token.Value)
		}
	} else if str, isStr := el.(string); isStr {

		// Remove whitespaces from value
		str = strings.Trim(str, WHITESPACE)

		// Add newline and spacing for first columns or if previous token ended with comma
		if group.ColumnCount == 0 || previousToken.Type == lexer.COMMA {
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, str))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, str))
		}

		// Increment column count
		group.ColumnCount++
	}
	return nil
}
