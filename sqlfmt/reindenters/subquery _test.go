package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentSubqueryGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name         string
		isColumnArea bool
		src          []Reindenter
		want         string
	}{
		{
			name:         "normalcase column area",
			isColumnArea: true,
			src: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
				&Select{
					Options: options,
					Element: []Reindenter{
						Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
						Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"}},
					},
					IndentLevel: 1,
				},
				&From{
					Options: options,
					Element: []Reindenter{
						Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
						Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"}},
					},
					IndentLevel: 1,
				},
				Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
			},
			want: "\n  (\n    SELECT\n      xxxxxx\n    FROM xxxxxx\n  )",
		},
		{
			name:         "normalcase outside column area",
			isColumnArea: false,
			src: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
				&Select{
					Options: options,
					Element: []Reindenter{
						Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
						Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"}},
					},
					IndentLevel: 1,
				},
				&From{
					Options: options,
					Element: []Reindenter{
						Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
						Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"}},
					},
					IndentLevel: 1,
				},
				Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
			},
			want: "\n  (\n    SELECT\n      xxxxxx\n    FROM xxxxxx\n  )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Subquery{Options: options, Element: tt.src, IndentLevel: 1}
			el.IsColumnArea = tt.isColumnArea

			_ = el.Reindent(buf, nil, 0)
			got := buf.String()
			if tt.want != got {
				t.Errorf("\n=======================\n=== WANT =============>\n%s\n=======================\n=== GOT ==============>\n%s\n=======================", tt.want, got)
			} else {
				fmt.Println(fmt.Sprintf("%s\n%s", got, "========================================================================"))
			}
		})
	}
}
