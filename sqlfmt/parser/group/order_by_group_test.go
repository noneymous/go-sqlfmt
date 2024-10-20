package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentOrderByGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.ORDER, Value: "ORDER"},
				lexer.Token{Type: lexer.BY, Value: "BY"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"},
			},
			want: "\nORDER BY\n  xxxxxx",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &OrderBy{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
