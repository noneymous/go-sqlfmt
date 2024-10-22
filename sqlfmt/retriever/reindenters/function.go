package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Function group reindenter
type Function struct {
	Options      *lexer.Options // Options used later to format element
	Element      []lexer.Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (group *Function) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Retrieve some information about parent elements
	var previousParentToken lexer.Token
	var isNestedFunction = false
	if len(parent) > parentIdx {

		// Get last token written by parent
		if token, ok := parent[parentIdx-1].(lexer.Token); ok {
			previousParentToken = token
		}

		// Check if function is nested within parent function
		if len(parent) > parentIdx {
			if token, ok2 := parent[0].(lexer.Token); ok2 && token.Type == lexer.FUNCTION {
				isNestedFunction = true
			}
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken lexer.Token
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(lexer.Token); ok {
			group.writeFunction(buf, token, previousToken, previousParentToken, group.IndentLevel, group.ColumnCount, group.IsColumnArea, isNestedFunction)
		} else {
			_ = el.Reindent(buf, elements, i)
		}

		// Remember last Token element
		if token, ok := el.(lexer.Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (group *Function) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Function) writeFunction(buf *bytes.Buffer, token, previousToken, previousParentToken lexer.Token, indent, columnCount int, isColumnArea, isNestedFunction bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch {
	case token.Type == lexer.FUNCTION && columnCount == 0 && isColumnArea:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
	case token.Type == lexer.FUNCTION && previousParentToken.Type == lexer.COMMA && !isNestedFunction:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
	case token.Type == lexer.FUNCTION:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case token.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case previousToken.Type == lexer.STARTPARENTHESIS || token.Type == lexer.STARTPARENTHESIS || token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
