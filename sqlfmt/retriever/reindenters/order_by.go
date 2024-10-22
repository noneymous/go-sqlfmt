package reindenters

import (
	"bytes"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// OrderBy group reindenter
type OrderBy struct {
	Options     *lexer.Options // Options used later to format element
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (group *OrderBy) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Reset column count
	columnCount = 0

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range separate(elements, WHITESPACE) {
		switch v := el.(type) {
		case lexer.Token, string:
			if errWrite := writeWithComma(buf, INDENT, NEWLINE, WHITESPACE, v, group.IndentLevel); errWrite != nil {
				return errWrite
			}
		case lexer.Reindenter:
			_ = v.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (group *OrderBy) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}
