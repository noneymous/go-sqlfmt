package formatters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentDelete(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Formatter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Formatter{
				Token{Options: options, Token: lexer.Token{Type: lexer.DELETE, Value: "DELETE"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"}},
			},
			want: "DELETE FROM xxxxxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Delete{Options: options, Elements: tt.tokenSource}

			_ = el.Format(buf, nil, 0)
			got := buf.String()
			if tt.want != got {
				t.Errorf("\n=======================\n=== WANT =============>\n%s\n=======================\n=== GOT ==============>\n%s\n=======================", tt.want, got)
			} else {
				fmt.Println(fmt.Sprintf("%s\n%s", got, "========================================================================"))
			}
		})
	}
}
