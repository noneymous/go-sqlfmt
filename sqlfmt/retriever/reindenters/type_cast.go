package reindenters

import (
	"bytes"
	"fmt"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// TypeCast group reindenter
type TypeCast struct {
	Options     *lexer.Options // Options used later to format element
	Element     []lexer.Reindenter
	IndentLevel int
}

// Reindent reindents its elements
func (group *TypeCast) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for _, el := range elements {
		if token, ok := el.(lexer.Token); ok {
			group.writeTypeCast(buf, token)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (group *TypeCast) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *TypeCast) writeTypeCast(buf *bytes.Buffer, token lexer.Token) {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch token.Type {
	case lexer.TYPE:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s%s", token.Value, WHITESPACE))
	default:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	}
}
