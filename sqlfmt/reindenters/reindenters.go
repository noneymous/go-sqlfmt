package reindenters

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// Options to define output format of Reindenters
type Options struct {
	Padding    string // Character sequence added as left padding on all lines, e.g. "" (none)
	Indent     string // Character sequence used left indentation on indented clauses, e.g. "    " (4 spaces)
	Newline    string // Character sequence used as line feeds, e.g. "\n" (newline character)
	Whitespace string // Character sequence used as whitespace in SQL string, e.g. " " (single space)
}

// DefaultOptions returns a default options set for Reindenters. Also used in unit tests.
func DefaultOptions() *Options {
	return &Options{
		Padding:    "",
		Indent:     "  ",
		Newline:    "\n",
		Whitespace: " ",
	}
}

// Reindenter interface. Example values of Reindenter would be clause group or token
type Reindenter interface {
	Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error
	IncrementIndent(lev int)
}

// Token wrapping lexer Token but extended with Options struct
type Token struct {
	Options *Options
	lexer.Token
}

// Reindent is a placeholder for implementing Reindenter interface
func (t Token) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {
	return nil
}

// IncrementIndent is a placeholder implementing Reindenter interface
func (t Token) IncrementIndent(lev int) {

}

// Define end keywords for each clause segment
var (
	EndOfSelect          = []lexer.TokenType{lexer.FROM, lexer.UNION, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfCase            = []lexer.TokenType{lexer.END, lexer.EOF}
	EndOfFrom            = []lexer.TokenType{lexer.WHERE, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.JOIN, lexer.NATURAL, lexer.CROSS, lexer.ORDER, lexer.GROUP, lexer.UNION, lexer.OFFSET, lexer.LIMIT, lexer.FETCH, lexer.EXCEPT, lexer.INTERSECT, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfJoin            = []lexer.TokenType{lexer.WHERE, lexer.ORDER, lexer.GROUP, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.ANDGROUP, lexer.ORGROUP, lexer.LEFT, lexer.RIGHT, lexer.INNER, lexer.OUTER, lexer.NATURAL, lexer.CROSS, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfWhere           = []lexer.TokenType{lexer.GROUP, lexer.ORDER, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.ANDGROUP, lexer.OR, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.RETURNING, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfAndGroup        = []lexer.TokenType{lexer.GROUP, lexer.ORDER, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ANDGROUP, lexer.ORGROUP, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfOrGroup         = []lexer.TokenType{lexer.GROUP, lexer.ORDER, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ANDGROUP, lexer.ORGROUP, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfGroupBy         = []lexer.TokenType{lexer.ORDER, lexer.LIMIT, lexer.FETCH, lexer.OFFSET, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.HAVING, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfHaving          = []lexer.TokenType{lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.ORDER, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfOrderBy         = []lexer.TokenType{lexer.LIMIT, lexer.FETCH, lexer.OFFSET, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfLimitClause     = []lexer.TokenType{lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfParenthesis     = []lexer.TokenType{lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfTieClause       = []lexer.TokenType{lexer.SELECT, lexer.EOF}
	EndOfUpdate          = []lexer.TokenType{lexer.WHERE, lexer.SET, lexer.RETURNING, lexer.EOF}
	EndOfSet             = []lexer.TokenType{lexer.WHERE, lexer.RETURNING, lexer.EOF}
	EndOfReturning       = []lexer.TokenType{lexer.EOF}
	EndOfDelete          = []lexer.TokenType{lexer.WHERE, lexer.FROM, lexer.EOF}
	EndOfInsert          = []lexer.TokenType{lexer.VALUES, lexer.EOF}
	EndOfValues          = []lexer.TokenType{lexer.UPDATE, lexer.RETURNING, lexer.EOF}
	EndOfTypeCast        = []lexer.TokenType{lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfLock            = []lexer.TokenType{lexer.EOF}
	EndOfWith            = []lexer.TokenType{lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfFunction        = []lexer.TokenType{lexer.ENDPARENTHESIS, lexer.EOF}
	EndOfFunctionKeyword []lexer.TokenType // No end types means everything is an end type
)

// Define keywords indicating certain segment groups
var (
	TokenTypesOfGroupMaker = []lexer.TokenType{lexer.SELECT, lexer.CASE, lexer.FROM, lexer.WHERE, lexer.ORDER, lexer.GROUP, lexer.LIMIT, lexer.ANDGROUP, lexer.ORGROUP, lexer.HAVING, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT, lexer.FUNCTION, lexer.STARTPARENTHESIS, lexer.TYPE, lexer.WITH}
	TokenTypesOfJoinMaker  = []lexer.TokenType{lexer.JOIN, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.NATURAL, lexer.CROSS}
	TokenTypeOfTieClause   = []lexer.TokenType{lexer.UNION, lexer.INTERSECT, lexer.EXCEPT}
	TokenTypeOfLimitClause = []lexer.TokenType{lexer.LIMIT, lexer.FETCH, lexer.OFFSET}
)

// IsTieClauseStart determines if token type is included in TokenTypesOfTieClause
func (t Token) IsTieClauseStart() bool {
	for _, v := range TokenTypeOfTieClause {
		if t.Type == v {
			return true
		}
	}
	return false
}

// IsLimitClauseStart determines token type is included in TokenTypesOfLimitClause
func (t Token) IsLimitClauseStart() bool {
	for _, v := range TokenTypeOfLimitClause {
		if t.Type == v {
			return true
		}
	}
	return false
}

// IsJoinStart determines if token type is included in TokenTypesOfJoinMaker
func (t Token) IsJoinStart() bool {
	for _, v := range TokenTypesOfJoinMaker {
		if t.Type == v {
			return true
		}
	}
	return false
}

// IsKeywordInSelect returns true if token is a keyword in select group
func (t Token) IsKeywordInSelect() bool {
	return t.Type == lexer.SELECT || t.Type == lexer.EXISTS || t.Type == lexer.DISTINCT || t.Type == lexer.DISTINCTROW || t.Type == lexer.INTO || t.Type == lexer.AS || t.Type == lexer.GROUP || t.Type == lexer.ORDER || t.Type == lexer.BY || t.Type == lexer.ON || t.Type == lexer.RETURNING || t.Type == lexer.SET || t.Type == lexer.UPDATE || t.Type == lexer.ANY
}

// RequiresNewline returns true if a token requires should be moved to a new line
func (t Token) RequiresNewline() bool {
	var ttypes = []lexer.TokenType{lexer.SELECT, lexer.UPDATE, lexer.INSERT, lexer.DELETE, lexer.ANDGROUP, lexer.FROM, lexer.GROUP, lexer.ORGROUP, lexer.ORDER, lexer.HAVING, lexer.LIMIT, lexer.OFFSET, lexer.FETCH, lexer.RETURNING, lexer.SET, lexer.UNION, lexer.INTERSECT, lexer.EXCEPT, lexer.VALUES, lexer.WHERE, lexer.ON, lexer.USING, lexer.UNION, lexer.EXCEPT, lexer.INTERSECT}
	for _, v := range ttypes {
		if t.Type == v {
			return true
		}
	}
	return false
}

// ContinueLine should be called on the last parent token to figure out, whether a token should continue in the same line
func (t Token) ContinueLine() bool {
	return t.IsComparator() || t.Type == lexer.FROM || t.Type == lexer.WHERE || t.Type == lexer.EXISTS || t.Type == lexer.AS || t.Type == lexer.IN || t.Type == lexer.ON || t.Type == lexer.ANY || t.Type == lexer.ARRAY
}

// IsComparator returns true if token is a comparator
func (t Token) IsComparator() bool {
	return t.Type == lexer.COMPARATOR
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
func separate(rs []Reindenter, WHITESPACE string) []interface{} {
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

// process bracket, singlequote and brace
// TODO: more elegant
func processPunctuation(rs []Reindenter, WHITESPACE string) ([]Reindenter, error) {
	var (
		result    []Reindenter
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
func extractSurroundingArea(rs []Reindenter, WHITESPACE string) (string, int, error) {
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
