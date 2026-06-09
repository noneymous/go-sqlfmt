package formatters

import (
	"bytes"
	"fmt"
	"strings"
)

// Generic group formatter
// The generic group formatter can be applied if no special formatting rules are required
type Generic struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Generic) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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

		// Write element or recursively call its Format function
		if token, ok := el.(Token); ok {
			formatter.write(buf, token, previousToken, formatter.IndentLevel, i)
		} else {

			// Set peripheral parameters to tell child elements to write to the same line
			switch v := el.(type) {
			case *Or:
				v.SameLine = true
			case *And:
				v.SameLine = true
			}

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
func (formatter *Generic) AddIndent(lev int) {
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

func (formatter *Generic) write(buf *bytes.Buffer, token, previousToken Token, indent, position int) {

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
	switch {
	case position == 0:
		buf.WriteString(fmt.Sprintf("%s", token.Value))

	// Write common token values
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
