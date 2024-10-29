package sqlfmt

import (
	"bytes"
	"github.com/noneymous/go-sqlfmt/sqlfmt/formatters"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strings"
)

// FormatFile formats SQL statement in .go file
func FormatFile(filename string, src []byte, options *formatters.Options) ([]byte, error) {

	// Prepare file set
	fileSet := token.NewFileSet()

	// Parse files
	f, errParse := parser.ParseFile(fileSet, filename, src, parser.ParseComments)
	if errParse != nil {
		return nil, errParse
	}

	// Inspect file, search applicable SQL strings, format them accordingly and replace them
	ast.Inspect(f, func(n ast.Node) bool {
		sql, found := astFind(n)
		if found {

			// Clean string from quotation
			quoteChar := ""
			if strings.HasPrefix(sql, "`") && strings.HasSuffix(sql, "`") {
				quoteChar = "`"
				sql = strings.Trim(sql, "`")
			} else if strings.HasPrefix(sql, "'") && strings.HasSuffix(sql, "'") {
				quoteChar = "'"
				sql = strings.Trim(sql, "'")
			} else if strings.HasPrefix(sql, `"`) && strings.HasSuffix(sql, `"`) {
				quoteChar = `"`
				sql = strings.Trim(sql, `"`)
			}

			// Format SQL string
			sqlFormatted, errFormat := Format(sql, options)
			if errFormat != nil {

				// Log invalid queries for debugging
				log.Println(errFormat)
				log.Println(strings.Trim(strings.Trim(sql, "\n"), " "))
			} else {
				astReplace(n, quoteChar+sqlFormatted+quoteChar)
			}
		}
		return true
	})

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

// astFind searches an ast node for SQL strings as function arguments to be formatted
func astFind(n ast.Node) (string, bool) {
	ce, ok := n.(*ast.CallExpr)
	if !ok {
		return "", false
	}
	se, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	// check func name
	ok = astTarget(se.Sel.Name)
	if !ok {
		return "", false
	}

	// check length of the parameter
	// this is not for parsing "url.Query()"
	// FIXME: very adhoc
	if len(ce.Args) == 0 {
		return "", false
	}

	// SQL statement should appear in the first parameter
	arg, ok := ce.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}

	// Return value
	return arg.Value, true
}

// astTarget checks whether the function name is part of the intended functions.
// E.g., "Exec(string, ... any type)", "Query(string, ... any type)" and "QueryRow(string, ... any type)"
func astTarget(name string) bool {
	switch name {
	case "Exec", "Query", "QueryRow":
		return true
	}
	return false
}

// astReplace replaces an ast node with the given SQL string
func astReplace(n ast.Node, sql string) {
	replaceFunc := func(cr *astutil.Cursor) bool {
		switch crNode := cr.Node().(type) {
		case *ast.CallExpr:
			b := crNode.Args[0].(*ast.BasicLit)
			b.Value = sql
		}
		return true
	}
	astutil.Apply(n, replaceFunc, nil)
}
