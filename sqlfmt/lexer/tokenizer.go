package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type tokenizer struct {
	r *bufio.Reader
}

// Tokenize src and returns slice of Token. Ignores Token of white-space, new-line and tab.
func Tokenize(src string) ([]Token, error) {

	// Prepare tokenizer
	t := &tokenizer{
		r: bufio.NewReader(strings.NewReader(src)),
	}

	// Execute tokenizer
	var tokens []Token
	for {

		// Get next token
		token, err := t.scan()
		if err != nil {
			return nil, fmt.Errorf("tokenizer error: %w", err)
		}

		// Abort loop at the end
		if token.Type == EOF {

			// Append EOF token to tokens, because parser will also run until EOF token
			tokens = append(tokens, token)

			// Return generated sequence of tokens
			return tokens, nil
		}

		// Skip empty formatting token
		if token.Type == WHITESPACE {
			continue
		}
		if token.Type == NEWLINE {
			continue
		}
		if token.Type == TAB {
			continue
		}

		// Append token to token slice
		tokens = append(tokens, token)
	}
}

// scan reads the first character of t.r and returns it as a Token
func (t *tokenizer) scan() (Token, error) {

	// Read first character
	ch, _, errCh := t.r.ReadRune()
	if errCh != nil {
		if errCh.Error() == "EOF" {
			return Token{Type: EOF, Value: "EOF"}, nil
		}
		return Token{}, errCh
	}

	// Prepare buffer for current value
	var buf bytes.Buffer

	// Put already read character into buffer
	buf.WriteRune(ch)

	// Decide character, create and return token
	switch {
	case isWhitespace(ch):

		return Token{Type: WHITESPACE, Value: buf.String()}, nil

	case isNewline(ch):

		return Token{Type: NEWLINE, Value: buf.String()}, nil

	case isTab(ch):

		return Token{Type: TAB, Value: buf.String()}, nil

	case isPunctuation(ch):

		// Punctuation characters are only comprised out of a single character, except for DOBLECOLON tokens.
		// In case of a double colon, the next character needs to be read too.
		if isColon(ch) {

			// Read next character if there is one
			nextCh, _, errNext := t.r.ReadRune()
			if errNext != nil {
				if errNext.Error() == "EOF" { // Single colon was at the end of the string, which is okay
					return Token{Type: COLON, Value: buf.String()}, nil
				} else {
					return Token{}, errNext
				}
			}

			// Return colon or double colon token, depending on situation. Unread last character if necessary
			if isColon(nextCh) {
				return Token{Type: DOUBLECOLON, Value: fmt.Sprintf("%s%s", buf.String(), string(nextCh))}, nil
			} else {
				_ = t.r.UnreadRune() // Revert last read, because already belonged to the next token
				return Token{Type: COLON, Value: buf.String()}, nil
			}
		}

		// Lookup token type in punctuation map
		if ttype, ok := punctuationMap[buf.String()]; ok {
			return Token{Type: ttype, Value: buf.String()}, nil
		}

		// Return with error in case of unexpected value
		return Token{}, fmt.Errorf("unexpected punctuation value: %v", buf.String())

	case isSingleQuote(ch): // scan string which appears in the SQL statement surrounded by single quote such as 'xxxxxxxx'

		// read until next single-quote appears
		for {
			chNext, _, errNext := t.r.ReadRune()
			if errNext != nil {
				if errNext.Error() == "EOF" {
					return Token{}, fmt.Errorf("unexpected EOF expected closing quote")
				} else {
					return Token{}, errNext
				}
			}

			// Append character to value
			buf.WriteRune(chNext)

			// Break loop
			if isSingleQuote(chNext) {
				break
			}
		}

		// Return string token
		return Token{Type: STRING, Value: buf.String()}, nil

	default: // scan keyword token and parameter names and values

		// Read additional characters until value is complete
		for {
			chNext, _, errNext := t.r.ReadRune()
			if errNext != nil {
				if errNext.Error() == "EOF" {
					break
				} else {
					return Token{}, errNext
				}
			}

			// Unread last terminating character or add to buffer
			if isPunctuation(chNext) || isSingleQuote(chNext) || isWhitespace(chNext) || isNewline(chNext) || isTab(chNext) {
				_ = t.r.UnreadRune()
				break
			} else {
				buf.WriteRune(chNext)
			}
		}

		// Get token key and value
		key := strings.ToUpper(buf.String())
		val := key

		// Sanitize key and value, if they include a target operator '.'.
		// If token value contains period, it's specifying a target, e.g. a table. Drop that for the lookup.
		if strings.Contains(buf.String(), ".") {
			slice := strings.Split(buf.String(), ".")
			key = strings.ToUpper(slice[len(slice)-1])
			val = strings.Join(slice[:len(slice)-1], ".") + "." + key
		}

		// Check if value is keyword
		if ttype, ok := keywordMap[key]; ok {

			// Return looked-up token type, if it's not a function
			if ttype != FUNCTION {
				return Token{Type: ttype, Value: val}, nil
			}

			// Resolve ambiguity between arguments and function names. E.g. "sum" might be a function or a
			// column name. If parenthesis follows, it's definitely a function name, rather than a column name.
			if t.peekSubsequent(isParenthesisStart) {
				return Token{Type: FUNCTION, Value: val}, nil // Return looked-up token type, since value is a function
			} else {
				return Token{Type: IDENT, Value: buf.String()}, nil // Return IDENT token type, since value is not a function
			}
		}

		// Return IDENT token type, since value is not a keyword
		return Token{Type: IDENT, Value: buf.String()}, nil

	}
}

// peekSubsequent looks into the subsequent characters searching for a certain follow-up character but
// reverts all read characters at the end.
func (t *tokenizer) peekSubsequent(isCharacter func(ch rune) bool) bool {

	// Unread character at the end
	defer func() { _ = t.r.UnreadRune() }()

	// Read character
	nextCh, _, errNext := t.r.ReadRune()
	if errNext != nil {
		return false
	}

	// Evaluate character or, if necessary, step into recursive call to check subsequent character
	if isWhitespace(nextCh) {
		return t.peekSubsequent(isCharacter)
	} else if isCharacter(nextCh) {
		return true
	} else {
		return false
	}
}

var keywordMap = map[string]TokenType{
	"SELECT":      SELECT,
	"FROM":        FROM,
	"WHERE":       WHERE,
	"CASE":        CASE,
	"ORDER":       ORDER,
	"BY":          BY,
	"AS":          AS,
	"JOIN":        JOIN,
	"LEFT":        LEFT,
	"RIGHT":       RIGHT,
	"INNER":       INNER,
	"OUTER":       OUTER,
	"ON":          ON,
	"WHEN":        WHEN,
	"END":         END,
	"GROUP":       GROUP,
	"DESC":        DESC,
	"ASC":         ASC,
	"LIMIT":       LIMIT,
	"AND":         AND,
	"OR":          OR,
	"IN":          IN,
	"IS":          IS,
	"NOT":         NOT,
	"NULL":        NULL,
	"DISTINCT":    DISTINCT,
	"LIKE":        LIKE,
	"ILIKE":       ILIKE,
	"BETWEEN":     BETWEEN,
	"UNION":       UNION,
	"ALL":         ALL,
	"HAVING":      HAVING,
	"EXISTS":      EXISTS,
	"UPDATE":      UPDATE,
	"SET":         SET,
	"RETURNING":   RETURNING,
	"DELETE":      DELETE,
	"INSERT":      INSERT,
	"INTO":        INTO,
	"DO":          DO,
	"VALUES":      VALUES,
	"FOR":         FOR,
	"THEN":        THEN,
	"ELSE":        ELSE,
	"DISTINCTROW": DISTINCTROW,
	"FILTER":      FILTER,
	"WITHIN":      WITHIN,
	"COLLATE":     COLLATE,
	"INTERSECT":   INTERSECT,
	"EXCEPT":      EXCEPT,
	"OFFSET":      OFFSET,
	"FETCH":       FETCH,
	"FIRST":       FIRST,
	"ROWS":        ROWS,
	"USING":       USING,
	"OVERLAPS":    OVERLAPS,
	"NATURAL":     NATURAL,
	"CROSS":       CROSS,
	"ZONE":        ZONE,
	"NULLS":       NULLS,
	"LAST":        LAST,
	"AT":          AT,
	"LOCK":        LOCK,
	"WITH":        WITH,
	"BIG":         TYPE,
	"BIGSERIAL":   TYPE,
	"BOOLEAN":     TYPE,
	"CHAR":        TYPE,
	"BIT":         TYPE,
	"TEXT":        TYPE,
	"INTEGER":     TYPE,
	"NUMERIC":     TYPE,
	"DECIMAL":     TYPE,
	"DEC":         TYPE,
	"FLOAT":       TYPE,
	"CUSTOMTYPE":  TYPE,
	"VARCHAR":     TYPE,
	"VARBIT":      TYPE,
	"TIMESTAMP":   TYPE,
	"TIME":        TYPE,
	"SECOND":      TYPE,
	"INTERVAL":    TYPE,

	// Common SQL Functions
	"SUM":             FUNCTION,
	"AVG":             FUNCTION,
	"MAX":             FUNCTION,
	"MIN":             FUNCTION,
	"COUNT":           FUNCTION,
	"COALESCE":        FUNCTION,
	"EXTRACT":         FUNCTION,
	"OVERLAY":         FUNCTION,
	"POSITION":        FUNCTION,
	"CAST":            FUNCTION,
	"SUBSTRING":       FUNCTION,
	"TRIM":            FUNCTION,
	"XMLELEMENT":      FUNCTION,
	"XMLFOREST":       FUNCTION,
	"XMLCONCAT":       FUNCTION,
	"RANDOM":          FUNCTION,
	"DATE_PART":       FUNCTION,
	"DATE_TRUNC":      FUNCTION,
	"TO_TIMESTAMP":    FUNCTION,
	"ARRAY_AGG":       FUNCTION,
	"PERCENTILE_DISC": FUNCTION,
	"GREATEST":        FUNCTION,
	"LEAST":           FUNCTION,
	"OVER":            FUNCTION,
	"ROW_NUMBER":      FUNCTION,

	// Additional SQLite functions
	"CHANGES":                   FUNCTION,
	"CONCAT":                    FUNCTION,
	"CONCAT_WS":                 FUNCTION,
	"FORMAT":                    FUNCTION,
	"GLOB":                      FUNCTION,
	"HEX":                       FUNCTION,
	"IFNULL":                    FUNCTION,
	"IIF":                       FUNCTION,
	"INSTR":                     FUNCTION,
	"LAST_INSERT_ROWID":         FUNCTION,
	"LIKELIHOOD":                FUNCTION,
	"LIKELY":                    FUNCTION,
	"LOAD_EXTENSION":            FUNCTION,
	"NULLIF":                    FUNCTION,
	"OCTET_LENGTH":              FUNCTION,
	"PRINTF":                    FUNCTION,
	"QUOTE":                     FUNCTION,
	"RANDOMBLOB":                FUNCTION,
	"SOUNDEX":                   FUNCTION,
	"SQLITE_COMPILEOPTION_GET":  FUNCTION,
	"SQLITE_COMPILEOPTION_USED": FUNCTION,
	"SQLITE_OFFSET":             FUNCTION,
	"SQLITE_SOURCE_ID":          FUNCTION,
	"SQLITE_VERSION":            FUNCTION,
	"SUBSTR":                    FUNCTION,
	"TOTAL_CHANGES":             FUNCTION,
	"TYPEOF":                    FUNCTION,
	"UNHEX":                     FUNCTION,
	"UNICODE":                   FUNCTION,
	"UNLIKELY":                  FUNCTION,
	"ZEROBLOB":                  FUNCTION,

	// Additional PostgreSQL functions
	"BTRIM":                  FUNCTION,
	"CHAR_LENGTH":            FUNCTION,
	"CHARACTER_LENGTH":       FUNCTION,
	"INITCAP":                FUNCTION,
	"LENGTH":                 FUNCTION,
	"LOWER":                  FUNCTION,
	"LPAD":                   FUNCTION,
	"LTRIM":                  FUNCTION,
	"REPEAT":                 FUNCTION,
	"REPLACE":                FUNCTION,
	"RPAD":                   FUNCTION,
	"RTRIM":                  FUNCTION,
	"STRPOS":                 FUNCTION,
	"TRANSLATE":              FUNCTION,
	"UPPER":                  FUNCTION,
	"ABS":                    FUNCTION,
	"CEIL":                   FUNCTION,
	"CEILING":                FUNCTION,
	"DIV":                    FUNCTION,
	"EXP":                    FUNCTION,
	"FLOOR":                  FUNCTION,
	"MOD":                    FUNCTION,
	"POWER":                  FUNCTION,
	"ROUND":                  FUNCTION,
	"SETSEED":                FUNCTION,
	"SIGN":                   FUNCTION,
	"SQRT":                   FUNCTION,
	"TRUNC":                  FUNCTION,
	"AGE":                    FUNCTION,
	"CURRENT_DATE":           FUNCTION,
	"CURRENT_TIME":           FUNCTION,
	"CURRENT_TIMESTAMP":      FUNCTION,
	"LOCALTIME":              FUNCTION,
	"LOCALTIMESTAMP":         FUNCTION,
	"NOW":                    FUNCTION,
	"TO_CHAR":                FUNCTION,
	"TO_DATE":                FUNCTION,
	"TO_NUMBER":              FUNCTION,
	"HAS_TABLE_PRIVILEGE":    FUNCTION,
	"HAS_SCHEMA_PRIVILEGE":   FUNCTION,
	"HAS_DATABASE_PRIVILEGE": FUNCTION,

	// Additional Oracle functions
	"ACOS":                         FUNCTION,
	"ASIN":                         FUNCTION,
	"ATAN":                         FUNCTION,
	"ATAN2":                        FUNCTION,
	"BITAND":                       FUNCTION,
	"COS":                          FUNCTION,
	"COSH":                         FUNCTION,
	"LN":                           FUNCTION,
	"LOG":                          FUNCTION,
	"NANVL":                        FUNCTION,
	"REMAINDER":                    FUNCTION,
	"SIN":                          FUNCTION,
	"SINH":                         FUNCTION,
	"TAN":                          FUNCTION,
	"TANH":                         FUNCTION,
	"WIDTH_BUCKET":                 FUNCTION,
	"CHR":                          FUNCTION,
	"NLS_INITCAP":                  FUNCTION,
	"NLS_LOWER":                    FUNCTION,
	"NLSSORT":                      FUNCTION,
	"NLS_UPPER":                    FUNCTION,
	"REGEXP_REPLACE":               FUNCTION,
	"REGEXP_SUBSTR":                FUNCTION,
	"TREAT":                        FUNCTION,
	"ASCII":                        FUNCTION,
	"REGEXP_INSTR":                 FUNCTION,
	"ADD_MONTHS":                   FUNCTION,
	"DBTIMEZONE":                   FUNCTION,
	"FROM_TZ":                      FUNCTION,
	"LAST_DAY":                     FUNCTION,
	"MONTHS_BETWEEN":               FUNCTION,
	"NEW_TIME":                     FUNCTION,
	"NEXT_DAY":                     FUNCTION,
	"NUMTODSINTERVAL":              FUNCTION,
	"NUMTOYMINTERVAL":              FUNCTION,
	"SESSIONTIMEZONE":              FUNCTION,
	"SYS_EXTRACT_UTC":              FUNCTION,
	"SYSDATE":                      FUNCTION,
	"SYSTIMESTAMP":                 FUNCTION,
	"TO_TIMESTAMP_TZ":              FUNCTION,
	"TO_DSINTERVAL":                FUNCTION,
	"TO_YMINTERVAL":                FUNCTION,
	"TZ_OFFSET":                    FUNCTION,
	"ASCIISTR":                     FUNCTION,
	"BIN_TO_NUM":                   FUNCTION,
	"CHARTOROWID":                  FUNCTION,
	"COMPOSE":                      FUNCTION,
	"CONVERT":                      FUNCTION,
	"DECOMPOSE":                    FUNCTION,
	"HEXTORAW":                     FUNCTION,
	"RAWTOHEX":                     FUNCTION,
	"RAWTONHEX":                    FUNCTION,
	"ROWIDTOCHAR":                  FUNCTION,
	"ROWIDTONCHAR":                 FUNCTION,
	"SCN_TO_TIMESTAMP":             FUNCTION,
	"TIMESTAMP_TO_SCN":             FUNCTION,
	"TO_BINARY_DOUBLE":             FUNCTION,
	"TO_BINARY_FLOAT":              FUNCTION,
	"TO_CLOB":                      FUNCTION,
	"TO_LOB":                       FUNCTION,
	"TO_MULTI_BYTE":                FUNCTION,
	"TO_NCHAR":                     FUNCTION,
	"TO_NCLOB":                     FUNCTION,
	"TO_SINGLE_BYTE":               FUNCTION,
	"UNISTR":                       FUNCTION,
	"CARDINALITY":                  FUNCTION,
	"COLLECT":                      FUNCTION,
	"POWERMULTISET":                FUNCTION,
	"POWERMULTISET_BY_CARDINALITY": FUNCTION,
	"BFILENAME":                    FUNCTION,
	"CV":                           FUNCTION,
	"DECODE":                       FUNCTION,
	"DEPTH":                        FUNCTION,
	"DUMP":                         FUNCTION,
	"EMPTY_BLOB":                   FUNCTION,
	"EMPTY_CLOB":                   FUNCTION,
	"EXISTSNODE":                   FUNCTION,
	"EXTRACTVALUE":                 FUNCTION,
	"LNNVL":                        FUNCTION,
	"NLS_CHARSET_DECL_LEN":         FUNCTION,
	"NLS_CHARSET_ID":               FUNCTION,
	"NLS_CHARSET_NAME":             FUNCTION,
	"NVL":                          FUNCTION,
	"NVL2":                         FUNCTION,
	"ORA_HASH":                     FUNCTION,
	"PATH":                         FUNCTION,
	"PRESENTNNV":                   FUNCTION,
	"PRESENTV":                     FUNCTION,
	"PREVIOUS":                     FUNCTION,
	"SYS_CONNECT_BY_PATH":          FUNCTION,
	"SYS_CONTEXT":                  FUNCTION,
	"SYS_DBURIGEN":                 FUNCTION,
	"SYS_GUID":                     FUNCTION,
	"SYS_TYPEID":                   FUNCTION,
	"SYS_XMLAGG":                   FUNCTION,
	"SYS_XMLGEN":                   FUNCTION,
	"UID":                          FUNCTION,
	"UPDATEXML":                    FUNCTION,
	"USER":                         FUNCTION,
	"USERENV":                      FUNCTION,
	"VSIZE":                        FUNCTION,
	"XMLAGG":                       FUNCTION,
	"XMLCOLATTVAL":                 FUNCTION,
	"XMLSEQUENCE":                  FUNCTION,
	"XMLTRANSFORM":                 FUNCTION,
	"CORR":                         FUNCTION,
	"CORR_S":                       FUNCTION,
	"CORR_K":                       FUNCTION,
	"COVAR_POP":                    FUNCTION,
	"COVAR_SAMP":                   FUNCTION,
	"CUME_DIST":                    FUNCTION,
	"DENSE_RANK":                   FUNCTION,
	"GROUP_ID":                     FUNCTION,
	"GROUPING":                     FUNCTION,
	"GROUPING_ID":                  FUNCTION,
	"MEDIAN":                       FUNCTION,
	"PERCENTILE_CONT":              FUNCTION,
	"PERCENT_RANK":                 FUNCTION,
	"RANK":                         FUNCTION,
	"STATS_BINOMIAL_TEST":          FUNCTION,
	"STATS_CROSSTAB":               FUNCTION,
	"STATS_F_TEST":                 FUNCTION,
	"STATS_KS_TEST":                FUNCTION,
	"STATS_MODE":                   FUNCTION,
	"STATS_MW_TEST":                FUNCTION,
	"STATS_ONE_WAY_ANOVA":          FUNCTION,
	"STATS_T_TEST_ONE":             FUNCTION,
	"STATS_T_TEST_PAIRD":           FUNCTION,
	"STATS_T_TEST_INDEP":           FUNCTION,
	"STATS_T_TEST_INDEPU":          FUNCTION,
	"STATS_WSR_TEST":               FUNCTION,
	"STDDEV":                       FUNCTION,
	"STDDEV_POP":                   FUNCTION,
	"STDDEV_SAMP":                  FUNCTION,
	"VAR_POP":                      FUNCTION,
	"VAR_SAMP":                     FUNCTION,
	"VARIANCE":                     FUNCTION,

	// Additional MsSQL functions
	"CHARINDEX":       FUNCTION,
	"DATALENGTH":      FUNCTION,
	"DIFFERENCE":      FUNCTION,
	"LEN":             FUNCTION,
	"NCHAR":           FUNCTION,
	"PATINDEX":        FUNCTION,
	"QUOTENAME":       FUNCTION,
	"REPLICATE":       FUNCTION,
	"REVERSE":         FUNCTION,
	"SPACE":           FUNCTION,
	"STR":             FUNCTION,
	"STUFF":           FUNCTION,
	"ATN2":            FUNCTION,
	"COT":             FUNCTION,
	"DEGREES":         FUNCTION,
	"LOG10":           FUNCTION,
	"PI":              FUNCTION,
	"RADIANS":         FUNCTION,
	"RAND":            FUNCTION,
	"SQUARE":          FUNCTION,
	"DATEADD":         FUNCTION,
	"DATEDIFF":        FUNCTION,
	"DATEFROMPARTS":   FUNCTION,
	"DATENAME":        FUNCTION,
	"DATEPART":        FUNCTION,
	"DAY":             FUNCTION,
	"GETDATE":         FUNCTION,
	"GETUTCDATE":      FUNCTION,
	"ISDATE":          FUNCTION,
	"MONTH":           FUNCTION,
	"SYSDATETIME":     FUNCTION,
	"YEAR":            FUNCTION,
	"CURRENT_USER":    FUNCTION,
	"ISNULL":          FUNCTION,
	"ISNUMERIC":       FUNCTION,
	"SESSION_USER":    FUNCTION,
	"SESSIONPROPERTY": FUNCTION,
	"SYSTEM_USER":     FUNCTION,
	"USER_NAME":       FUNCTION,
}

var punctuationMap = map[string]TokenType{
	"(": STARTPARENTHESIS,
	")": ENDPARENTHESIS,
	"[": STARTBRACKET,
	"]": ENDBRACKET,
	"{": STARTBRACE,
	"}": ENDBRACKET,
	",": COMMA,
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '　'
}

func isTab(ch rune) bool {
	return ch == '\t'
}

func isNewline(ch rune) bool {
	return ch == '\n'
}

func isComma(ch rune) bool {
	return ch == ','
}

func isColon(ch rune) bool {
	return ch == ':'
}

func isParenthesisStart(ch rune) bool {
	return ch == '('
}

func isParenthesisEnd(ch rune) bool {
	return ch == ')'
}

func isBracketStart(ch rune) bool {
	return ch == '['
}

func isBracketEnd(ch rune) bool {
	return ch == ']'
}

func isBraceStart(ch rune) bool {
	return ch == '{'
}

func isBraceEnd(ch rune) bool {
	return ch == '}'
}

func isPunctuation(ch rune) bool {
	return isComma(ch) || isColon(ch) || isParenthesisStart(ch) || isParenthesisEnd(ch) || isBracketStart(ch) || isBracketEnd(ch) || isBraceStart(ch) || isBraceEnd(ch)
}

func isSingleQuote(ch rune) bool {
	return ch == '\''
}
