package lexer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	var testingSQLStatement = strings.Trim(`select name, age,sum, sum(case xxx) from user where name xxx and age = 'xxx' limit 100 except 100`, "`")
	want := []Token{
		{Type: SELECT, Value: "SELECT"},
		{Type: IDENT, Value: "name"},
		{Type: COMMA, Value: ","},
		{Type: IDENT, Value: "age"},
		{Type: COMMA, Value: ","},
		{Type: IDENT, Value: "sum"},
		{Type: COMMA, Value: ","},
		{Type: FUNCTION, Value: "SUM"},
		{Type: STARTPARENTHESIS, Value: "("},
		{Type: CASE, Value: "CASE"},
		{Type: IDENT, Value: "xxx"},
		{Type: ENDPARENTHESIS, Value: ")"},
		{Type: FROM, Value: "FROM"},
		{Type: IDENT, Value: "user"},
		{Type: WHERE, Value: "WHERE"},
		{Type: IDENT, Value: "name"},
		{Type: IDENT, Value: "xxx"},
		{Type: AND, Value: "AND"},
		{Type: IDENT, Value: "age"},
		{Type: IDENT, Value: "="},
		{Type: STRING, Value: "'xxx'"},
		{Type: LIMIT, Value: "LIMIT"},
		{Type: IDENT, Value: "100"},
		{Type: EXCEPT, Value: "EXCEPT"},
		{Type: IDENT, Value: "100"},
		{Type: EOF, Value: "EOF"},
	}
	got, err := Tokenize(testingSQLStatement)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
