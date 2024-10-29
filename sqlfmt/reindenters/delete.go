package reindenters

import (
	"bytes"
)

// Delete group reindenter
type Delete struct {
	Options     *Options // Options used later to format element
	Element     []Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (group *Delete) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(Token); ok {
			write(buf, INDENT, NEWLINE, WHITESPACE, token, previousToken, group.IndentLevel, false)
		} else {
			_ = el.Reindent(buf, elements, i)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndent increments by its specified indent level
func (group *Delete) IncrementIndent(lev int) {
	group.IndentLevel += lev

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, group.Options.Whitespace)
	if err != nil {
		elements = group.Element
	}

	// Iterate and increase indent of child elements too
	for _, el := range elements {
		el.IncrementIndent(lev)
	}
}
