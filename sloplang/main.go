package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saad039/sloplang/pkg/codegen"
	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

//go:embed pkg/runtime/slop_value.go
var rtSlopValue string

//go:embed pkg/runtime/ops.go
var rtOps string

//go:embed pkg/runtime/io.go
var rtIO string

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
	gen := codegen.New()
	output, err := gen.Generate(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codegen error: %v\n", err)
		os.Exit(1)
	}

	// Assemble user code with inlined runtime into a single package main file
	assembled, err := codegen.AssembleWithRuntime(output, rtSlopValue, rtOps, rtIO)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble error: %v\n", err)
		os.Exit(1)
	}

	// Build in a temp directory — no external dependencies needed
	tmpDir, err := os.MkdirTemp("", "slop-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	goMod := "module sloprun\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing go.mod: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), assembled, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing main.go: %v\n", err)
		os.Exit(1)
	}

	// go build
	binaryPath := filepath.Join(tmpDir, "prog")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = tmpDir
	if buildOut, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, string(buildOut))
		os.Exit(1)
	}

	// Run the compiled binary from the original .slop file's directory
	// so that relative file I/O paths work correctly
	slopDir, err := filepath.Abs(filepath.Dir(inputPath))
	if err != nil {
		slopDir = filepath.Dir(inputPath)
	}
	runCmd := exec.Command(binaryPath)
	runCmd.Dir = slopDir
	runCmd.Stdin = os.Stdin
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Write generated Go source next to the .slop file (for inspection)
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".go"
	_ = os.WriteFile(outputPath, assembled, 0644)

	if err := runCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}
