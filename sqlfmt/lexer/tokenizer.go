package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// DisableFunctionKeywords - Postgres has a few functions without parenthesis. They look like normal keywords,
// but they might conflict with table/column names. To address ambiguity, Postgres clients must therefore
// wrap affected names with double quotes. Otherwise, a name might be understood as a function keyword. In
// other dialects clients wouldn't care and quote these names. Disable keyword functions to avoid ambiguous
// names to be capitalized like function names.
// Affected names: LOCALTIME, LOCALTIMESTAMP, CURRENT_DATE, CURRENT_TIME, CURRENT_TIMESTAMP, CURRENT_USER,
// CURRENT_CATALOG, SESSION_USER, USER
var DisableFunctionKeywords = false

// Tokenize sql string and returns slice of Token. Ignores Token of white-space, new-line and tab, as
// they have no semantic meaning.
func Tokenize(sql string) ([]Token, error) {

	// Prepare tokenizer
	t := &tokenizer{
		r: bufio.NewReader(strings.NewReader(sql)),
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

// tokenizer holds a working buffer to process and defines functions to execute against it
type tokenizer struct {
	r *bufio.Reader
}

// scan reads the first character of the buffer and, depending on it, proceeds to read additional ones until a
// full token is detected and returns it
func (t *tokenizer) scan() (Token, error) {

	// Peek if next characters represent a valid comparator. If so, read the according amount of bytes
	// from the buffer and return comparator token
	if comparatorNext, _ := peekComparator(t.r); comparatorNext != "" {
		for i := len(comparatorNext); i > 0; i-- {
			_, _, _ = t.r.ReadRune()
		}
		return Token{Type: COMPARATOR, Value: comparatorNext}, nil
	}

	// Read first character from buffer
	ch, _, errCh := t.r.ReadRune()
	if errCh != nil {
		if errCh.Error() == "EOF" {
			return Token{Type: EOF, Value: "EOF"}, nil
		}
		return Token{}, errCh
	}

	// Prepare buffer for token and write read character to it
	var buf bytes.Buffer
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
				_ = t.r.UnreadRune() // Revert last read, because it belonged to the next token
				return Token{Type: COLON, Value: buf.String()}, nil
			}
		}

		// Lookup other token type in punctuation map
		if ttype, ok := punctuationMap[buf.String()]; ok {
			return Token{Type: ttype, Value: buf.String()}, nil
		}

		// Return with error in case of unexpected value
		return Token{}, fmt.Errorf("invalid punctuation value: %v", buf.String())

	case isSlash(ch):

		// Check if next character opens comment and read it
		comment, errComment := t.readComment(&buf, ch)
		if errComment != nil {
			return Token{}, errComment
		}

		// Return comment if one was read
		if comment != "" {
			return Token{Type: COMMENT, Value: comment}, nil
		}

		// Continue after select with reading other tokens otherwise
		break

	case isSingleQuote(ch):

		// Read subsequent characters until closing single quote
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

			// Break loop once closing quote is found
			if isSingleQuote(chNext) {
				break
			}
		}

		// Return string token
		return Token{Type: STRING, Value: buf.String()}, nil
	}

	// Read subsequent characters until value is complete
	var comparator = ""
	var comparatorErr error
	for {

		// Stop if next character starts comparator sequence. But only if previous check didn't return
		// an invalid comparator sequence, otherwise an invalid comparator might turn into a valid one
		// after reading further bytes. For example, ~~~ might be understood as ~~. An input like 'a~~~1'
		// would be interpreted as 'a~ ~~1'.
		if comparatorErr == nil {
			comparator, comparatorErr = peekComparator(t.r)
			if comparator != "" {
				break // Nothing was read yet, no need to unread
			}
		}

		// Read next character
		chNext, _, errNext := t.r.ReadRune()
		if errNext != nil {
			if errNext.Error() == "EOF" {
				break
			} else {
				return Token{}, errNext
			}
		}

		// Stop if next character doesn't belong to the value anymore. Unread last unnecessary character.
		if isPunctuation(chNext) || isSingleQuote(chNext) || isWhitespace(chNext) || isNewline(chNext) || isTab(chNext) {
			_ = t.r.UnreadRune()
			break
		}

		// Append character to value
		buf.WriteRune(chNext)
	}

	// Prepare default lookup key and token value
	key := strings.ToUpper(buf.String())
	val := key

	// Sanitize key and value, if they include a target operator '.'.
	// If token value contains period, it's specifying a target, e.g. a table. Put that aside for the lookup.
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
	if ttype, ok := keywordMap[val]; ok { // Use val instead of key, because "." should not be splitted for keyword lookups!

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

// readComment looks into the subsequent characters to identify a comment start sequence and reads until its end
func (t *tokenizer) readComment(buf *bytes.Buffer, chPrev rune) (string, error) {

	// Check kind of comment
	singleLine := false
	if t.peekSubsequent(isSlash) {
		singleLine = true // one-line comment //
	} else if t.peekSubsequent(isWildcard) {
		singleLine = false // multi-line comment /* ... */
	} else {
		return "", nil
	}

	// Read subsequent characters until closing single quote
	for {
		chNext, _, errNext := t.r.ReadRune()
		if errNext != nil {
			if singleLine && errNext.Error() == "EOF" {
				return buf.String(), nil
			} else {
				return buf.String(), errNext
			}
		}

		// Stop reading single-line comment at new line
		if singleLine && isNewline(chNext) {
			return buf.String(), nil
		}

		// Append character to value
		buf.WriteRune(chNext)

		// Stop reading multi-line comment at termination sequence
		if !singleLine && isWildcard(chPrev) && isSlash(chNext) {
			return buf.String(), nil
		}

		// Remember last ch
		chPrev = chNext
	}
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

func isSlash(ch rune) bool {
	return ch == '/'
}

func isWildcard(ch rune) bool {
	return ch == '*'
}
