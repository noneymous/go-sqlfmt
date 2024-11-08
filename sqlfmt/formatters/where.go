package formatters

import (
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"strings"
)

const maxWhereClausesPerLine = 2

// Where group formatter
type Where struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format reindents and formats elements accordingly
func (formatter *Where) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Check how many clauses there are. Linebreak if too many
	var clauses = 1 // WHERE clause starts with first clause
	for _, el := range elements {
		switch el.(type) {
		case *And:
			clauses++
		case *Or:
			clauses++
		}
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var hasMany = clauses > maxWhereClausesPerLine
	for i, el := range elements {

		// Write element or recursively call it's Format function
		if token, ok := el.(Token); ok {
			writeWhere(buf, INDENT, NEWLINE, WHITESPACE, token, formatter.IndentLevel, i, hasMany)
		} else {

			// Set peripheral parameters to tell child elements to write to the same line
			if !hasMany {
				switch v := el.(type) {
				case *Or:
					v.SameLine = true
				case *And:
					v.SameLine = true
				}
			}

			// Increment indent, if WHERE clauses should be written into new lines
			if hasMany {
				el.AddIndent(1)
			}

			// Recursively format nested elements
			_ = el.Format(buf, elements, i)
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Where) AddIndent(lev int) {
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

func writeWhere(
	buf *bytes.Buffer,
	INDENT,
	NEWLINE,
	WHITESPACE string,
	token Token,
	indent,
	position int,
	hasMany bool,
) {
	// Print WHERE token into new line
	if token.ContinueNewline() {
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		return
	}

	// Move each clause into a new line if there are many causes
	if hasMany {
		switch {
		case position == 1: // First element of first clause
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
			return

		case token.Type == lexer.AND || token.Type == lexer.OR: // Any additional where clause introduced by AND / OR
			buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
			return
		}
	}

	// Print to same line with WHITESPACE
	switch {
	case strings.HasPrefix(token.Value, "::"): // Write cast token without whitespace
		buf.WriteString(fmt.Sprintf("%s", token.Value))
		return
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		return
	}
}
