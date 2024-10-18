package sqlfmt

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
)

// Options for go-sqlfmt
type Options struct {
	Distance int
}

// Process formats SQL statement in .go file
func Process(filename string, src []byte, options *Options) ([]byte, error) {

	// Prepare file set
	fileSet := token.NewFileSet()

	// Parse files
	f, errParse := parser.ParseFile(fileSet, filename, src, parser.ParseComments)
	if errParse != nil {
		return nil, errParse
	}

	// Replace ast nodes in the file with formatted SQL
	Replace(f, options)

	// Prepare
	var buf bytes.Buffer

	if errPrint := printer.Fprint(&buf, fileSet, f); errPrint != nil {
		return nil, errPrint
	}

	// Format buffer
	out, errSource := format.Source(buf.Bytes())
	if errSource != nil {
		return nil, errSource
	}

	// Return output
	return out, nil
}
