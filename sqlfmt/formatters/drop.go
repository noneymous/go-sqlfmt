package formatters

import (
	"bytes"
)

// Drop group formatter
type Drop struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Drop) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
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
			writeCreate(buf, INDENT, NEWLINE, WHITESPACE, token, previousToken, formatter.IndentLevel, i)
		} else {

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
func (formatter *Drop) AddIndent(lev int) {
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
