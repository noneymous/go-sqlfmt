package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentCaseGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.CASE, Value: "CASE"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.WHEN, Value: "WHEN"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.END, Value: "END"}},
			},
			want: "\n  CASE\n    WHEN something something\n  END",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Case{Options: options, Element: tt.tokenSource}

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
