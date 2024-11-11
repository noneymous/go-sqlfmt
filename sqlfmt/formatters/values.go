package formatters

import (
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"strings"
)

// Values group formatter
type Values struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Values) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.WriteValues(buf, token, formatter.IndentLevel)
		} else {

			// Break parenthesis group (list of values) into new line
			switch v := el.(type) {
			case *Parenthesis:
				v.IsColumnArea = true
			}

			// Recursively format nested elements
			_ = el.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Values) AddIndent(lev int) {
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

func (formatter *Values) WriteValues(buf *bytes.Buffer, token Token, indent int) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	switch {
	case token.ContinueNewline():
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case token.Type == lexer.COMMA: // Write comma token without whitespace
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
