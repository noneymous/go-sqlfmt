package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentOrGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.ORGROUP, Value: "OR"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "something1"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "something2"},
			},
			want: "\nOR something1 something2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &OrGroup{Options: options, Element: tt.tokenSource}

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
