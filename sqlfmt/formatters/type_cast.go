package formatters

import (
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// TypeCast group formatter
type TypeCast struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format reindents and formats elements accordingly
func (formatter *TypeCast) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for _, el := range elements {
		if token, ok := el.(Token); ok {
			formatter.writeTypeCast(buf, token)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *TypeCast) AddIndent(lev int) {
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

func (formatter *TypeCast) writeTypeCast(buf *bytes.Buffer, token Token) {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Write element
	switch token.Type {
	case lexer.TYPE:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s%s", token.Value, WHITESPACE))
	default:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	}
}
