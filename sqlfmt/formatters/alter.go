package formatters

import (
	"bytes"
)

// Alter group formatter
type Alter struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Alter) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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

			// In some CREATE cases sub queries don't need to be put in between parentheses,
			// hence aren't detected as sub queries, so it is necessary to manually indent them
			switch v := el.(type) {
			case *Parenthesis, *Subquery:
				// Will already be indented correctly
			default:
				v.AddIndent(1)
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
func (formatter *Alter) AddIndent(lev int) {
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
