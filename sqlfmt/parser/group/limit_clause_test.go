package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentLimitGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.LIMIT, Value: "LIMIT"},
				lexer.Token{Type: lexer.IDENT, Value: "123"},
			},
			want: "\nLIMIT 123",
		},
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.OFFSET, Value: "OFFSET"},
			},
			want: "\nOFFSET",
		},
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.FETCH, Value: "FETCH"},
				lexer.Token{Type: lexer.FIRST, Value: "FIRST"},
			},
			want: "\nFETCH FIRST",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &LimitClause{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
