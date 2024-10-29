package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Case group formatter
type Case struct {
	Elements       []Formatter
	IndentLevel    int
	*Options       // Options used later to format element
	hasCommaBefore bool
}

// Format reindents and formats elements accordingly
func (formatter *Case) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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
			formatter.writeCase(buf, token, formatter.IndentLevel)
		} else {
			el.AddIndent(2)
			_ = el.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Case) AddIndent(lev int) {
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

func (formatter *Case) writeCase(buf *bytes.Buffer, token Token, indent int) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element

	switch token.Type {
	case lexer.CASE, lexer.END:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
	case lexer.WHEN, lexer.ELSE:
		buf.WriteString(fmt.Sprintf("%s%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, INDENT, token.Value))
	case lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		if strings.HasPrefix(token.Value, "::") {
			buf.WriteString(fmt.Sprintf("%s", token.Value))
		} else {
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		}
	}
}
