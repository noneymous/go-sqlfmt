package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentInsertGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.INSERT, Value: "INSERT"},
				lexer.Token{Type: lexer.INTO, Value: "INTO"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxxx"},
			},
			want: "\nINSERT INTO xxxxxx xxxxxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Insert{Element: tt.tokenSource}

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
