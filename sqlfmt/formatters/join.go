package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Join group formatter
type Join struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format reindents and formats elements accordingly
func (formatter *Join) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var isFirst = true
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.writeJoin(buf, token, formatter.IndentLevel, isFirst)
		} else {
			_ = el.Format(buf, elements, i)
		}
		isFirst = false
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Join) AddIndent(lev int) {
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

func (formatter *Join) writeJoin(buf *bytes.Buffer, token Token, indent int, isFirst bool) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element
	switch {
	case isFirst && token.IsJoinStart():
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case token.Type == lexer.ON || token.Type == lexer.USING:
		buf.WriteString(fmt.Sprintf(" %s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
