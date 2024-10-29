package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Subquery reindenters
type Subquery struct {
	Options      *Options // Options used later to format element
	Element      []Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (group *Subquery) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {

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

	// Figure out whether the surrounding brackets of this subquery need to start in the
	// same line or be shifted to a new line.
	newLine := true
	if previousParentToken.ContinueLine() {
		newLine = false
	} else if group.IsColumnArea {
		newLine = true           // In column area, every value is put in a new line
		group.IncrementIndent(1) // In column area, every value is indented by one
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(Token); ok {
			group.writeSubquery(buf, token, previousParentToken, group.IndentLevel, group.ColumnCount, newLine)
		} else {

			// Parenthesis used first indent, increment indent of content again.
			el.IncrementIndent(1)

			// Reindent content
			_ = el.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndent increments by its specified indent level
func (group *Subquery) IncrementIndent(lev int) {
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

func (group *Subquery) writeSubquery(buf *bytes.Buffer, token, previousParentToken Token, indent, columnCount int, newLine bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	switch {

	case previousParentToken.ContinueLine() && token.Type == lexer.STARTPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s", group.Options.Whitespace, token.Value))
	case previousParentToken.ContinueLine() && token.Type == lexer.ENDPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent), token.Value)) // One whitespace was already written by select column

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
