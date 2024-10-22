package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentLimitGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.LIMIT, Value: "LIMIT"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "123"},
			},
			want: "\nLIMIT 123",
		},
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.OFFSET, Value: "OFFSET"},
			},
			want: "\nOFFSET",
		},
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.FETCH, Value: "FETCH"},
				lexer.Token{Options: options, Type: lexer.FIRST, Value: "FIRST"},
			},
			want: "\nFETCH FIRST",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Limit{Options: options, Element: tt.tokenSource}

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
