package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentAndGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal test",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.ANDGROUP, Value: "AND"},
				lexer.Token{Type: lexer.IDENT, Value: "something1"},
				lexer.Token{Type: lexer.IDENT, Value: "something2"},
			},
			want: "\nAND something1 something2",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &AndGroup{Element: tt.tokenSource}

		if err := el.Reindent(buf, lexer.Token{}); err != nil {
			t.Errorf("error %#v", err)
		}
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
