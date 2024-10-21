package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentUnionGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case1",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.UNION, Value: "UNION"},
				lexer.Token{Type: lexer.ALL, Value: "ALL"},
			},
			want: "\nUNION ALL",
		},
		{
			name: "normal case2",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.INTERSECT, Value: "INTERSECT"},
			},
			want: "\nINTERSECT",
		},
		{
			name: "normal case3",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.EXCEPT, Value: "EXCEPT"},
			},
			want: "\nEXCEPT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &TieClause{Element: tt.tokenSource}

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
