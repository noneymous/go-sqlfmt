package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentFunctionGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.FUNCTION, Value: "SUM"},
				lexer.Token{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
				lexer.Token{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
			},
			want: " SUM(xxx)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Function{Options: options, Element: tt.tokenSource}

			_ = el.Reindent(buf, nil, 0)
			got := buf.String()
			if tt.want != got {
				t.Errorf("\n=======================\n=== WANT =============>\n%s\n=======================\n=== GOT ==============>\n%s\n=======================", tt.want, got)
			} else {
				fmt.Println(fmt.Sprintf("%s\n%s", got, "========================================================================"))
			}
		})
	}
}
