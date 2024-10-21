package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentValuesGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.VALUES, Value: "VALUES"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxx"},
				lexer.Token{Type: lexer.ON, Value: "ON"},
				lexer.Token{Type: lexer.IDENT, Value: "xxxxx"},
				lexer.Token{Type: lexer.DO, Value: "DO"},
			},
			want: "\nVALUES xxxxx\nON xxxxx\nDO ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Values{Element: tt.tokenSource}

			_ = el.Reindent(buf, lexer.Token{})
			got := buf.String()
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}
