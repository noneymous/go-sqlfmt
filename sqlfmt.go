package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/reindenters"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// Main operation modes
	list    = flag.Bool("l", false, "list files whose formatting differs from goreturns's")
	write   = flag.Bool("w", false, "write result to (source) file instead of stdout")
	doDiff  = flag.Bool("d", false, "display diffs instead of rewriting files")
	options = &reindenters.Options{
		Padding:    "",
		Indent:     "  ",
		Newline:    "\n",
		Whitespace: " ",
	}
)

func init() {
	flag.StringVar(&options.Padding, "padding", "", "define a string to use for left-padding the output.")
	flag.StringVar(&options.Indent, "indent", "", "define a string to use for indentation.")
	flag.StringVar(&options.Newline, "newline", "", "define a string to use for line breaks.")
	flag.StringVar(&options.Whitespace, "whitespace", "", "define a string to use as a whitespace between values.")
}

func main() {

	// Initialize
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Usage = usage
	flag.Parse()

	// Process if user is piping source into go-sqlfmt
	if flag.NArg() == 0 {
		if info, _ := os.Stdin.Stat(); !isPipeInput(info) {
			flag.Usage()
			return
		}
		if *write {
			log.Fatal("can not use -w while using pipeline")
		}
		if err := processFile("<standard input>", os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Process according to command line arguments
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, errStat := os.Stat(path); {
		case errStat != nil:
			log.Fatal(errStat)
		case dir.IsDir():
			walkDir(path)
		default:
			if isGoFile(dir) {
				errParse := processFile(path, nil, os.Stdout)
				if errParse != nil {
					log.Fatal(errParse)
				}
			}
		}
	}
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: sqlfmt [flags] [path ...]\n")
	flag.PrintDefaults()
}

func isPipeInput(info os.FileInfo) bool {
	return info.Mode()&os.ModeCharDevice == 0
}

func processFile(filename string, in io.Reader, out io.Writer) error {
	if in == nil {
		f, errOpen := os.Open(filename)
		if errOpen != nil {
			return errOpen
		}
		in = f
	}

	src, errRead := io.ReadAll(in)
	if errRead != nil {
		return errRead
	}

	res, errParse := sqlfmt.FormatFile(filename, src, options)
	if errParse != nil {
		return errParse
	}

	if !bytes.Equal(src, res) {
		if *list {
			_, _ = fmt.Fprintln(out, filename)
		}
		if *write {
			if errRead = os.WriteFile(filename, res, 0); errRead != nil {
				return errRead
			}
		}
		if *doDiff {
			data, errDiff := diff(src, res)
			if errDiff != nil {
				return errDiff
			}
			fmt.Printf("diff %s gofmt/%s\n", filename, filename)
			_, _ = out.Write(data)
		}
		if !*list && !*write && !*doDiff {
			_, errRead = out.Write(res)
			if errRead != nil {
				return errRead
			}
		}
	}
	return nil
}

func walkDir(path string) {
	_ = filepath.Walk(path, visitFile)
}

func isGoFile(info os.FileInfo) bool {
	name := info.Name()
	return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func visitFile(path string, info os.FileInfo, err error) error {
	if err == nil && isGoFile(info) {
		err = processFile(path, nil, os.Stdout)
	}
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// diff compares two byte sequences and generates a diff report utilizing temporary files
func diff(b1, b2 []byte) ([]byte, error) {
	f1, errCreate := os.CreateTemp("", "sqlfmt")
	if errCreate != nil {
		return nil, errCreate
	}
	defer func() { _ = os.Remove(f1.Name()) }()
	defer func() { _ = f1.Close() }()

	f2, errCreate2 := os.CreateTemp("", "sqlfmt")
	if errCreate2 != nil {
		return nil, errCreate2
	}
	defer func() { _ = os.Remove(f2.Name()) }()
	defer func() { _ = f2.Close() }()

	_, _ = f1.Write(b1)
	_, _ = f2.Write(b2)

	data, errCommand := exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) == 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		return nil, errCommand
	}
	return data, nil
}
