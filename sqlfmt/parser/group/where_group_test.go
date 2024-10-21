package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentWhereGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.WHERE, Value: "WHERE"},
				lexer.Token{Type: lexer.IDENT, Value: "something1"},
				lexer.Token{Type: lexer.IDENT, Value: "="},
				lexer.Token{Type: lexer.IDENT, Value: "something2"},
			},
			want: "\nWHERE something1 = something2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Where{Element: tt.tokenSource}

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
