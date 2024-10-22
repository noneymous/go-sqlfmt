package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentUnionGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case1",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.UNION, Value: "UNION"},
				lexer.Token{Options: options, Type: lexer.ALL, Value: "ALL"},
			},
			want: "\nUNION ALL",
		},
		{
			name: "normal case2",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.INTERSECT, Value: "INTERSECT"},
			},
			want: "\nINTERSECT",
		},
		{
			name: "normal case3",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.EXCEPT, Value: "EXCEPT"},
			},
			want: "\nEXCEPT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &TieGroup{Options: options, Element: tt.tokenSource}

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
