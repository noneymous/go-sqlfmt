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
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.writeSelect(buf, token, previousToken, formatter.IndentLevel, i)
		} else {

			// Set peripheral parameters
			switch v2 := el.(type) {
			case *Parenthesis:
				v2.IsColumnArea = true
				v2.PositionInParent = i
			case *Subquery:
				v2.IsColumnArea = true
			case *Function:
				v2.IsColumnArea = true
			}

			// Increment indent, as everything within SELECT should be indented
			el.AddIndent(1)

			// Recursively format nested elements
			_ = el.Format(buf, elements, i)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		} else {
			previousToken = Token{}
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

func (formatter *Select) writeSelect(buf *bytes.Buffer, token, previousToken Token, indent int, position int) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element
	switch {
	case token.Type == lexer.SELECT || token.Type == lexer.INTO:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case token.Type == lexer.EXISTS:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case token.Type == lexer.AS || token.Type == lexer.DISTINCT || token.Type == lexer.DISTINCTROW || token.Type == lexer.WITHIN || token.Type == lexer.GROUP || token.Type == lexer.ON:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case position == 1 || previousToken.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
	case token.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}