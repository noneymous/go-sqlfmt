package parser

import (
	"github.com/noneymous/go-sqlfmt/sqlfmt/reindenters"
	"reflect"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

func TestParse(t *testing.T) {
	options := reindenters.DefaultOptions()

	tests := []struct {
		name        string
		tokenSource []lexer.Token
		want        []reindenters.Reindenter
	}{
		{
			name: "normal test case 1",
			tokenSource: []lexer.Token{
				{Type: lexer.SELECT, Value: "SELECT"},
				{Type: lexer.IDENT, Value: "name"},
				{Type: lexer.COMMA, Value: ","},
				{Type: lexer.IDENT, Value: "age"},
				{Type: lexer.COMMA, Value: ","},

				{Type: lexer.FUNCTION, Value: "SUM"},
				{Type: lexer.STARTPARENTHESIS, Value: "("},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ENDPARENTHESIS, Value: ")"},

				{Type: lexer.STARTPARENTHESIS, Value: "("},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ENDPARENTHESIS, Value: ")"},

				{Type: lexer.TYPE, Value: "TEXT"},
				{Type: lexer.STARTPARENTHESIS, Value: "("},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ENDPARENTHESIS, Value: ")"},

				{Type: lexer.FROM, Value: "FROM"},
				{Type: lexer.IDENT, Value: "user"},
				{Type: lexer.WHERE, Value: "WHERE"},
				{Type: lexer.IDENT, Value: "name"},
				{Type: lexer.IDENT, Value: "="},
				{Type: lexer.STRING, Value: "'xxx'"},
				{Type: lexer.EOF, Value: "EOF"},
			},
			want: []reindenters.Reindenter{
				&reindenters.Select{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "name"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.COMMA, Value: ","}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "age"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.COMMA, Value: ","}},
						&reindenters.Function{
							Options: options,
							Element: []reindenters.Reindenter{
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.FUNCTION, Value: "SUM"}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
							},
						},
						&reindenters.Parenthesis{
							Options: options,
							Element: []reindenters.Reindenter{
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
							},
						},
						&reindenters.TypeCast{
							Options: options,
							Element: []reindenters.Reindenter{
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.TYPE, Value: "TEXT"}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
							},
						},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "user"}},
					},
				},
				&reindenters.Where{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.WHERE, Value: "WHERE"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "name"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "="}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.STRING, Value: "'xxx'"}},
					},
				},
			},
		},
		{
			name: "normal test case 2",
			tokenSource: []lexer.Token{
				{Type: lexer.SELECT, Value: "SELECT"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.FROM, Value: "FROM"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.WHERE, Value: "WHERE"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.IN, Value: "IN"},
				{Type: lexer.STARTPARENTHESIS, Value: "("},
				{Type: lexer.SELECT, Value: "SELECT"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.FROM, Value: "FROM"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.JOIN, Value: "JOIN"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ON, Value: "ON"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.IDENT, Value: "="},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ENDPARENTHESIS, Value: ")"},
				{Type: lexer.GROUP, Value: "GROUP"},
				{Type: lexer.BY, Value: "BY"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ORDER, Value: "ORDER"},
				{Type: lexer.BY, Value: "BY"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.LIMIT, Value: "LIMIT"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.UNION, Value: "UNION"},
				{Type: lexer.ALL, Value: "ALL"},
				{Type: lexer.SELECT, Value: "SELECT"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.FROM, Value: "FROM"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.EOF, Value: "EOF"},
			},
			want: []reindenters.Reindenter{
				&reindenters.Select{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.Where{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.WHERE, Value: "WHERE"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IN, Value: "IN"}},
						&reindenters.Subquery{
							Options: options,
							Element: []reindenters.Reindenter{
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.STARTPARENTHESIS, Value: "("}},
								&reindenters.Select{
									Options: options,
									Element: []reindenters.Reindenter{
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
									},
									IndentLevel: 0,
								},
								&reindenters.From{
									Options: options,
									Element: []reindenters.Reindenter{
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
									},
									IndentLevel: 0,
								},
								&reindenters.Join{
									Options: options,
									Element: []reindenters.Reindenter{
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.JOIN, Value: "JOIN"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ON, Value: "ON"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "="}},
										reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
									},
									IndentLevel: 0,
								},
								reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}},
							},
							IndentLevel: 0,
						},
					},
				},
				&reindenters.GroupBy{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.GROUP, Value: "GROUP"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.BY, Value: "BY"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.OrderBy{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ORDER, Value: "ORDER"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.BY, Value: "BY"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.Limit{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.LIMIT, Value: "LIMIT"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.TieGroup{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.UNION, Value: "UNION"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.ALL, Value: "ALL"}},
					},
				},
				&reindenters.Select{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.SELECT, Value: "SELECT"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.FROM, Value: "FROM"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "xxx"}},
					},
				},
			},
		},
		{
			name: "normal test case 3",
			tokenSource: []lexer.Token{
				{Type: lexer.UPDATE, Value: "UPDATE"},
				{Type: lexer.IDENT, Value: "user"},
				{Type: lexer.SET, Value: "SET"},
				{Type: lexer.IDENT, Value: "point"},
				{Type: lexer.IDENT, Value: "="},
				{Type: lexer.IDENT, Value: "0"},
				{Type: lexer.EOF, Value: "EOF"},
			},
			want: []reindenters.Reindenter{
				&reindenters.Update{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.UPDATE, Value: "UPDATE"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "user"}},
					},
				},
				&reindenters.Set{
					Options: options,
					Element: []reindenters.Reindenter{
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.SET, Value: "SET"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "point"}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "="}},
						reindenters.Token{Options: options, Token: lexer.Token{Type: lexer.IDENT, Value: "0"}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.tokenSource, options)
			if err != nil {
				t.Errorf("ERROR: %#v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	options := reindenters.DefaultOptions()
	testingData := []lexer.Token{
		{Type: lexer.SELECT, Value: "SELECT"},
		{Type: lexer.IDENT, Value: "name"},
		{Type: lexer.COMMA, Value: ","},
		{Type: lexer.IDENT, Value: "age"},
		{Type: lexer.FROM, Value: "FROM"},
		{Type: lexer.IDENT, Value: "user"},
		{Type: lexer.EOF, Value: "EOF"},
	}
	r, _ := NewParser(testingData, options)
	want := []lexer.Token{
		{Type: lexer.SELECT, Value: "SELECT"},
		{Type: lexer.IDENT, Value: "name"},
		{Type: lexer.COMMA, Value: ","},
		{Type: lexer.IDENT, Value: "age"},
		{Type: lexer.FROM, Value: "FROM"},
		{Type: lexer.IDENT, Value: "user"},
		{Type: lexer.EOF, Value: "EOF"},
	}
	got := r.tokens

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("initialize parser failed: want %#v got %#v", want, got)
	}
}

func Test_parseSegment(t *testing.T) {
	type want struct {
		stmt    []string
		lastIdx int
	}

	tests := []struct {
		name          string
		source        []lexer.Token
		endTokenTypes []lexer.TokenType
		want          *want
	}{
		{
			name: "normal_test",
			source: []lexer.Token{
				{Type: lexer.SELECT, Value: "SELECT"},
				{Type: lexer.IDENT, Value: "name"},
				{Type: lexer.COMMA, Value: ","},
				{Type: lexer.IDENT, Value: "age"},
				{Type: lexer.FROM, Value: "FROM"},
				{Type: lexer.IDENT, Value: "user"},
				{Type: lexer.EOF, Value: "EOF"},
			},
			endTokenTypes: []lexer.TokenType{lexer.FROM},
			want: &want{
				stmt:    []string{"SELECT", "name", ",", "age"},
				lastIdx: 4,
			},
		},
		{
			name: "normal_test3",
			source: []lexer.Token{
				{Type: lexer.LEFT, Value: "LEFT"},
				{Type: lexer.JOIN, Value: "JOIN"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.ON, Value: "ON"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.IDENT, Value: "="},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.WHERE, Value: "WHERE"},
			},
			endTokenTypes: []lexer.TokenType{lexer.WHERE},
			want: &want{
				stmt:    []string{"LEFT", "JOIN", "xxx", "ON", "xxx", "=", "xxx"},
				lastIdx: 7,
			},
		},
		{
			name: "normal_test4",
			source: []lexer.Token{
				{Type: lexer.UPDATE, Value: "UPDATE"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.SET, Value: "SET"},
			},
			endTokenTypes: []lexer.TokenType{lexer.SET},
			want: &want{
				stmt:    []string{"UPDATE", "xxx"},
				lastIdx: 2,
			},
		},
		{
			name: "normal_test5",
			source: []lexer.Token{
				{Type: lexer.INSERT, Value: "INSERT"},
				{Type: lexer.INTO, Value: "INTO"},
				{Type: lexer.IDENT, Value: "xxx"},
				{Type: lexer.VALUES, Value: "VALUES"},
			},
			endTokenTypes: []lexer.TokenType{lexer.VALUES},
			want: &want{
				stmt:    []string{"INSERT", "INTO", "xxx"},
				lastIdx: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Prepare parser with test data
			r := &Parser{tokens: tt.source, endTypes: tt.endTokenTypes}

			// Process test parser
			endIdx, err := r.parseSegment()
			if err != nil {
				t.Errorf("ERROR:%#v", err)
			}

			// Convert token sequence to string sequence
			var gotStmt []string
			for _, v := range r.result {
				if tok, ok := v.(reindenters.Token); ok {
					gotStmt = append(gotStmt, tok.Value)
				}
			}

			// Evaluate string sequence (parser results) and last index as expected
			if !reflect.DeepEqual(gotStmt, tt.want.stmt) {
				t.Errorf("want %v, got %v", tt.want.stmt, gotStmt)
			} else if endIdx != tt.want.lastIdx {
				t.Errorf("want %v, got %v", tt.want.lastIdx, endIdx)
			}
		})
	}
}
