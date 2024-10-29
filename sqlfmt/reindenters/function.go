package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Function group reindenter
type Function struct {
	Options      *Options // Options used later to format element
	Element      []Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (group *Function) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Get last token written by parent
	var previousParentToken Token
	if len(parent) > parentIdx {
		if token, ok := parent[parentIdx-1].(Token); ok {
			previousParentToken = token
		}
	}

	// Check if function is nested within parent function
	var isNested = false
	if len(parent) > parentIdx { // Required for unit test where parent might be omitted
		if token, ok2 := parent[0].(Token); ok2 && token.Type == lexer.FUNCTION {
			isNested = true
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(Token); ok {
			group.writeFunction(buf, token, previousToken, previousParentToken, group.IndentLevel, group.IsColumnArea, group.ColumnCount, isNested)
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
func (group *Function) IncrementIndent(lev int) {
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

func (group *Function) writeFunction(buf *bytes.Buffer, token, previousToken, previousParentToken Token, indent int, isColumnArea bool, columnCount int, isNested bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch {

	// Write start/end parenthesis of function without whitespace
	case token.Type == lexer.STARTPARENTHESIS || token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s", token.Value))

	// Move function to new line if it is the first element in the column area
	case token.Type == lexer.FUNCTION && isColumnArea && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))

	// Move function to new line if it isn't nested within another one
	case token.Type == lexer.FUNCTION && previousParentToken.Type == lexer.COMMA && !isNested:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))

	// Write function to the same line otherwise
	case token.Type == lexer.FUNCTION:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))

	// Token values
	case token.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case previousToken.Type == lexer.STARTPARENTHESIS: // Write first value of function without whitespace
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
