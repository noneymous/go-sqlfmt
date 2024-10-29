package formatters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentValuesGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Formatter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Formatter{
				Token{Options: options, Token: lexer.Token{Type: lexer.VALUES, Value: "VALUES"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxx"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.ON, Value: "ON"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxxxx"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.DO, Value: "DO"}},
			},
			want: "\nVALUES xxxxx\nON xxxxx\nDO ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Values{Options: options, Elements: tt.tokenSource}

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
