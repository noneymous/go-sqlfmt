package formatters

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Options to define output format of Formatters
type Options struct {
	Padding    string // Character sequence added as left padding on all lines, e.g. "" (none)
	Indent     string // Character sequence used left indentation on indented clauses, e.g. "    " (4 spaces)
	Newline    string // Character sequence used as line feeds, e.g. "\n" (newline character)
	Whitespace string // Character sequence used as whitespace in SQL string, e.g. " " (single space)
}

// DefaultOptions returns a default options set for Formatters. Also used in unit tests.
func DefaultOptions() *Options {
	return &Options{
		Padding:    "",
		Indent:     "  ",
		Newline:    "\n",
		Whitespace: " ",
	}
}

// Formatter interface. Example values of Formatter would be clause group or token
type Formatter interface {
	Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error
	AddIndent(lev int)
}

// Token wrapping lexer Token but extended with Options struct
type Token struct {
	lexer.Token
	*Options
}

// Format reindents and formats elements accordingly
func (formatter Token) Format(buf *bytes.Buffer, parent []Formatter, parentIdx int) error {
	return nil
}

// AddIndent increments indentation level by the given amount
func (formatter Token) AddIndent(lev int) {

}

// IsTieClauseStart determines if token type is included in TokenTypesOfTieClause
func (formatter Token) IsTieClauseStart() bool {
	for _, v := range lexer.TokenTypesOfTieClause {
		if formatter.Type == v {
			return true
		}
	}
	return false
}

// IsLimitClauseStart determines token type is included in TokenTypesOfLimitClause
func (formatter Token) IsLimitClauseStart() bool {
	for _, v := range lexer.TokenTypesOfLimitClause {
		if formatter.Type == v {
			return true
		}
	}
	return false
}

// IsJoinStart determines if token type is included in TokenTypesOfJoinMaker
func (formatter Token) IsJoinStart() bool {
	for _, v := range lexer.TokenTypesOfJoinMaker {
		if formatter.Type == v {
			return true
		}
	}
	return false
}

// IsKeywordInSelect returns true if token is a keyword in select group
func (formatter Token) IsKeywordInSelect() bool {
	return formatter.Type == lexer.SELECT || formatter.Type == lexer.EXISTS || formatter.Type == lexer.DISTINCT || formatter.Type == lexer.DISTINCTROW || formatter.Type == lexer.INTO || formatter.Type == lexer.AS || formatter.Type == lexer.GROUP || formatter.Type == lexer.ORDER || formatter.Type == lexer.BY || formatter.Type == lexer.ON || formatter.Type == lexer.RETURNING || formatter.Type == lexer.SET || formatter.Type == lexer.UPDATE || formatter.Type == lexer.ANY
}

// RequiresNewline returns true if a token requires should be moved to a new line
func (formatter Token) RequiresNewline() bool {
	var ttypes = []lexer.TokenType{lexer.SELECT, lexer.UPDATE, lexer.INSERT, lexer.DELETE, lexer.ANDGROUP, lexer.FROM, lexer.GROUP, lexer.ORGROUP, lexer.ORDER, lexer.HAVING, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.RETURNING, lexer.SET, lexer.UNION, lexer.INTERSECT, lexer.EXCEPT, lexer.VALUES, lexer.WHERE, lexer.ON, lexer.USING, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT}
	for _, v := range ttypes {
		if formatter.Type == v {
			return true
		}
	}
	return false
}

// ContinueLine should be called on the last parent token to figure out, whether a token should continue in the same line
func (formatter Token) ContinueLine() bool {
	return formatter.IsComparator() || formatter.Type == lexer.FROM || formatter.Type == lexer.WHERE || formatter.Type == lexer.EXISTS || formatter.Type == lexer.AS || formatter.Type == lexer.IN || formatter.Type == lexer.ON || formatter.Type == lexer.ANY || formatter.Type == lexer.ARRAY
}

// IsComparator returns true if token is a comparator
func (formatter Token) IsComparator() bool {
	return formatter.Type == lexer.COMPARATOR
}

func write(
	buf *bytes.Buffer, INDENT, NEWLINE, WHITESPACE string, token Token, previousToken Token, indent int, hasMany bool) {
	switch {
	case token.RequiresNewline():
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	case token.Type == lexer.COMMA:
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case token.Type == lexer.DO:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, token.Value, WHITESPACE))
	case strings.HasPrefix(token.Value, "::"):
		buf.WriteString(fmt.Sprintf("%s", token.Value))
	case token.Type == lexer.WITH:
		buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
	default:
		if hasMany {

			// Use newlines as separators
			if previousToken.RequiresNewline() || previousToken.Type == lexer.AND || previousToken.Type == lexer.OR {
				buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, token.Value))
			} else {
				buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
			}
		} else {

			// Use whitespaces as separators
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		}
	}
}

func writeWithComma(buf *bytes.Buffer, INDENT, NEWLINE, WHITESPACE string, v interface{}, indent int, columnCount *int) error {
	if token, ok := v.(Token); ok {
		switch {
		case token.RequiresNewline():
			buf.WriteString(fmt.Sprintf("%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), token.Value))
		case token.Type == lexer.BY:
			buf.WriteString(fmt.Sprintf("%s%s", WHITESPACE, token.Value))
		case token.Type == lexer.COMMA:
			buf.WriteString(fmt.Sprintf("%s%s%s%s", token.Value, NEWLINE, strings.Repeat(INDENT, indent), INDENT))
		default:
			return fmt.Errorf("can not reindent %#v", token.Value)
		}
	} else if str, isString := v.(string); isString {
		str = strings.TrimRight(str, " ")
		if *(columnCount) == 0 {
			buf.WriteString(fmt.Sprintf("%s%s%s%s", NEWLINE, strings.Repeat(INDENT, indent), INDENT, str))
		} else if strings.HasPrefix(token.Value, "::") {
			buf.WriteString(fmt.Sprintf("%s", str))
		} else {
			buf.WriteString(fmt.Sprintf("%s", str))
		}

		// Update number at pointer address
		*columnCount = *(columnCount) + 1
	}
	return nil
}

// separate elements by comma and the reserved word in select clause
func separate(rs []Formatter, WHITESPACE string) []interface{} {
	var (
		result           []interface{}
		skipRange, count int
	)
	buf := &bytes.Buffer{}

	for _, r := range rs {
		if token, ok := r.(Token); !ok {
			if buf.String() != "" {
				result = append(result, buf.String())
				buf.Reset()
			}
			result = append(result, r)
		} else {
			switch {
			case skipRange > 0:
				skipRange--
				// TODO: more elegant
			case token.IsKeywordInSelect():
				if buf.String() != "" {
					result = append(result, buf.String())
					buf.Reset()
				}
				result = append(result, token)
			case token.Type == lexer.COMMA:
				if buf.String() != "" {
					result = append(result, buf.String())
				}
				result = append(result, token)
				buf.Reset()
				count = 0
			case strings.HasPrefix(token.Value, "::"):
				buf.WriteString(token.Value)
			default:
				if count == 0 {
					buf.WriteString(token.Value)
				} else {
					buf.WriteString(WHITESPACE + token.Value)
				}
				count++
			}
		}
	}
	// append the last element in buf
	if buf.String() != "" {
		result = append(result, buf.String())
	}
	return result
}

// process bracket, single quote and brace
// TODO: more elegant
func processPunctuation(rs []Formatter, WHITESPACE string) ([]Formatter, error) {
	var (
		result    []Formatter
		skipRange int
	)

	for i, v := range rs {
		if token, ok := v.(Token); ok {
			switch {
			case skipRange > 0:
				skipRange--
			case token.Type == lexer.STARTBRACE || token.Type == lexer.STARTBRACKET:
				surrounding, sr, err := extractSurroundingArea(rs[i:], WHITESPACE)
				if err != nil {
					return nil, err
				}
				result = append(result, Token{
					Options: token.Options,
					Token: lexer.Token{
						Type:  lexer.SURROUNDING,
						Value: surrounding,
					},
				})
				skipRange += sr
			default:
				result = append(result, token)
			}
		} else {
			result = append(result, v)
		}
	}
	return result, nil
}

// returns surrounding area including punctuation such as {xxx, xxx}
func extractSurroundingArea(rs []Formatter, WHITESPACE string) (string, int, error) {
	var (
		countOfStart int
		countOfEnd   int
		result       string
		skipRange    int
	)
	for i, r := range rs {
		if token, ok := r.(Token); ok {
			switch {
			case token.Type == lexer.COMMA || token.Type == lexer.STARTBRACKET || token.Type == lexer.STARTBRACE || token.Type == lexer.ENDBRACKET || token.Type == lexer.ENDBRACE:
				result += fmt.Sprint(token.Value)
				// for next token of StartToken
			case i == 1:
				result += fmt.Sprint(token.Value)
			default:
				result += fmt.Sprint(WHITESPACE + token.Value)
			}

			if token.Type == lexer.STARTBRACKET || token.Type == lexer.STARTBRACE || token.Type == lexer.STARTPARENTHESIS {
				countOfStart++
			}
			if token.Type == lexer.ENDBRACKET || token.Type == lexer.ENDBRACE || token.Type == lexer.ENDPARENTHESIS {
				countOfEnd++
			}
			if countOfStart == countOfEnd {
				break
			}
			skipRange++
		} else {
			// TODO: should support group type in surrounding area?
			// I have not encountered any groups in surrounding area so far
			return "", -1, errors.New("group type is not supposed be here")
		}
	}
	return result, skipRange, nil
}
