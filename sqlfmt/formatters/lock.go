package formatters

import (
	"bytes"
	"fmt"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Lock group formatter
type Lock struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format reindents and formats elements accordingly
func (formatter *Lock) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range formatter.Elements {
		if token, ok := el.(Token); ok {
			formatter.writeLock(buf, token)
		} else {
			_ = el.Format(buf, formatter.Elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Lock) AddIndent(lev int) {
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

func (formatter *Lock) writeLock(buf *bytes.Buffer, token Token) {

	// Prepare short variables for better visibility
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element
	switch token.Type {
	case lexer.LOCK:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))
	case lexer.IN:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
