package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentSubqueryGroup(t *testing.T) {
	tests := []struct {
		name         string
		isColumnArea bool
		src          []lexer.Reindenter
		want         string
	}{
		{
			name:         "normalcase column area",
			isColumnArea: true,
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
			want: "\n  (\n    SELECT\n      xxxxxx\n    FROM xxxxxx\n  )",
		},
		{
			name:         "normalcase outside column area",
			isColumnArea: false,
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
			want: " (\n  SELECT\n    xxxxxx\n  FROM xxxxxx\n)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Subquery{Element: tt.src, IndentLevel: 1}
			el.IsColumnArea = tt.isColumnArea

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
