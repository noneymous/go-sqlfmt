package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Select group formatter
type Select struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
	ColumnCount int
}

// Format reindents and formats elements accordingly
func (formatter *Select) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case Token, string:
			if errWrite := formatter.writeSelect(buf, el, previousToken, formatter.IndentLevel); errWrite != nil {
				return fmt.Errorf("failed writing select: %s", errWrite)
			}
		case *Case:
			if previousToken.Type == lexer.COMMA {
				v.hasCommaBefore = true
			}
			_ = v.Format(buf, elements, i)
			formatter.ColumnCount++ // Case group in Select clause must be in column area
		case *Parenthesis:
			v.IsColumnArea = true
			v.ColumnCount = formatter.ColumnCount
			_ = v.Format(buf, elements, i)
			formatter.ColumnCount++
		case *Subquery:
			if previousToken.Type == lexer.EXISTS {
				_ = v.Format(buf, elements, i)
				continue
			}
			v.IsColumnArea = true
			v.ColumnCount = formatter.ColumnCount
			_ = v.Format(buf, elements, i)
		case *Function:
			v.IsColumnArea = true
			v.ColumnCount = formatter.ColumnCount
			_ = v.Format(buf, elements, i)
			formatter.ColumnCount++
		case Formatter:
			_ = v.Format(buf, elements, i)
			formatter.ColumnCount++
		default:
			return fmt.Errorf("invalid element '%#v'", v)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Select) AddIndent(lev int) {
	formatter.IndentLevel += lev

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, formatter.Whitespace)
	if err != nil {
		elements = formatter.Elements
	}

	// Iterate and increase indent of child elements too
	for _, el := range elements {
		el.AddIndent(lev)
	}
}

func (formatter *Select) writeSelect(buf *bytes.Buffer, el interface{}, previousToken Token, indent int) error {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element
	if token, isToken := el.(Token); isToken {
		switch token.Type {
		case lexer.SELECT, lexer.INTO:
			buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		case lexer.EXISTS, lexer.ANY:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			formatter.ColumnCount++
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
		if formatter.ColumnCount == 0 || previousToken.Type == lexer.COMMA {
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, str))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, str))
		}

		// Increment column count
		formatter.ColumnCount++
	}
	return nil
}
