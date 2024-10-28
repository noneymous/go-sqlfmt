package reindenters

import (
	"bytes"
	"fmt"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Lock group reindenter
type Lock struct {
	Options     *Options // Options used later to format element
	Element     []Reindenter
	IndentLevel int
}

// Reindent reindent its elements
func (group *Lock) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range group.Element {
		if token, ok := el.(Token); ok {
			group.writeLock(buf, token)
		} else {
			_ = el.Reindent(buf, group.Element, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified increment level
func (group *Lock) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Lock) writeLock(buf *bytes.Buffer, token Token) {

	// Prepare short variables for better visibility
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch token.Type {
	case lexer.LOCK:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))
	case lexer.IN:
		buf.WriteString(fmt.Sprintf("%s%s", NEWLINE, token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
