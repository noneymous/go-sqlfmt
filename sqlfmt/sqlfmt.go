package sqlfmt

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/formatters"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/noneymous/go-sqlfmt/sqlfmt/parser"
	"strings"
)

// Format parse tokens, and build
func Format(sql string, options *formatters.Options) (string, error) {

	// Tokenize SQL query string
	tokens, errTokenize := lexer.Tokenize(sql)
	if errTokenize != nil {
		return "", fmt.Errorf("tokenization error: %w", errTokenize)
	}

	// Parse tokens and group them into a sequence of query segments
	tokensParsed, errParse := parser.Parse(tokens, options)
	if errParse != nil {
		return "", fmt.Errorf("parse error: %w", errParse)
	}

	// Format parsed tokens into prettified and uniformly formatted SQL string
	var buf bytes.Buffer
	for i, tokenParsed := range tokensParsed {
		if err := tokenParsed.Format(&buf, tokensParsed, i); err != nil {
			return "", err
		}
	}

	// Get formatted SQL string
	sqlFormatted := strings.Trim(buf.String(), "\n")

	// Add left spacing if desired
	if options.Padding != "" {
		sqlFormatted = addPadding(sqlFormatted, options.Padding)
	}

	// Safety check, compare if formatted query still has the same logic as input
	if !CompareSemantic(sql, sqlFormatted) {
		fmt.Println(sqlFormatted)
		return "", fmt.Errorf("formatted result does not match input semantically")
	}

	// Return successfully formatted SQL string
	return sqlFormatted, nil
}

// CompareSemantic compares a formatted SQL string with the original input and checks whether they are
// logically still the same.
func CompareSemantic(sql string, formattedSql string) bool {

	// Unify inputs
	before := removeSymbols(removeComments(sql))
	after := removeSymbols(removeComments(formattedSql))

	// CompareSemantic strings
	if v := strings.Compare(before, after); v != 0 {
		return false
	}

	// Return true if strings were equal
	return true
}

// addPadding adds desired left-side padding to each line of the string
func addPadding(s string, leftPadding string) string {

	// Prepare result string
	var result []string

	// Prepare scanner
	scanner := bufio.NewScanner(strings.NewReader(s))

	// Scan over input string and format accordingly
	for scanner.Scan() {
		result = append(result, fmt.Sprintf("%s%s", leftPadding, scanner.Text()))
	}

	// Return result
	return strings.Join(result, "\n")
}

// removeComments removes one-line comments to make sure their formatting did not manipulate semantics,
// e.g. by putting a part of the SQL query at the end of a one-line comment where it would be ignored.
func removeComments(str string) string {
	var strNew string
	var quoted bool
	var skip bool
	for i, c := range str {

		// Don't remove quoted strings, even if they might contain comment indicators
		if !quoted && c == '\'' {
			quoted = true
			strNew += string(c)
			continue
		} else if quoted && c == '\'' {
			quoted = false
			strNew += string(c)
			continue
		} else if quoted {
			strNew += string(c)
			continue
		}

		// Check for start of comment
		if !skip {

			// Remove slash comment
			if c == '/' && len(str) > i+1 && str[i+1] == '/' {
				skip = true
				continue
			}

			// Remove dash comment
			if c == '-' && len(str) > i+1 && str[i+1] == '-' {
				skip = true
				continue
			}
		}

		// Remove until end of line
		if skip && c == '\n' {
			skip = false
		}

		// Add character if not within comment
		if !skip {
			strNew += string(c)
		}
	}
	return strNew
}

// removeSymbols removes semantically unnecessary characters, such as whitespaces, tabs and newlines, for comparison
func removeSymbols(s string) string {
	var result []rune
	for _, r := range s {
		if string(r) == "\n" || string(r) == " " || string(r) == "\t" || string(r) == "　" {
			continue
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
