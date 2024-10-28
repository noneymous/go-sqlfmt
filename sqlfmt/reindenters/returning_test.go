package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentReturningGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.RETURNING, Value: "RETURNING"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something1"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.COMMA, Value: ","}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something1"}},
			},
			want: "\nRETURNING\n  something1,\n  something1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Returning{Options: options, Element: tt.tokenSource}

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
