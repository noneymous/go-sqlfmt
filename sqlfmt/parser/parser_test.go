package parser

import (
	"reflect"
	"testing"

	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/noneymous/go-sqlfmt/sqlfmt/retriever/reindenters"
)

func TestParseTokens(t *testing.T) {
	options := lexer.DefaultOptions()
	testingData := []lexer.Token{
		{Options: options, Type: lexer.SELECT, Value: "SELECT"},
		{Options: options, Type: lexer.IDENT, Value: "name"},
		{Options: options, Type: lexer.COMMA, Value: ","},
		{Options: options, Type: lexer.IDENT, Value: "age"},
		{Options: options, Type: lexer.COMMA, Value: ","},

		{Options: options, Type: lexer.FUNCTION, Value: "SUM"},
		{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},

		{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},

		{Options: options, Type: lexer.TYPE, Value: "TEXT"},
		{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},

		{Options: options, Type: lexer.FROM, Value: "FROM"},
		{Options: options, Type: lexer.IDENT, Value: "user"},
		{Options: options, Type: lexer.WHERE, Value: "WHERE"},
		{Options: options, Type: lexer.IDENT, Value: "name"},
		{Options: options, Type: lexer.IDENT, Value: "="},
		{Options: options, Type: lexer.STRING, Value: "'xxx'"},
		{Options: options, Type: lexer.EOF, Value: "EOF"},
	}
	testingData2 := []lexer.Token{
		{Options: options, Type: lexer.SELECT, Value: "SELECT"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.FROM, Value: "FROM"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.WHERE, Value: "WHERE"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.IN, Value: "IN"},
		{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
		{Options: options, Type: lexer.SELECT, Value: "SELECT"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.FROM, Value: "FROM"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.JOIN, Value: "JOIN"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ON, Value: "ON"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.IDENT, Value: "="},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
		{Options: options, Type: lexer.GROUP, Value: "GROUP"},
		{Options: options, Type: lexer.BY, Value: "BY"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.ORDER, Value: "ORDER"},
		{Options: options, Type: lexer.BY, Value: "BY"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.LIMIT, Value: "LIMIT"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.UNION, Value: "UNION"},
		{Options: options, Type: lexer.ALL, Value: "ALL"},
		{Options: options, Type: lexer.SELECT, Value: "SELECT"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.FROM, Value: "FROM"},
		{Options: options, Type: lexer.IDENT, Value: "xxx"},
		{Options: options, Type: lexer.EOF, Value: "EOF"},
	}
	testingData3 := []lexer.Token{
		{Options: options, Type: lexer.UPDATE, Value: "UPDATE"},
		{Options: options, Type: lexer.IDENT, Value: "user"},
		{Options: options, Type: lexer.SET, Value: "SET"},
		{Options: options, Type: lexer.IDENT, Value: "point"},
		{Options: options, Type: lexer.IDENT, Value: "="},
		{Options: options, Type: lexer.IDENT, Value: "0"},
		{Options: options, Type: lexer.EOF, Value: "EOF"},
	}

	tests := []struct {
		name        string
		tokenSource []lexer.Token
		want        []lexer.Reindenter
	}{
		{
			name:        "normal test case 1",
			tokenSource: testingData,
			want: []lexer.Reindenter{
				&reindenters.Select{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.SELECT, Value: "SELECT"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "name"},
						lexer.Token{Options: options, Type: lexer.COMMA, Value: ","},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "age"},
						lexer.Token{Options: options, Type: lexer.COMMA, Value: ","},
						&reindenters.Function{
							Options: options,
							Element: []lexer.Reindenter{
								lexer.Token{Options: options, Type: lexer.FUNCTION, Value: "SUM"},
								lexer.Token{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
								lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
								lexer.Token{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
							},
						},
						&reindenters.Parenthesis{
							Options: options,
							Element: []lexer.Reindenter{
								lexer.Token{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
								lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
								lexer.Token{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
							},
						},
						&reindenters.TypeCast{
							Options: options,
							Element: []lexer.Reindenter{
								lexer.Token{Options: options, Type: lexer.TYPE, Value: "TEXT"},
								lexer.Token{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
								lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
								lexer.Token{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
							},
						},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.FROM, Value: "FROM"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "user"},
					},
				},
				&reindenters.Where{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.WHERE, Value: "WHERE"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "name"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "="},
						lexer.Token{Options: options, Type: lexer.STRING, Value: "'xxx'"},
					},
				},
			},
		},
		{
			name:        "normal test case 2",
			tokenSource: testingData2,
			want: []lexer.Reindenter{
				&reindenters.Select{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.SELECT, Value: "SELECT"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.FROM, Value: "FROM"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.Where{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.WHERE, Value: "WHERE"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
						lexer.Token{Options: options, Type: lexer.IN, Value: "IN"},
						&reindenters.Subquery{
							Options: options,
							Element: []lexer.Reindenter{
								lexer.Token{Options: options, Type: lexer.STARTPARENTHESIS, Value: "("},
								&reindenters.Select{
									Options: options,
									Element: []lexer.Reindenter{
										lexer.Token{Options: options, Type: lexer.SELECT, Value: "SELECT"},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
									},
									IndentLevel: 1,
								},
								&reindenters.From{
									Options: options,
									Element: []lexer.Reindenter{
										lexer.Token{Options: options, Type: lexer.FROM, Value: "FROM"},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
									},
									IndentLevel: 1,
								},
								&reindenters.Join{
									Options: options,
									Element: []lexer.Reindenter{
										lexer.Token{Options: options, Type: lexer.JOIN, Value: "JOIN"},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
										lexer.Token{Options: options, Type: lexer.ON, Value: "ON"},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "="},
										lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
									},
									IndentLevel: 1,
								},
								lexer.Token{Options: options, Type: lexer.ENDPARENTHESIS, Value: ")"},
							},
							IndentLevel: 1,
						},
					},
				},
				&reindenters.GroupBy{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.GROUP, Value: "GROUP"},
						lexer.Token{Options: options, Type: lexer.BY, Value: "BY"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.OrderBy{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.ORDER, Value: "ORDER"},
						lexer.Token{Options: options, Type: lexer.BY, Value: "BY"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.Limit{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.LIMIT, Value: "LIMIT"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.TieGroup{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.UNION, Value: "UNION"},
						lexer.Token{Options: options, Type: lexer.ALL, Value: "ALL"},
					},
				},
				&reindenters.Select{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.SELECT, Value: "SELECT"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
				&reindenters.From{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.FROM, Value: "FROM"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "xxx"},
					},
				},
			},
		},
		{
			name:        "normal test case 3",
			tokenSource: testingData3,
			want: []lexer.Reindenter{
				&reindenters.Update{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.UPDATE, Value: "UPDATE"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "user"},
					},
				},
				&reindenters.Set{
					Options: options,
					Element: []lexer.Reindenter{
						lexer.Token{Options: options, Type: lexer.SET, Value: "SET"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "point"},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "="},
						lexer.Token{Options: options, Type: lexer.IDENT, Value: "0"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		got, err := ParseTokens(tt.tokenSource, options)
		if err != nil {
			t.Errorf("ERROR: %#v", err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
		}
	}
}
