package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentHavingGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.HAVING, Value: "HAVING"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxxxxx"},
			},
			want: "\nHAVING xxxxxxxx",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Having{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
