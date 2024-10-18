package lexer

import (
	"bytes"
)

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

	QUOTEAREA
	SURROUNDING
	COLON
	DOUBLECOLON
)

// Token is a token struct
type Token struct {
	Type  TokenType
	Value string
}

// Reindent is a placeholder for implementing Reindenter interface
func (t Token) Reindent(buf *bytes.Buffer) error {
	return nil
}

// IncrementIndentLevel is a placeholder implementing Reindenter interface
func (t Token) IncrementIndentLevel(lev int) {

}

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

// IsNeedNewLineBefore returns true if token needs new line before written in buffer
func (t Token) IsNeedNewLineBefore() bool {
	var ttypes = []TokenType{SELECT, UPDATE, INSERT, DELETE, ANDGROUP, FROM, GROUP, ORGROUP, ORDER, HAVING, LIMIT, OFFSET, FETCH, RETURNING, SET, UNION, INTERSECT, EXCEPT, VALUES, WHERE, ON, USING, UNION, EXCEPT, INTERSECT}
	for _, v := range ttypes {
		if t.Type == v {
			return true
		}
	}
	return false
}

// IsKeyWordInSelect returns true if token is a keyword in select group
func (t Token) IsKeyWordInSelect() bool {
	return t.Type == SELECT || t.Type == EXISTS || t.Type == DISTINCT || t.Type == DISTINCTROW || t.Type == INTO || t.Type == AS || t.Type == GROUP || t.Type == ORDER || t.Type == BY || t.Type == ON || t.Type == RETURNING || t.Type == SET || t.Type == UPDATE
}
