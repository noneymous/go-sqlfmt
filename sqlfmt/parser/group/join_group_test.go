package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentJoinGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.LEFT, Value: "LEFT"},
				lexer.Token{Type: lexer.OUTER, Value: "OUTER"},
				lexer.Token{Type: lexer.JOIN, Value: "JOIN"},
				lexer.Token{Type: lexer.IDENT, Value: "sometable"},
				lexer.Token{Type: lexer.ON, Value: "ON"},
				lexer.Token{Type: lexer.IDENT, Value: "status1"},
				lexer.Token{Type: lexer.IDENT, Value: "="},
				lexer.Token{Type: lexer.IDENT, Value: "status2"},
			},

			want: "\nLEFT OUTER JOIN sometable ON status1 = status2",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Join{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
