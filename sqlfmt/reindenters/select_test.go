package reindenters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestReindentSelectGroup(t *testing.T) {
	options := DefaultOptions()
	tests := []struct {
		name        string
		tokenSource []Reindenter
		want        string
	}{
		{
			name: "normal case",
			tokenSource: []Reindenter{
				Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "name"}},
				Token{Options: options, Token: lexer.Token{Type: lexer.COMMA, Value: ","}},
				Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "age"}},
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

func TestIncrementIndent(t *testing.T) {
	options := DefaultOptions()
	s := &Select{
		Options:     options,
		Element:     nil,
		IndentLevel: 0,
		ColumnCount: 0,
	}
	s.IncrementIndent(1)
	got := s.IndentLevel
	want := 1
	if got != want {
		t.Errorf("want %#v got %#v", want, got)
	}
}
