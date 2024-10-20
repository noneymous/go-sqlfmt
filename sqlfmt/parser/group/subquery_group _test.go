package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentSubqueryGroup(t *testing.T) {
	tests := []struct {
		name string
		src  []lexer.Reindenter
		want string
	}{
		{
			name: "normalcase",
			src: []lexer.Reindenter{
				lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("},
				&Select{
					Element: []lexer.Reindenter{
						lexer.Token{Type: lexer.SELECT, Value: "SELECT"},
						lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"},
					},
					IndentLevel: 1,
				},
				&From{
					Element: []lexer.Reindenter{
						lexer.Token{Type: lexer.FROM, Value: "FROM"},
						lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"},
					},
					IndentLevel: 1,
				},
				lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"},
			},
			want: " (\n  SELECT\n    xxxxxx\n  FROM xxxxxx)",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Parenthesis{Element: tt.src, IndentLevel: 1}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
