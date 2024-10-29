package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Join group reindenter
type Join struct {
	Options     *Options // Options used later to format element
	Element     []Reindenter
	IndentLevel int
}

// Reindent reindent its elements
func (group *Join) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var isFirst = true
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(Token); ok {
			group.writeJoin(buf, token, group.IndentLevel, isFirst)
		} else {
			_ = el.Reindent(buf, elements, i)
		}
		isFirst = false
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndent increments by its specified increment level
func (group *Join) IncrementIndent(lev int) {
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

func (group *Join) writeJoin(buf *bytes.Buffer, token Token, indent int, isFirst bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch {
	case isFirst && token.IsJoinStart():
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case token.Type == lexer.ON || token.Type == lexer.USING:
		buf.WriteString(fmt.Sprintf(" %s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
