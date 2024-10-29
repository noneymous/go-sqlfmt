package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Parenthesis group formatter
type Parenthesis struct {
	Elements     []Formatter
	IndentLevel  int
	*Options     // Options used later to format element
	IsColumnArea bool
	ColumnCount  int
}

// Format reindents and formats elements accordingly
func (formatter *Parenthesis) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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

	// Check if parenthesis group has nested element
	var hasIntermediate = false
	for _, el := range elements {
		if _, ok := el.(Token); !ok {
			hasIntermediate = true
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.writeParenthesis(buf, token, previousParentToken, formatter.IndentLevel, formatter.IsColumnArea, formatter.ColumnCount, i, hasIntermediate)
		} else {
			el.AddIndent(1)
			_ = el.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Parenthesis) AddIndent(lev int) {
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

func (formatter *Parenthesis) writeParenthesis(buf *bytes.Buffer, token, previousParentToken Token, indent int, isColumnArea bool, columnCount, i int, hasIntermediate bool) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Write element
	switch {

	// Fix missing whitespaces for select clauses moved to the new line in select columns
	case isColumnArea && previousParentToken.ContinueLine() && token.Type == lexer.STARTPARENTHESIS:
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
	case hasIntermediate && i == 1:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
	case i == 1:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}
