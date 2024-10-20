package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentFromGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.FROM, Value: "FROM"},
				lexer.Token{Type: lexer.IDENT, Value: "sometable"},
			},
			want: "\nFROM sometable",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &From{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
