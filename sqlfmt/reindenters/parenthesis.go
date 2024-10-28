package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Parenthesis group reindenter
type Parenthesis struct {
	Options      *Options // Options used later to format element
	Element      []Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (group *Parenthesis) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {
	var hasStartBefore bool

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

	// Check if parenthesis group has nested element
	var hasIntermediate = false
	for _, el := range elements {
		if _, ok := el.(Token); !ok {
			hasIntermediate = true
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(Token); ok {
			hasStartBefore = i == 1
			group.writeParenthesis(buf, token, previousParentToken, group.IndentLevel, group.ColumnCount, group.IsColumnArea, hasStartBefore, hasIntermediate)
		} else {
			_ = el.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel indents by its specified indent level
func (group *Parenthesis) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Parenthesis) writeParenthesis(buf *bytes.Buffer, token, previousParentToken Token, indent, columnCount int, isColumnArea, isFirstValue, hasIntermediate bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	switch {

	// Fix missing whitespaces for select clauses moved to the new line in select columns
	case isColumnArea && previousParentToken.Type == lexer.ON && token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))

	// Don't move ENDPARENTHESIS to new line if there is no intermediate segment as a reason
	case !hasIntermediate && token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s", token.Value))

	// In the column area, parenthesis are shifted to the next line and indented
	case isColumnArea && token.Type == lexer.STARTPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case isColumnArea && token.Type == lexer.ENDPARENTHESIS: // Parenthesis ends in a new line with indent
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))

	case hasIntermediate && token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case hasIntermediate && token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))

	// Parenthesis starts in same line without indent
	case token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case token.Type == lexer.ENDPARENTHESIS: // Parenthesis ends in next line without indent
		buf.WriteString(fmt.Sprintf("%s", token.Value))

	// Token values
	case token.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case isFirstValue:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
