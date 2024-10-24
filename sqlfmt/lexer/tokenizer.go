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

// DisableFunctionKeywords - Postgres has a few functions without parenthesis. They look like normal keywords,
// but they might conflict with table/column names. To address ambiguity, Postgres clients must therefore
// wrap affected names with double quotes. Otherwise, a name might be understood as a function keyword. In
// other dialects clients wouldn't care and quote these names. Disable keyword functions to avoid ambiguous
// names to be capitalized like function names.
// Affected names:
//   - LOCALTIME
//   - LOCALTIMESTAMP
//   - CURRENT_DATE
//   - CURRENT_TIME
//   - CURRENT_TIMESTAMP
//   - CURRENT_USER
//   - CURRENT_CATALOG
//   - SESSION_USER
//   - USER
var DisableFunctionKeywords = false

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

	// Check if comparator sequence starts with next read
	isComparator, _ := peekComparator(t.r)

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

	case isComparator != "":

		// First character of comparator was already read, read missing ones to remove them form buffer
		for i := len(isComparator) - 1; i > 0; i-- {
			_, _, _ = t.r.ReadRune()
		}
		return Token{Type: COMPARATOR, Value: isComparator}, nil

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
		var comparator = ""
		var comparatorErr error
		for {

			// Check if comparator can be detected in the next bytes, but only do so if no broken
			// comparator was peeked previously. otherwise an invalid symbol sequence might partially
			// be understood as a comparator.
			if comparatorErr == nil {
				comparator, comparatorErr = peekComparator(t.r)
				if comparator != "" {
					break // Nothing was read yet, no need to unread
				}
			}

			// Read next rune
			chNext, _, errNext := t.r.ReadRune()
			if errNext != nil {
				if errNext.Error() == "EOF" {
					break
				} else {
					return Token{}, errNext
				}
			}

			// Stop reading value if end is detected. Unread last unnecessary character
			if isPunctuation(chNext) || isSingleQuote(chNext) || isWhitespace(chNext) || isNewline(chNext) || isTab(chNext) {
				_ = t.r.UnreadRune()
				break
			}

			// Append character to value
			buf.WriteRune(chNext)
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

		// Check if value is function name, indicated by a lookup match and a subsequent parenthesis
		if ttype, ok := functionMap[key]; ok {
			if ttype == FUNCTION && t.peekSubsequent(isParenthesisStart) {
				return Token{Type: FUNCTION, Value: val}, nil
			}
		}

		// Check if value is keyword. Subsequent parenthesis would not indicate a function but a sub query.
		if ttype, ok := keywordMap[key]; ok {

			// Ambiguous edge case. Table name might (such as "user") might collide with Postgres'
			// parenthesis-less function "USER". To address ambiguity, Postgres clients must put "user" into
			// double quotes. However, in other databases this ambiguity does not exist, so clients would never
			// double quote in this situation. By default, these function keywords are enabled and handled as such.
			if ttype == FUNCTIONKEYWORD && DisableFunctionKeywords {
				return Token{Type: IDENT, Value: buf.String()}, nil
			}

			// Return keyword token
			return Token{Type: ttype, Value: val}, nil // Return looked-up token type
		}

		// Return IDENT token type without any sanitization, since it's neither a keyword nor a function
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
	"ANY":         ANY,
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

	/*
	 * Data types
	 */
	"BIG":        TYPE,
	"BIGSERIAL":  TYPE,
	"BOOLEAN":    TYPE,
	"CHAR":       TYPE,
	"BIT":        TYPE,
	"TEXT":       TYPE,
	"INTEGER":    TYPE,
	"NUMERIC":    TYPE,
	"DECIMAL":    TYPE,
	"DEC":        TYPE,
	"FLOAT":      TYPE,
	"CUSTOMTYPE": TYPE,
	"VARCHAR":    TYPE,
	"VARBIT":     TYPE,
	"TIMESTAMP":  TYPE,
	"TIME":       TYPE,
	"SECOND":     TYPE,
	"INTERVAL":   TYPE,

	/*
	 * Additional Postgres functions without parenthesis.
	 * These
	 */
	"LOCALTIME":         FUNCTIONKEYWORD,
	"LOCALTIMESTAMP":    FUNCTIONKEYWORD,
	"CURRENT_DATE":      FUNCTIONKEYWORD,
	"CURRENT_TIME":      FUNCTIONKEYWORD,
	"CURRENT_TIMESTAMP": FUNCTIONKEYWORD,
	"CURRENT_USER":      FUNCTIONKEYWORD,
	"CURRENT_CATALOG":   FUNCTIONKEYWORD,
	"SESSION_USER":      FUNCTIONKEYWORD,
	"USER":              FUNCTIONKEYWORD,
}

var functionMap = map[string]TokenType{

	/*
	 * PostgreSQL functions
	 */
	"BIT_LENGTH":                         FUNCTION,
	"CHAR_LENGTH":                        FUNCTION,
	"LOWER":                              FUNCTION,
	"OCTET_LENGTH":                       FUNCTION,
	"OVERLAY":                            FUNCTION,
	"POSITION":                           FUNCTION,
	"SUBSTRING":                          FUNCTION,
	"TRIM":                               FUNCTION,
	"UPPER":                              FUNCTION,
	"ASCII":                              FUNCTION,
	"BTRIM":                              FUNCTION,
	"CHR":                                FUNCTION,
	"CONCAT":                             FUNCTION,
	"CONCAT_WS":                          FUNCTION,
	"CONVERT":                            FUNCTION,
	"CONVERT_FROM":                       FUNCTION,
	"CONVERT_TO":                         FUNCTION,
	"DECODE":                             FUNCTION,
	"ENCODE":                             FUNCTION,
	"FORMAT":                             FUNCTION,
	"INITCAP":                            FUNCTION,
	"LEFT":                               FUNCTION,
	"LENGTH":                             FUNCTION,
	"LPAD":                               FUNCTION,
	"LTRIM":                              FUNCTION,
	"MD5":                                FUNCTION,
	"PG_CLIENT_ENCODING":                 FUNCTION,
	"QUOTE_IDENT":                        FUNCTION,
	"QUOTE_LITERAL":                      FUNCTION,
	"QUOTE_NULLABLE":                     FUNCTION,
	"REGEXP_MATCHES":                     FUNCTION,
	"REGEXP_REPLACE":                     FUNCTION,
	"REGEXP_SPLIT_TO_ARRAY":              FUNCTION,
	"REGEXP_SPLIT_TO_TABLE":              FUNCTION,
	"REPEAT":                             FUNCTION,
	"REPLACE":                            FUNCTION,
	"REVERSE":                            FUNCTION,
	"RIGHT":                              FUNCTION,
	"RPAD":                               FUNCTION,
	"RTRIM":                              FUNCTION,
	"SPLIT_PART":                         FUNCTION,
	"STRPOS":                             FUNCTION,
	"SUBSTR":                             FUNCTION,
	"TO_ASCII":                           FUNCTION,
	"TO_HEX":                             FUNCTION,
	"TRANSLATE":                          FUNCTION,
	"GET_BIT":                            FUNCTION,
	"GET_BYTE":                           FUNCTION,
	"SET_BIT":                            FUNCTION,
	"SET_BYTE":                           FUNCTION,
	"TO_CHAR":                            FUNCTION,
	"TO_DATE":                            FUNCTION,
	"TO_NUMBER":                          FUNCTION,
	"TO_TIMESTAMP":                       FUNCTION,
	"AGE":                                FUNCTION,
	"CLOCK_TIMESTAMP":                    FUNCTION,
	"DATE_PART":                          FUNCTION,
	"DATE_TRUNC":                         FUNCTION,
	"EXTRACT":                            FUNCTION,
	"ISFINITE":                           FUNCTION,
	"JUSTIFY_DAYS":                       FUNCTION,
	"JUSTIFY_HOURS":                      FUNCTION,
	"JUSTIFY_INTERVAL":                   FUNCTION,
	"NOW":                                FUNCTION,
	"STATEMENT_TIMESTAMP":                FUNCTION,
	"TIMEOFDAY":                          FUNCTION,
	"TRANSACTION_TIMESTAMP":              FUNCTION,
	"ENUM_FIRST":                         FUNCTION,
	"ENUM_LAST":                          FUNCTION,
	"ENUM_RANGE":                         FUNCTION,
	"AREA":                               FUNCTION,
	"CENTER":                             FUNCTION,
	"DIAMETER":                           FUNCTION,
	"HEIGHT":                             FUNCTION,
	"ISCLOSED":                           FUNCTION,
	"ISOPEN":                             FUNCTION,
	"NPOINTS":                            FUNCTION,
	"PCLOSE":                             FUNCTION,
	"POPEN":                              FUNCTION,
	"RADIUS":                             FUNCTION,
	"WIDTH":                              FUNCTION,
	"BOX":                                FUNCTION,
	"CIRCLE":                             FUNCTION,
	"LSEG":                               FUNCTION,
	"PATH":                               FUNCTION,
	"POINT":                              FUNCTION,
	"POLYGON":                            FUNCTION,
	"ABBREV":                             FUNCTION,
	"BROADCAST":                          FUNCTION,
	"FAMILY":                             FUNCTION,
	"HOST":                               FUNCTION,
	"HOSTMASK":                           FUNCTION,
	"MASKLEN":                            FUNCTION,
	"NETMASK":                            FUNCTION,
	"NETWORK":                            FUNCTION,
	"SET_MASKLEN":                        FUNCTION,
	"TEXT":                               FUNCTION,
	"TRUNC":                              FUNCTION,
	"GET_CURRENT_TS_CONFIG":              FUNCTION,
	"NUMNODE":                            FUNCTION,
	"PLAINTO_TSQUERY":                    FUNCTION,
	"QUERYTREE":                          FUNCTION,
	"SETWEIGHT":                          FUNCTION,
	"STRIP":                              FUNCTION,
	"TO_TSQUERY":                         FUNCTION,
	"TO_TSVECTOR":                        FUNCTION,
	"TS_HEADLINE":                        FUNCTION,
	"TS_RANK":                            FUNCTION,
	"TS_RANK_CD":                         FUNCTION,
	"TS_REWRITE":                         FUNCTION,
	"TSVECTOR_UPDATE_TRIGGER":            FUNCTION,
	"TSVECTOR_UPDATE_TRIGGER_COLUMN":     FUNCTION,
	"TS_DEBUG":                           FUNCTION,
	"TS_LEXIZE":                          FUNCTION,
	"TS_PARSE":                           FUNCTION,
	"TS_TOKEN_TYPE":                      FUNCTION,
	"TS_STAT":                            FUNCTION,
	"XMLCOMMENT":                         FUNCTION,
	"XMLCONCAT":                          FUNCTION,
	"XMLELEMENT":                         FUNCTION,
	"XMLFOREST":                          FUNCTION,
	"XMLPI":                              FUNCTION,
	"XMLROOT":                            FUNCTION,
	"XMLAGG":                             FUNCTION,
	"XMLEXISTS":                          FUNCTION,
	"XML_IS_WELL_FORMED":                 FUNCTION,
	"XML_IS_WELL_FORMED_DOCUMENT":        FUNCTION,
	"XML_IS_WELL_FORMED_CONTENT":         FUNCTION,
	"XPATH":                              FUNCTION,
	"XPATH_EXISTS":                       FUNCTION,
	"TABLE_TO_XML":                       FUNCTION,
	"QUERY_TO_XML":                       FUNCTION,
	"CURSOR_TO_XML":                      FUNCTION,
	"TABLE_TO_XMLSCHEMA":                 FUNCTION,
	"QUERY_TO_XMLSCHEMA":                 FUNCTION,
	"CURSOR_TO_XMLSCHEMA":                FUNCTION,
	"TABLE_TO_XML_AND_XMLSCHEMA":         FUNCTION,
	"QUERY_TO_XML_AND_XMLSCHEMA":         FUNCTION,
	"SCHEMA_TO_XML":                      FUNCTION,
	"SCHEMA_TO_XMLSCHEMA":                FUNCTION,
	"SCHEMA_TO_XML_AND_XMLSCHEMA":        FUNCTION,
	"DATABASE_TO_XML":                    FUNCTION,
	"DATABASE_TO_XMLSCHEMA":              FUNCTION,
	"DATABASE_TO_XML_AND_XMLSCHEMA":      FUNCTION,
	"CURRVAL":                            FUNCTION,
	"LASTVAL":                            FUNCTION,
	"NEXTVAL":                            FUNCTION,
	"SETVAL":                             FUNCTION,
	"ARRAY_APPEND":                       FUNCTION,
	"ARRAY_CAT":                          FUNCTION,
	"ARRAY_NDIMS":                        FUNCTION,
	"ARRAY_DIMS":                         FUNCTION,
	"ARRAY_FILL":                         FUNCTION,
	"ARRAY_LENGTH":                       FUNCTION,
	"ARRAY_LOWER":                        FUNCTION,
	"ARRAY_PREPEND":                      FUNCTION,
	"ARRAY_TO_STRING":                    FUNCTION,
	"ARRAY_UPPER":                        FUNCTION,
	"STRING_TO_ARRAY":                    FUNCTION,
	"UNNEST":                             FUNCTION,
	"ARRAY_AGG":                          FUNCTION,
	"AVG":                                FUNCTION,
	"BIT_AND":                            FUNCTION,
	"BIT_OR":                             FUNCTION,
	"BOOL_AND":                           FUNCTION,
	"BOOL_OR":                            FUNCTION,
	"COUNT":                              FUNCTION,
	"EVERY":                              FUNCTION,
	"MAX":                                FUNCTION,
	"MIN":                                FUNCTION,
	"STRING_AGG":                         FUNCTION,
	"SUM":                                FUNCTION,
	"CORR":                               FUNCTION,
	"COVAR_POP":                          FUNCTION,
	"COVAR_SAMP":                         FUNCTION,
	"REGR_AVGX":                          FUNCTION,
	"REGR_AVGY":                          FUNCTION,
	"REGR_COUNT":                         FUNCTION,
	"REGR_INTERCEPT":                     FUNCTION,
	"REGR_R2":                            FUNCTION,
	"REGR_SLOPE":                         FUNCTION,
	"REGR_SXX":                           FUNCTION,
	"REGR_SXY":                           FUNCTION,
	"REGR_SYY":                           FUNCTION,
	"STDDEV":                             FUNCTION,
	"STDDEV_POP":                         FUNCTION,
	"STDDEV_SAMP":                        FUNCTION,
	"VARIANCE":                           FUNCTION,
	"VAR_POP":                            FUNCTION,
	"VAR_SAMP":                           FUNCTION,
	"ROW_NUMBER":                         FUNCTION,
	"RANK":                               FUNCTION,
	"DENSE_RANK":                         FUNCTION,
	"PERCENT_RANK":                       FUNCTION,
	"CUME_DIST":                          FUNCTION,
	"NTILE":                              FUNCTION,
	"LAG":                                FUNCTION,
	"LEAD":                               FUNCTION,
	"FIRST_VALUE":                        FUNCTION,
	"LAST_VALUE":                         FUNCTION,
	"NTH_VALUE":                          FUNCTION,
	"GENERATE_SERIES":                    FUNCTION,
	"GENERATE_SUBSCRIPTS":                FUNCTION,
	"CURRENT_DATABASE":                   FUNCTION,
	"CURRENT_QUERY":                      FUNCTION,
	"CURRENT_SCHEMA[":                    FUNCTION,
	"CURRENT_SCHEMAS":                    FUNCTION,
	"INET_CLIENT_ADDR":                   FUNCTION,
	"INET_CLIENT_PORT":                   FUNCTION,
	"INET_SERVER_ADDR":                   FUNCTION,
	"INET_SERVER_PORT":                   FUNCTION,
	"PG_BACKEND_PID":                     FUNCTION,
	"PG_CONF_LOAD_TIME":                  FUNCTION,
	"PG_IS_OTHER_TEMP_SCHEMA":            FUNCTION,
	"PG_LISTENING_CHANNELS":              FUNCTION,
	"PG_MY_TEMP_SCHEMA":                  FUNCTION,
	"PG_POSTMASTER_START_TIME":           FUNCTION,
	"VERSION":                            FUNCTION,
	"HAS_ANY_COLUMN_PRIVILEGE":           FUNCTION,
	"HAS_COLUMN_PRIVILEGE":               FUNCTION,
	"HAS_DATABASE_PRIVILEGE":             FUNCTION,
	"HAS_FOREIGN_DATA_WRAPPER_PRIVILEGE": FUNCTION,
	"HAS_FUNCTION_PRIVILEGE":             FUNCTION,
	"HAS_LANGUAGE_PRIVILEGE":             FUNCTION,
	"HAS_SCHEMA_PRIVILEGE":               FUNCTION,
	"HAS_SEQUENCE_PRIVILEGE":             FUNCTION,
	"HAS_SERVER_PRIVILEGE":               FUNCTION,
	"HAS_TABLE_PRIVILEGE":                FUNCTION,
	"HAS_TABLESPACE_PRIVILEGE":           FUNCTION,
	"PG_HAS_ROLE":                        FUNCTION,
	"PG_COLLATION_IS_VISIBLE":            FUNCTION,
	"PG_CONVERSION_IS_VISIBLE":           FUNCTION,
	"PG_FUNCTION_IS_VISIBLE":             FUNCTION,
	"PG_OPCLASS_IS_VISIBLE":              FUNCTION,
	"PG_OPERATOR_IS_VISIBLE":             FUNCTION,
	"PG_TABLE_IS_VISIBLE":                FUNCTION,
	"PG_TS_CONFIG_IS_VISIBLE":            FUNCTION,
	"PG_TS_DICT_IS_VISIBLE":              FUNCTION,
	"PG_TS_PARSER_IS_VISIBLE":            FUNCTION,
	"PG_TS_TEMPLATE_IS_VISIBLE":          FUNCTION,
	"PG_TYPE_IS_VISIBLE":                 FUNCTION,
	"FORMAT_TYPE":                        FUNCTION,
	"PG_DESCRIBE_OBJECT":                 FUNCTION,
	"PG_GET_CONSTRAINTDEF":               FUNCTION,
	"PG_GET_EXPR":                        FUNCTION,
	"PG_GET_FUNCTIONDEF":                 FUNCTION,
	"PG_GET_FUNCTION_ARGUMENTS":          FUNCTION,
	"PG_GET_FUNCTION_IDENTITY_ARGUMENTS": FUNCTION,
	"PG_GET_FUNCTION_RESULT":             FUNCTION,
	"PG_GET_INDEXDEF":                    FUNCTION,
	"PG_GET_KEYWORDS":                    FUNCTION,
	"PG_GET_RULEDEF":                     FUNCTION,
	"PG_GET_SERIAL_SEQUENCE":             FUNCTION,
	"PG_GET_TRIGGERDEF":                  FUNCTION,
	"PG_GET_USERBYID":                    FUNCTION,
	"PG_GET_VIEWDEF":                     FUNCTION,
	"PG_OPTIONS_TO_TABLE":                FUNCTION,
	"PG_TABLESPACE_DATABASES":            FUNCTION,
	"PG_TYPEOF":                          FUNCTION,
	"COL_DESCRIPTION":                    FUNCTION,
	"OBJ_DESCRIPTION":                    FUNCTION,
	"SHOBJ_DESCRIPTION":                  FUNCTION,
	"TXID_CURRENT":                       FUNCTION,
	"TXID_CURRENT_SNAPSHOT":              FUNCTION,
	"TXID_SNAPSHOT_XIP":                  FUNCTION,
	"TXID_SNAPSHOT_XMAX":                 FUNCTION,
	"TXID_SNAPSHOT_XMIN":                 FUNCTION,
	"TXID_VISIBLE_IN_SNAPSHOT":           FUNCTION,
	"CURRENT_SETTING":                    FUNCTION,
	"SET_CONFIG":                         FUNCTION,
	"PG_CANCEL_BACKEND":                  FUNCTION,
	"PG_RELOAD_CONF":                     FUNCTION,
	"PG_ROTATE_LOGFILE":                  FUNCTION,
	"PG_TERMINATE_BACKEND":               FUNCTION,
	"PG_CREATE_RESTORE_POINT":            FUNCTION,
	"PG_CURRENT_XLOG_INSERT_LOCATION":    FUNCTION,
	"PG_CURRENT_XLOG_LOCATION":           FUNCTION,
	"PG_START_BACKUP":                    FUNCTION,
	"PG_STOP_BACKUP":                     FUNCTION,
	"PG_SWITCH_XLOG":                     FUNCTION,
	"PG_XLOGFILE_NAME":                   FUNCTION,
	"PG_XLOGFILE_NAME_OFFSET":            FUNCTION,
	"PG_IS_IN_RECOVERY":                  FUNCTION,
	"PG_LAST_XLOG_RECEIVE_LOCATION":      FUNCTION,
	"PG_LAST_XLOG_REPLAY_LOCATION":       FUNCTION,
	"PG_LAST_XACT_REPLAY_TIMESTAMP":      FUNCTION,
	"PG_IS_XLOG_REPLAY_PAUSED":           FUNCTION,
	"PG_XLOG_REPLAY_PAUSE":               FUNCTION,
	"PG_XLOG_REPLAY_RESUME":              FUNCTION,
	"PG_COLUMN_SIZE":                     FUNCTION,
	"PG_DATABASE_SIZE":                   FUNCTION,
	"PG_INDEXES_SIZE":                    FUNCTION,
	"PG_RELATION_SIZE":                   FUNCTION,
	"PG_SIZE_PRETTY":                     FUNCTION,
	"PG_TABLE_SIZE":                      FUNCTION,
	"PG_TABLESPACE_SIZE":                 FUNCTION,
	"PG_TOTAL_RELATION_SIZE":             FUNCTION,
	"PG_RELATION_FILENODE":               FUNCTION,
	"PG_RELATION_FILEPATH":               FUNCTION,
	"PG_LS_DIR":                          FUNCTION,
	"PG_READ_FILE":                       FUNCTION,
	"PG_READ_BINARY_FILE":                FUNCTION,
	"PG_STAT_FILE":                       FUNCTION,
	"PG_ADVISORY_LOCK":                   FUNCTION,
	"PG_ADVISORY_LOCK_SHARED":            FUNCTION,
	"PG_ADVISORY_UNLOCK":                 FUNCTION,
	"PG_ADVISORY_UNLOCK_ALL":             FUNCTION,
	"PG_ADVISORY_UNLOCK_SHARED":          FUNCTION,
	"PG_ADVISORY_XACT_LOCK":              FUNCTION,
	"PG_ADVISORY_XACT_LOCK_SHARED":       FUNCTION,
	"PG_TRY_ADVISORY_LOCK":               FUNCTION,
	"PG_TRY_ADVISORY_LOCK_SHARED":        FUNCTION,
	"PG_TRY_ADVISORY_XACT_LOCK":          FUNCTION,
	"PG_TRY_ADVISORY_XACT_LOCK_SHARED":   FUNCTION,

	/*
	 * Additional SQL Functions
	 */
	"CHAR":              FUNCTION,
	"CHARINDEX":         FUNCTION,
	"DATALENGTH":        FUNCTION,
	"DIFFERENCE":        FUNCTION,
	"LEN":               FUNCTION,
	"NCHAR":             FUNCTION,
	"PATINDEX":          FUNCTION,
	"QUOTENAME":         FUNCTION,
	"REPLICATE":         FUNCTION,
	"SOUNDEX":           FUNCTION,
	"SPACE":             FUNCTION,
	"STR":               FUNCTION,
	"STUFF":             FUNCTION,
	"UNICODE":           FUNCTION,
	"ABS":               FUNCTION,
	"ACOS":              FUNCTION,
	"ASIN":              FUNCTION,
	"ATAN":              FUNCTION,
	"ATN2":              FUNCTION,
	"CEILING":           FUNCTION,
	"COS":               FUNCTION,
	"COT":               FUNCTION,
	"DEGREES":           FUNCTION,
	"EXP":               FUNCTION,
	"FLOOR":             FUNCTION,
	"LOG":               FUNCTION,
	"LOG10":             FUNCTION,
	"PI":                FUNCTION,
	"POWER":             FUNCTION,
	"RADIANS":           FUNCTION,
	"RAND":              FUNCTION,
	"ROUND":             FUNCTION,
	"SIGN":              FUNCTION,
	"SIN":               FUNCTION,
	"SQRT":              FUNCTION,
	"SQUARE":            FUNCTION,
	"TAN":               FUNCTION,
	"CURRENT_TIMESTAMP": FUNCTION,
	"DATEADD":           FUNCTION,
	"DATEDIFF":          FUNCTION,
	"DATEFROMPARTS":     FUNCTION,
	"DATENAME":          FUNCTION,
	"DATEPART":          FUNCTION,
	"DAY":               FUNCTION,
	"GETDATE":           FUNCTION,
	"GETUTCDATE":        FUNCTION,
	"ISDATE":            FUNCTION,
	"MONTH":             FUNCTION,
	"SYSDATETIME":       FUNCTION,
	"YEAR":              FUNCTION,
	"CAST":              FUNCTION,
	"COALESCE":          FUNCTION,
	"CURRENT_USER":      FUNCTION,
	"IIF":               FUNCTION,
	"ISNULL":            FUNCTION,
	"ISNUMERIC":         FUNCTION,
	"NULLIF":            FUNCTION,
	"SESSION_USER":      FUNCTION,
	"SESSIONPROPERTY":   FUNCTION,
	"SYSTEM_USER":       FUNCTION,
	"USER_NAME":         FUNCTION,

	/*
	 * Additional SQLite functions
	 */
	"CHANGES":                   FUNCTION,
	"GLOB":                      FUNCTION,
	"HEX":                       FUNCTION,
	"IFNULL":                    FUNCTION,
	"INSTR":                     FUNCTION,
	"LAST_INSERT_ROWID":         FUNCTION,
	"LIKELIHOOD":                FUNCTION,
	"LIKELY":                    FUNCTION,
	"LOAD_EXTENSION":            FUNCTION,
	"PRINTF":                    FUNCTION,
	"QUOTE":                     FUNCTION,
	"RANDOMBLOB":                FUNCTION,
	"SQLITE_COMPILEOPTION_GET":  FUNCTION,
	"SQLITE_COMPILEOPTION_USED": FUNCTION,
	"SQLITE_OFFSET":             FUNCTION,
	"SQLITE_SOURCE_ID":          FUNCTION,
	"SQLITE_VERSION":            FUNCTION,
	"TOTAL_CHANGES":             FUNCTION,
	"TYPEOF":                    FUNCTION,
	"UNHEX":                     FUNCTION,
	"UNLIKELY":                  FUNCTION,
	"ZEROBLOB":                  FUNCTION,
	"GROUP_CONCAT":              FUNCTION,
	"DATE":                      FUNCTION,
	"TIME":                      FUNCTION,
	"DATETIME":                  FUNCTION,
	"JULIANDAY":                 FUNCTION,
	"STRFTIME":                  FUNCTION,
	"RANDOM":                    FUNCTION,

	/*
	 * Additional Oracle functions
	 */
	"ATAN2":                        FUNCTION,
	"BITAND":                       FUNCTION,
	"COSH":                         FUNCTION,
	"LN":                           FUNCTION,
	"NANVL":                        FUNCTION,
	"REMAINDER":                    FUNCTION,
	"SINH":                         FUNCTION,
	"TANH":                         FUNCTION,
	"WIDTH_BUCKET":                 FUNCTION,
	"NLS_INITCAP":                  FUNCTION,
	"NLS_LOWER":                    FUNCTION,
	"NLSSORT":                      FUNCTION,
	"NLS_UPPER":                    FUNCTION,
	"REGEXP_SUBSTR":                FUNCTION,
	"TREAT":                        FUNCTION,
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
	"USERENV":                      FUNCTION,
	"VSIZE":                        FUNCTION,
	"XMLCOLATTVAL":                 FUNCTION,
	"XMLSEQUENCE":                  FUNCTION,
	"XMLTRANSFORM":                 FUNCTION,
	"CORR_S":                       FUNCTION,
	"CORR_K":                       FUNCTION,
	"GROUP_ID":                     FUNCTION,
	"GROUPING":                     FUNCTION,
	"GROUPING_ID":                  FUNCTION,
	"MEDIAN":                       FUNCTION,
	"PERCENTILE_CONT":              FUNCTION,
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
	"PERCENTILE_DISC":              FUNCTION,
}

var comparatorMap = map[string]TokenType{
	"~~":   COMPARATOR, // LIKE
	"~~*":  COMPARATOR, // ILIKE
	"!~~":  COMPARATOR, // NOT LIKE
	"!~~*": COMPARATOR, // NOT ILIKE
	"=":    COMPARATOR,
	"!=":   COMPARATOR, // NOT EQUAL
	"<>":   COMPARATOR, // NOT EQUAL
	">":    COMPARATOR,
	"<":    COMPARATOR,
	">=":   COMPARATOR,
	"<=":   COMPARATOR,
}

// peekComparator peeks into the subsequent characters trying to identify valid comparator substrings, but
// tries not to match on broken substrings that are not really valid comparators.
func peekComparator(r *bufio.Reader) (string, error) {

	// Peek step by step into subsequent characters to search for a valid comparator
	steps := 1
	sequence := ""
	for {
		b, errPeek := r.Peek(steps)
		if errPeek != nil {
			break
		}

		// Convert to string and get last character to analyze
		s := string(b)
		ch := s[len(s)-1]

		// Check if character is plausible comparator
		if strings.Contains("~*!=<>", string(ch)) {
			sequence += string(ch)
		} else {
			break
		}

		// Increment bytes to read
		steps++
	}

	// Check if read sequence is valid comparator
	if sequence != "" {

		// Return comparator sequence
		if _, ok := comparatorMap[sequence]; ok {
			return sequence, nil
		}

		// Return error if invalid comparator sequence was detected
		return "", fmt.Errorf("invalid comparator sequence: %s", sequence)
	}

	// Return empty string if sequence was not a comparator
	return "", nil
}

var punctuationMap = map[string]TokenType{
	"(": STARTPARENTHESIS,
	")": ENDPARENTHESIS,
	"[": STARTBRACKET,
	"]": ENDBRACKET,
	"{": STARTBRACE,
	"}": ENDBRACKET,
	",": COMMA,
	":": COLON,
}

func isPunctuation(ch rune) bool {
	_, is := punctuationMap[string(ch)]
	return is
}

func isNewline(ch rune) bool {
	return ch == '\n'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '　'
}

func isTab(ch rune) bool {
	return ch == '\t'
}

func isColon(ch rune) bool {
	return ch == ':'
}

func isParenthesisStart(ch rune) bool {
	return ch == '('
}

func isSingleQuote(ch rune) bool {
	return ch == '\''
}
