package formatters

import (
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"strings"
)

// Lock group formatter
type Lock struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Lock) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range formatter.Elements {
		if token, ok := el.(Token); ok {
			formatter.writeLock(buf, token, previousToken, formatter.IndentLevel)
		} else {
			_ = el.Format(buf, formatter.Elements, i)
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

func (formatter *Lock) writeLock(buf *bytes.Buffer, token, previousToken Token, indent int) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Any token following a line comment must start on a new line
	if previousToken.IsLineComment() {
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		return
	}

	// Write element
	switch token.Type {
	case lexer.LOCK:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))
	case lexer.IN:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))

	// Write common token values
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
