package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentCaseGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.CASE, Value: "CASE"},
				lexer.Token{Type: lexer.WHEN, Value: "WHEN"},
				lexer.Token{Type: lexer.IDENT, Value: "something"},
				lexer.Token{Type: lexer.IDENT, Value: "something"},
				lexer.Token{Type: lexer.END, Value: "END"},
			},
			want: "\n  CASE\n     WHEN something something\n  END",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Case{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
