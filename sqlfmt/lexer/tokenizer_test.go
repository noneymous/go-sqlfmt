package lexer

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTokenize_DollarQuotes_PreservesBodiesAndParameters verifies quote delimiters do not alter identifiers or parameters
func TestTokenize_DollarQuotes_PreservesBodiesAndParameters(t *testing.T) {

	// Prepare and run test cases
	for _, test := range []struct {
		name      string
		sql       string
		wantType  TokenType
		wantError bool
	}{
		{
			name:     "empty-body",
			sql:      `$$$$`,
			wantType: STRING,
		},
		{
			name:     "empty-tagged-body",
			sql:      `$a$$a$`,
			wantType: STRING,
		},
		{
			name: "unicode-tag",
			sql: `$ä_1$'quoted'; -- comment
$ä_1$`,
			wantType: STRING,
		},
		{
			name:     "different-inner-tag",
			sql:      `$outer$ $$inner$$ $other$ $outer$`,
			wantType: STRING,
		},
		{
			name:     "parameter",
			sql:      `$1`,
			wantType: IDENT,
		},
		{
			name:     "digit-not-tag",
			sql:      `$1$`,
			wantType: IDENT,
		},
		{
			name:     "identifier-with-dollars",
			sql:      `name$tag$`,
			wantType: IDENT,
		},
		{
			name:     "doubled-single-quote",
			sql:      `'a''b'`,
			wantType: STRING,
		},
		{
			name:      "missing-close",
			sql:       `$body$unfinished`,
			wantError: true,
		},
		{
			name:      "case-sensitive-close",
			sql:       `$body$x$BODY$`,
			wantError: true,
		},
		{
			name:      "overlapping-close",
			sql:       `$body$body$`,
			wantError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {

			// Verify the whole quoted value remains one token or reports its missing delimiter
			tokens, errTokens := Tokenize(test.sql)
			if test.wantError {
				if errTokens == nil {
					t.Error("Tokenize() error = 'nil', want = 'unterminated quote'")
				}
				return
			}
			if errTokens != nil {
				t.Fatalf("Tokenize() error = '%v', want = 'nil'", errTokens)
			}
			if len(tokens) != 2 || tokens[0].Type != test.wantType || tokens[0].Value != test.sql || tokens[1].Type != EOF {
				t.Errorf("Tokenize() tokens = '%v', want = 'one unchanged token and EOF'", tokens)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	var testingSQLStatement = strings.Trim(`select name, age, sum, sum(case xxx) from users where name xxx and age = 'xxx' limit 100 except 100`, "`")
	want := []Token{
		{Type: SELECT, Value: "SELECT"},
		{Type: IDENT, Value: "name"},
		{Type: COMMA, Value: ","},
		{Type: IDENT, Value: "age"},
		{Type: COMMA, Value: ","},
		{Type: IDENT, Value: "sum"},
		{Type: COMMA, Value: ","},
		{Type: FUNCTION, Value: "SUM"},
		{Type: STARTPARENTHESIS, Value: "("},
		{Type: CASE, Value: "CASE"},
		{Type: IDENT, Value: "xxx"},
		{Type: ENDPARENTHESIS, Value: ")"},
		{Type: FROM, Value: "FROM"},
		{Type: IDENT, Value: "users"},
		{Type: WHERE, Value: "WHERE"},
		{Type: IDENT, Value: "name"},
		{Type: IDENT, Value: "xxx"},
		{Type: AND, Value: "AND"},
		{Type: IDENT, Value: "age"},
		{Type: COMPARATOR, Value: "="},
		{Type: STRING, Value: "'xxx'"},
		{Type: LIMIT, Value: "LIMIT"},
		{Type: IDENT, Value: "100"},
		{Type: EXCEPT, Value: "EXCEPT"},
		{Type: IDENT, Value: "100"},
		{Type: EOF, Value: "EOF"},
	}
	got, err := Tokenize(testingSQLStatement)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func Test_peekComparator(t *testing.T) {

	tests := []struct {
		testSequence   string
		wantComperator string
		wantError      assert.ErrorAssertionFunc
	}{
		{
			testSequence:   "~~42", // LIKE
			wantComperator: "~~",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "~~*42", // ILIKE
			wantComperator: "~~*",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "!~~42", // NOT LIKE
			wantComperator: "!~~",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "!~~*42", // NOT ILIKE
			wantComperator: "!~~*",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "=42", // EQUAL
			wantComperator: "=",
			wantError:      assert.NoError,
		},
		{
			testSequence:   ">42", // GREATER
			wantComperator: ">",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "<42", // LOWER
			wantComperator: "<",
			wantError:      assert.NoError,
		},
		{
			testSequence:   ">=42", // GREATER OR EQUAL
			wantComperator: ">=",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "<=42", // LOWER OR EQUAL
			wantComperator: "<=",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "<>42", // NOT EQUAL
			wantComperator: "<>",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "!=42", // NOT EQUAL
			wantComperator: "!=",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "a!=42", // not STARTING with valid comparator but no error
			wantComperator: "",
			wantError:      assert.NoError,
		},
		{
			testSequence:   "!a!=42", // not STARTING with valid comparator
			wantComperator: "",
			wantError:      assert.Error,
		},
		{
			testSequence:   "!=!=42", // not STARTING with valid comparator
			wantComperator: "",
			wantError:      assert.Error,
		},
		{
			testSequence:   "~~~~~~~~~~*42", // not STARTING with valid comparator
			wantComperator: "",
			wantError:      assert.Error,
		},
		{
			testSequence:   "~~~~~~~~~~*~~42", // not STARTING with valid comparator
			wantComperator: "",
			wantError:      assert.Error,
		},
		{
			testSequence:   "!=a!=42", // starting with valid validator, although it's invalid later down the street
			wantComperator: "!=",
			wantError:      assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testSequence, func(t *testing.T) {

			// Fill test reader with input sequence
			r := bufio.NewReader(strings.NewReader(tt.testSequence))

			got, err := peekComparator(r)
			if !tt.wantError(t, err, fmt.Sprintf("peekComparator(%v)", r)) {
				return
			}
			assert.Equalf(t, tt.wantComperator, got, "peekComparator(%v)", r)

			// Check remaining characters in buffer
			remaining, _ := r.ReadString('\n')
			if len(remaining) != len(tt.testSequence) {
				t.Errorf("Invalid remaining buffer length. Want %d got %d", len(tt.testSequence), len(remaining))
				return
			}

			// Check if original string is reverted
			if remaining != tt.testSequence {
				t.Errorf("Invalid remaining buffer content. Want '%s' got '%s'", tt.testSequence, remaining)
				return
			}
		})
	}
}
