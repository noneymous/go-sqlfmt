package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentSelectGroup(t *testing.T) {
	options := lexer.DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []lexer.Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []lexer.Reindenter{
				lexer.Token{Options: options, Type: lexer.SELECT, Value: "SELECT"},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "name"},
				lexer.Token{Options: options, Type: lexer.COMMA, Value: ","},
				lexer.Token{Options: options, Type: lexer.IDENT, Value: "age"},
			},
			want: "\nSELECT\n  name,\n  age",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			el := &Select{Options: options, Element: tt.tokenSource}

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

func TestIncrementIndentLevel(t *testing.T) {
	s := &Select{}
	s.IncrementIndentLevel(1)
	got := s.IndentLevel
	want := 1
	if got != want {
		t.Errorf("want %#v got %#v", want, got)
	}
}
