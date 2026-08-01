package formatters

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Generic group formatter
// The generic group formatter can be applied if no special formatting rules are required
type Generic struct {
	Elements    []Formatter
	IndentLevel int
	*Options    // Options used later to format element
}

// Format component accordingly with necessary indents, newlines,...
func (formatter *Generic) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {

	// Prepare short variables for better visibility
	var WHITESPACE = formatter.Whitespace

	// Preprocess punctuation and enrich with surrounding information
	elements, err := processPunctuation(formatter.Elements, WHITESPACE)
	if err != nil {
		return err
	}

	// Determine the leading token of the preceding sibling segment. An EXPLAIN prefix is emitted as its own
	// Generic segment, so a wrapped UPDATE/DELETE statement (also a Generic segment) must be pushed onto a
	// new line to avoid gluing to the prefix, e.g. "EXPLAINUPDATE".
	var previousParentToken Token
	if parentIdx > 0 && parentIdx <= len(parent) {
		previousParentToken = firstToken(parent[parentIdx-1])
	}

	// Iterate and write elements to the buffer. Recursively step into nested elements.
	var previousToken Token
	for i, el := range elements {

		// Write element or recursively call its Format function
		if token, ok := el.(Token); ok {
			formatter.write(buf, token, previousToken, previousParentToken, formatter.IndentLevel, i)
		} else {

			// Set peripheral parameters to tell child elements to write to the same line
			switch v := el.(type) {
			case *Or:
				v.SameLine = true
			case *And:
				v.SameLine = true
			}

			// Recursively format nested elements
			_ = el.Format(buf, elements, i)
		}

		// Remember last Token element
		if token, ok := el.(Token); ok {
			previousToken = token
		} else {
			previousToken = Token{}
		}
	}

	// Return nil and continue with parent Formatter
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter *Generic) AddIndent(lev int) {
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

func (formatter *Generic) write(buf *bytes.Buffer, token, previousToken, previousParentToken Token, indent, position int) {

	// Prepare short variables for better visibility
	var INDENT = formatter.Indent
	var NEWLINE = formatter.Newline
	var WHITESPACE = formatter.Whitespace

	// Any token following a line comment must start on a new line
	if previousToken.IsLineComment() {
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		return
	}

	// Write element
	switch {

	// A Generic statement wrapped by an EXPLAIN prefix (e.g. "EXPLAIN UPDATE ...") is emitted as its own
	// segment following the EXPLAIN segment. Break its leading keyword onto a new line, mirroring how
	// SELECT/INSERT formatters newline their leading keyword, to avoid gluing to the prefix ("EXPLAINUPDATE").
	case position == 0 && previousParentToken.Type == lexer.EXPLAIN:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))

	case position == 0:
		buf.WriteString(fmt.Sprintf("%s", token.Value))

	// Within a COPY statement, move the TO target clause onto its own line, mirroring
	// how the FROM clause is broken out. This is scoped to COPY so other Generic
	// statements (e.g. ROLLBACK TO SAVEPOINT) keep TO on the same line.
	case token.Type == lexer.TO && formatter.isCopy():
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))

	// Write common token values
	default:
		buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
	}
}

// firstToken returns the leading Token of a Formatter segment, if determinable. It unwraps a Generic
// group to its first token element. An empty Token (Type 0) is returned when no leading token applies.
func firstToken(f Formatter) Token {
	switch v := f.(type) {
	case Token:
		return v
	case *Token:
		return *v
	case *Generic:
		if len(v.Elements) > 0 {
			if token, ok := v.Elements[0].(Token); ok {
				return token
			}
		}
	}
	return Token{}
}

// isCopy reports whether this Generic formatter represents a COPY statement,
// determined by its first token element.
func (formatter *Generic) isCopy() bool {
	if len(formatter.Elements) == 0 {
		return false
	}
	if token, ok := formatter.Elements[0].(Token); ok {
		return token.Type == lexer.COPY
	}
	return false
}
