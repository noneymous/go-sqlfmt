package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentJoinGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normalcase",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.LEFT, Value: "LEFT"},
				lexer.Token{Options: options, Type: lexer.OUTER, Value: "OUTER"},
				lexer.Token{Options: options, Type: lexer.JOIN, Value: "JOIN"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "sometable"},
				lexer.Token{Options: options, Type: lexer.ON, Value: "ON"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "status1"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "="},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "status2"},
			},

			want: "\nLEFT OUTER JOIN sometable ON status1 = status2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Join{Options: options, Element: tt.tokenSource}

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
