package lexer

import (
	"bytes"
)

// Options for go-sqlfmt
type Options struct {
	Padding    string // Character sequence added as left padding on all lines, e.g. "" (none)
	Indent     string // Character sequence used left indentation on indented clauses, e.g. "    " (4 spaces)
	Newline    string // Character sequence used as line feeds, e.g. "\n" (newline character)
	Whitespace string // Character sequence used as whitespace in SQL string, e.g. " " (single space)
}

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
	IncrementIndentLevel(lev int)
}

// Token is a token struct
type Token struct {
	Options *Options // Options used later to format element
	Type    TokenType
	Value   string
}

// Reindent is a placeholder for implementing Reindenter interface
func (t Token) Reindent(buf *bytes.Buffer, parent []Reindenter, parentIdx int) error {
	return nil
}

// IncrementIndentLevel is a placeholder implementing Reindenter interface
func (t Token) IncrementIndentLevel(lev int) {

}

// TokenType is an alias type that represents a kind of token
//
//go:generate stringer -type=TokenType
type TokenType int

// Token types
const (
	EOF TokenType = 1 + iota // eof
	WHITESPACE
	NEWLINE
	TAB
	FUNCTION
	COMMA
	STARTPARENTHESIS
	ENDPARENTHESIS
	STARTBRACKET
	ENDBRACKET
	STARTBRACE
	ENDBRACE
	TYPE
	IDENT  // field or table name
	STRING // values surrounded with single quotes
	SELECT
	FROM
	WHERE
	CASE
	ORDER
	BY
	AS
	JOIN
	LEFT
	RIGHT
	INNER
	OUTER
	ON
	WHEN
	END
	GROUP
	DESC
	ASC
	LIMIT
	AND
	ANDGROUP
	OR
	ORGROUP
	IN
	ANY
	IS
	NOT
	NULL
	DISTINCT
	LIKE
	ILIKE
	BETWEEN
	UNION
	ALL
	HAVING
	OVER
	EXISTS
	UPDATE
	SET
	RETURNING
	DELETE
	INSERT
	INTO
	DO
	VALUES
	FOR
	THEN
	ELSE
	DISTINCTROW
	FILTER
	WITHIN
	COLLATE
	INTERVAL
	INTERSECT
	EXCEPT
	OFFSET
	FETCH
	FIRST
	ROWS
	USING
	OVERLAPS
	NATURAL
	CROSS
	TIME
	ZONE
	NULLS
	LAST
	AT
	LOCK
	WITH

	/*
	*	Custom types used by sqlfmt for intermediate representations
	 */
	QUOTEAREA
	SURROUNDING
	COLON
	DOUBLECOLON
)

// End keywords of each clause
var (
	EndOfSelect      = []TokenType{FROM, UNION, ENDPARENTHESIS, EOF}
	EndOfCase        = []TokenType{END, EOF}
	EndOfFrom        = []TokenType{WHERE, INNER, OUTER, LEFT, RIGHT, JOIN, NATURAL, CROSS, ORDER, GROUP, UNION, OFFSET, LIMIT, FETCH, EXCEPT, INTERSECT, ENDPARENTHESIS, EOF}
	EndOfJoin        = []TokenType{WHERE, ORDER, GROUP, LIMIT, OFFSET, FETCH, ANDGROUP, ORGROUP, LEFT, RIGHT, INNER, OUTER, NATURAL, CROSS, UNION, EXCEPT, INTERSECT, ENDPARENTHESIS, EOF}
	EndOfWhere       = []TokenType{GROUP, ORDER, LIMIT, OFFSET, FETCH, ANDGROUP, OR, UNION, EXCEPT, INTERSECT, RETURNING, ENDPARENTHESIS, EOF}
	EndOfAndGroup    = []TokenType{GROUP, ORDER, LIMIT, OFFSET, FETCH, UNION, EXCEPT, INTERSECT, ANDGROUP, ORGROUP, ENDPARENTHESIS, EOF}
	EndOfOrGroup     = []TokenType{GROUP, ORDER, LIMIT, OFFSET, FETCH, UNION, EXCEPT, INTERSECT, ANDGROUP, ORGROUP, ENDPARENTHESIS, EOF}
	EndOfGroupBy     = []TokenType{ORDER, LIMIT, FETCH, OFFSET, UNION, EXCEPT, INTERSECT, HAVING, ENDPARENTHESIS, EOF}
	EndOfHaving      = []TokenType{LIMIT, OFFSET, FETCH, ORDER, UNION, EXCEPT, INTERSECT, ENDPARENTHESIS, EOF}
	EndOfOrderBy     = []TokenType{LIMIT, FETCH, OFFSET, UNION, EXCEPT, INTERSECT, ENDPARENTHESIS, EOF}
	EndOfLimitClause = []TokenType{UNION, EXCEPT, INTERSECT, ENDPARENTHESIS, EOF}
	EndOfParenthesis = []TokenType{ENDPARENTHESIS, EOF}
	// 微妙
	EndOfTieClause = []TokenType{SELECT, EOF}
	EndOfUpdate    = []TokenType{WHERE, SET, RETURNING, EOF}
	EndOfSet       = []TokenType{WHERE, RETURNING, EOF}
	EndOfReturning = []TokenType{EOF}
	EndOfDelete    = []TokenType{WHERE, FROM, EOF}
	EndOfInsert    = []TokenType{VALUES, EOF}
	EndOfValues    = []TokenType{UPDATE, RETURNING, EOF}
	EndOfFunction  = []TokenType{ENDPARENTHESIS, EOF}
	EndOfTypeCast  = []TokenType{ENDPARENTHESIS, EOF}
	EndOfLock      = []TokenType{EOF}
	EndOfWith      = []TokenType{EOF}
)

// Token types that contain the keyword to make subGroup
var (
	TokenTypesOfGroupMaker = []TokenType{SELECT, CASE, FROM, WHERE, ORDER, GROUP, LIMIT, ANDGROUP, ORGROUP, HAVING, UNION, EXCEPT, INTERSECT, FUNCTION, STARTPARENTHESIS, TYPE}
	TokenTypesOfJoinMaker  = []TokenType{JOIN, INNER, OUTER, LEFT, RIGHT, NATURAL, CROSS}
	TokenTypeOfTieClause   = []TokenType{UNION, INTERSECT, EXCEPT}
	TokenTypeOfLimitClause = []TokenType{LIMIT, FETCH, OFFSET}
)

// IsJoinStart determines if token type is included in TokenTypesOfJoinMaker
func (t Token) IsJoinStart() bool {
	for _, v := range TokenTypesOfJoinMaker {
		if t.Type == v {
			return true
		}
	}
	return false
}

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

// IsNewlineKeyword returns true if token needs new line before written in buffer
func (t Token) IsNewlineKeyword() bool {
	var ttypes = []TokenType{SELECT, UPDATE, INSERT, DELETE, ANDGROUP, FROM, GROUP, ORGROUP, ORDER, HAVING, LIMIT, OFFSET, FETCH, RETURNING, SET, UNION, INTERSECT, EXCEPT, VALUES, WHERE, ON, USING, UNION, EXCEPT, INTERSECT}
	for _, v := range ttypes {
		if t.Type == v {
			return true
		}
	}
	return false
}

// IsKeywordInSelect returns true if token is a keyword in select group
func (t Token) IsKeywordInSelect() bool {
	return t.Type == SELECT || t.Type == EXISTS || t.Type == DISTINCT || t.Type == DISTINCTROW || t.Type == INTO || t.Type == AS || t.Type == GROUP || t.Type == ORDER || t.Type == BY || t.Type == ON || t.Type == RETURNING || t.Type == SET || t.Type == UPDATE || t.Type == ANY
}

// IsKeywordWithoutLinebreak returns true if token doesn't require newline for subsequent sub query
func (t Token) IsKeywordWithoutLinebreak() bool {
	return t.Type == FROM || t.Type == WHERE || t.Type == EXISTS || t.Type == IN || t.Type == ANY || t.Type == AS
}

func (t Token) IsIdentWithoutLinebreak() bool {
	return t.Type == IDENT && (t.Value == "=" ||
		t.Value == ">" ||
		t.Value == "<" ||
		t.Value == ">=" ||
		t.Value == "<=" ||
		t.Value == "<>")
}
