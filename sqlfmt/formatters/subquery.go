package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Subquery formatters
type Subquery struct {
	Elements     []Formatter
	IndentLevel  int
	*Options     // Options used later to format element
	IsColumnArea bool
	ColumnCount  int
}

// Format reindents and formats elements accordingly
func (formatter *Subquery) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

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

	// Figure out whether the surrounding brackets of this subquery need to start in the
	// same line or be shifted to a new line.
	newLine := true
	if previousParentToken.ContinueLine() {
		newLine = false
	} else if formatter.IsColumnArea {
		newLine = true         // In column area, every value is put in a new line
		formatter.AddIndent(1) // In column area, every value is indented by one
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			formatter.writeSubquery(buf, token, previousParentToken, formatter.IndentLevel, formatter.ColumnCount, newLine)
		} else {

			// Parenthesis used first indent, increment indent of content again.
			el.AddIndent(1)

			// Format content
			_ = el.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Subquery) AddIndent(lev int) {
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

func (formatter *Subquery) writeSubquery(buf *bytes.Buffer, token, previousParentToken Token, indent, columnCount int, newLine bool) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	switch {

	case previousParentToken.ContinueLine() && token.Type == lexer.STARTPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	case previousParentToken.ContinueLine() && token.Type == lexer.ENDPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value)) // One whitespace was already written by select column

	// Put start parenthesis into same line, if desired
	case !newLine && token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))

	// Put start and end parenthesis each in a new line with standard indent
	case token.Type == lexer.STARTPARENTHESIS || token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))

	// Token values
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", strings.Repeat(INDENT, indent), token.Value))
	}
}
