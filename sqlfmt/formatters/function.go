package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Function group formatter
type Function struct {
	Elements     []Formatter
	IndentLevel  int
	*Options     // Options used later to format element
	IsColumnArea bool
	ColumnCount  int
}

// Format reindents and formats elements accordingly
func (formatter *Function) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
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

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.writeFunction(buf, token, previousToken, previousParentToken, formatter.IndentLevel, formatter.IsColumnArea, formatter.ColumnCount, isNested)
		} else {
			_ = el.Format(buf, elements, i)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Function) AddIndent(lev int) {
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

func (formatter *Function) writeFunction(buf *bytes.Buffer, token, previousToken, previousParentToken Token, indent int, isColumnArea bool, columnCount int, isNested bool) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

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
