package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentLockGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.LOCK, Value: "LOCK"},
				lexer.Token{Type: lexer.IDENT, Value: "table"},
				lexer.Token{Type: lexer.IN, Value: "IN"},
				lexer.Token{Type: lexer.IDENT, Value: "xxx"},
			},
			want: "\nLOCK table\nIN xxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Lock{Element: tt.tokenSource}

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
