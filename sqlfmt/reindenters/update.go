package reindenters

import (
	"bytes"
)

// Update group reindenter
type Update struct {
	Options     *Options // Options used later to format element
	Element     []Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (group *Update) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

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
	columnCount := 0
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case Token, string:
			if errWrite := writeWithComma(buf, INDENT, NEWLINE, WHITESPACE, v, group.IndentLevel, &columnCount); errWrite != nil {
				return errWrite
			}
		case Reindenter:
			_ = v.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndent increments by its specified indent level
func (group *Update) IncrementIndent(lev int) {
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
