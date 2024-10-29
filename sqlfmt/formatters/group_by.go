package formatters

import (
	"bytes"
)

// GroupBy group formatter
type GroupBy struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format reindents and formats elements accordingly
func (formatter *GroupBy) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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
	columnCount := 0
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case Token, string:
			if errWrite := writeWithComma(buf, INDENT, NEWLINE, WHITESPACE, v, formatter.IndentLevel, &columnCount); errWrite != nil {
				return errWrite
			}
		case Formatter:
			_ = v.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *GroupBy) AddIndent(lev int) {
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
