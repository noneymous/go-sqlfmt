package group

import (
	"bytes"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentFunctionGroup(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Type: lexer.FUNCTION, Value: "SUM"},
				lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("},
				lexer.Token{Type: lexer.IDENT, Value: "xxx"},
				lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"},
			},
			want: " SUM(xxx)",
		},
	}
	for _, tt := range tests {
		buf := &bytes.Buffer{}
		el := &Function{Element: tt.tokenSource}

		_ = el.Reindent(buf, lexer.Token{})
		got := buf.String()
		if tt.want != got {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
