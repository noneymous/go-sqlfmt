package sqlfmt

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/noneymous/go-sqlfmt/sqlfmt/parser"
	"github.com/noneymous/go-sqlfmt/sqlfmt/parser/group"
	"strings"
)

// Format parse tokens, and build
func Format(sql string, options *Options) (string, error) {

	// Tokenize SQL query string
	tokens, errTokens := lexer.Tokenize(sql)
	if errTokens != nil {
		return "", fmt.Errorf("tokenization error: %w", errTokens)
	}

	// Parse tokens and group them into a sequence of query segments
	tokensParsed, errTokensParsed := parser.ParseTokens(tokens)
	if errTokensParsed != nil {
		return "", fmt.Errorf("parse error: %w", errTokensParsed)
	}

	// Format parsed tokens into prettified and uniformly formatted SQL string
	sqlFormatted, errFormat := generateFormattedStmt(tokensParsed, options.Distance)
	if errFormat != nil {
		return "", fmt.Errorf("format error: %w", errFormat)
	}

	// Safety check, compare if formatted query still has the same logic as input
	if !compare(sql, sqlFormatted) {
		return "", fmt.Errorf("an internal error has occurred")
	}

	// Return successfully formatted SQL string
	return sqlFormatted, nil
}

// generateFormattedStmt turns a sequence of parsed token groups into a formatted string.
func generateFormattedStmt(tokensParsed []group.Reindenter, leftPadding int) (string, error) {

	// Prepare buffer
	var buf bytes.Buffer

	// Iterate parsed tokens and reindent accordingly
	for _, token := range tokensParsed {
		if err := token.Reindent(&buf); err != nil {
			return "", err
		}
	}

	// Get formatted SQL string
	sqlFormatted := buf.String()

	// Add left spacing if desired
	if leftPadding == 0 {
		sqlFormatted = putDistance(sqlFormatted, leftPadding)
	}

	// Return generated string
	return sqlFormatted, nil
}

func putDistance(s string, leftPadding int) string {

	// Prepare result string
	var result string

	// Prepare scanner
	scanner := bufio.NewScanner(strings.NewReader(s))

	// Scan over input string and format accordingly
	for scanner.Scan() {
		result += fmt.Sprintf("%s%s%s", strings.Repeat(group.WhiteSpace, leftPadding), scanner.Text(), "\n")
	}

	// Return result
	return result
}

// compare returns false if the value of formatted statement (without any space, tab or newline symbols)
// differs from source statement
func compare(src string, res string) bool {

	// Unify inputs
	before := removeSymbol(src)
	after := removeSymbol(res)

	// Compare strings
	if v := strings.Compare(before, after); v != 0 {
		return false
	}

	// Return true if strings were equal
	return true
}

// removeSymbol removes whitespaces, tabs and newlines from string
func removeSymbol(s string) string {
	var result []rune
	for _, r := range s {
		if string(r) == "\n" || string(r) == " " || string(r) == "\t" || string(r) == "　" {
			continue
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
