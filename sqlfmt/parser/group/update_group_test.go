package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentUpdateGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.UPDATE, Value: "UPDATE"},
				lexer.Token{Type: lexer.IDENT, Value: "something1"},
			},
			want: "\nUPDATE\n  something1",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Update{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
