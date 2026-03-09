package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saad039/sloplang/pkg/codegen"
	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: slop <file.slop>\n")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	source, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(source))
	tokens := l.Tokenize()

	// Parse
	p := parser.New(tokens)
	program, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		os.Exit(1)
	}

	// Generate
	gen := codegen.New("github.com/saad039/sloplang")
	output, err := gen.Generate(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codegen error: %v\n", err)
		os.Exit(1)
	}

	// Write output
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".go"
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("Transpiled %s -> %s\n", inputPath, outputPath)
}
