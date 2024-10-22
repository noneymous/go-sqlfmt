package reindenters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Case group reindenter
type Case struct {
	Options        *lexer.Options // Options used later to format element
	Element        []lexer.Reindenter
	IndentLevel    int
	hasCommaBefore bool
}

// Reindent reindents its elements
func (group *Case) Reindent(buf *bytes.Buffer, parent []lexer.Reindenter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = group.Options.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(group.Element, WHITESPACE)
	if err != nil {
		return err
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	for i, el := range elements {

		// Write element or recursively call it's Reindent function
		if token, ok := el.(lexer.Token); ok {
			group.writeCase(buf, token, group.IndentLevel, group.hasCommaBefore)
		} else {
			_ = el.Reindent(buf, elements, i)
		}
	}

	// Return nil and continue with parent Reindenter
	return nil
}

// IncrementIndentLevel increments by its specified increment level
func (group *Case) IncrementIndentLevel(lev int) {
	group.IndentLevel += lev

	// Iterate and increase indent of child elements too
	for _, el := range group.Element {
		el.IncrementIndentLevel(lev)
	}
}

func (group *Case) writeCase(buf *bytes.Buffer, token lexer.Token, indent int, hasCommaBefore bool) {

	// Prepare short variables for better visibility
	var INDENT = group.Options.Indent
	var NEWLINE = group.Options.Newline
	var WHITESPACE = group.Options.Whitespace

	// Write element
	if hasCommaBefore {
		switch token.Type {
		case lexer.CASE:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		case lexer.WHEN, lexer.ELSE:
			buf.WriteString(fmt.Sprintf("%s%s%s%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, WHITESPACE, WHITESPACE, INDENT, token.Value))
		case lexer.END:
			buf.WriteString(fmt.Sprintf("%s%s%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, WHITESPACE, WHITESPACE, token.Value))
		case lexer.COMMA:
			buf.WriteString(fmt.Sprintf("%s", token.Value))
		default:
			if strings.HasPrefix(token.Value, "::") {
				buf.WriteString(fmt.Sprintf("%s", token.Value))
			} else {
				buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			}
		}
	} else {
		switch token.Type {
		case lexer.CASE, lexer.END:
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
		case lexer.WHEN, lexer.ELSE:
			buf.WriteString(fmt.Sprintf("%s%s%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, WHITESPACE, INDENT, token.Value))
		case lexer.COMMA:
			buf.WriteString(fmt.Sprintf("%s", token.Value))
		default:
			if strings.HasPrefix(token.Value, "::") {
				buf.WriteString(fmt.Sprintf("%s", token.Value))
			} else {
				buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			}
		}
	}
}
