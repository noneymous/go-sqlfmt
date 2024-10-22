package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Subquery reindenters
type Subquery struct {
	Options      *lexer.Options // Options used later to format element
	Element      []lexer.Reindenter
	IndentLevel  int
	IsColumnArea bool
	ColumnCount  int
}

// Reindent reindents its elements
func (group *Subquery) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Get last token written by parent
	var previousParentToken lexer.Token
	if len(parent) > parentIdx {
		if token, ok := parent[parentIdx-1].(lexer.Token); ok {
			previousParentToken = token
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(lexer.Token); ok {
			group.writeSubquery(buf, token, previousParentToken, group.IndentLevel, group.ColumnCount, group.IsColumnArea)
		} else {
			if !previousParentToken.IsKeywordWithoutLinebreak() && !previousParentToken.IsIdentWithoutLinebreak() {
				el.IncrementIndentLevel(1)
			}
			_ = el.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified indent level
func (group *Subquery) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Subquery) writeSubquery(buf *bytes.Buffer, token, previousParentToken lexer.Token, indent, columnCount int, isColumnArea bool) {

	switch {

	// Open parenthesis in same line if compatible with previous keyword (FROM, WHERE, EXISTS,...)
	case previousParentToken.IsKeywordWithoutLinebreak() || previousParentToken.IsIdentWithoutLinebreak():
		if token.Type == lexer.STARTPARENTHESIS {
			buf.WriteString(fmt.Sprintf("%s%s", group.Options.Whitespace, token.Value))
		} else if token.Type == lexer.ENDPARENTHESIS {
			buf.WriteString(fmt.Sprintf("%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent-1), token.Value))
		}

	case previousParentToken.Type == lexer.AS && token.Type == lexer.STARTPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s", strings.Repeat(group.Options.Indent, indent-1), group.Options.Whitespace, token.Value)) // One whitespace was already written by select column
	case previousParentToken.Type == lexer.AS && token.Type == lexer.ENDPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent-1), token.Value)) // One whitespace was already written by select column

	// Select columns do already end with newline, no additional one needed
	case isColumnArea && token.Type == lexer.STARTPARENTHESIS && columnCount == 0:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent-1), group.Options.Indent, token.Value)) // One whitespace was already written by select column
	case isColumnArea && token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent-1), group.Options.Indent, token.Value)) // One whitespace was already written by select column

	// Standard sub query moved to new line with extra indent
	case token.Type == lexer.STARTPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent), token.Value))
	case token.Type == lexer.ENDPARENTHESIS:
		buf.WriteString(fmt.Sprintf("%s%s%s", group.Options.Newline, strings.Repeat(group.Options.Indent, indent), token.Value))

	// Token values
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	default:
		buf.WriteString(fmt.Sprintf("%s%s", strings.Repeat(group.Options.Indent, indent), token.Value))
	}
}
