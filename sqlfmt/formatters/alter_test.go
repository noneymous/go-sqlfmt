package formatters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestFormatAlter(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Formatter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Formatter{
				Token{Options: options, Token: lexer.Token{Type: lexer.ALTER, Value: "ALTER"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.TABLE, Value: "TABLE"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "something1"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.SET, Value: "SET"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "val"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "="}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "1"}},
			},
			want: "ALTER TABLE something1 SET val = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Update{Options: options, Elements: tt.tokenSource}

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
