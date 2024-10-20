package parser

import (
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
)

// ParseTokens parses a sequence of tokens returning a slice of Reindenter.
// Each Reindenter is an SQL segment (group of SQL clauses) such as SelectGroup, FromGroup, etc..
func ParseTokens(tokens []lexer.Token) ([]lexer.Reindenter, error) {

	// Check if tokenized string is actually an SQL query
	if !isSql(tokens[0].Type) {
		return nil, fmt.Errorf("invalid sql statement")
	}

	// Prepare retriever for segment
	retriever, errRetriever := NewRetriever(tokens)
	if errRetriever != nil {
		return nil, errRetriever
	}

	// Process tokens
	result, errProcess := retriever.Process()
	if errProcess != nil {
		return nil, errProcess
	}

	// Return process result
	return result, nil
}

// isSql returns true if token type is valid SQL opening keyword
func isSql(ttype lexer.TokenType) bool {
	return ttype == lexer.SELECT || ttype == lexer.UPDATE || ttype == lexer.DELETE || ttype == lexer.INSERT || ttype == lexer.LOCK || ttype == lexer.WITH || ttype == lexer.SET
}
