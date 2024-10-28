package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentAndGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Reindenter
		want        string
	}{
		{
			name: "normal test",
			tokenSource: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.ANDGROUP, Value: "AND"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something1"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something2"}},
			},
			want: "\nAND something1 something2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &AndGroup{Options: options, Element: tt.tokenSource}

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
