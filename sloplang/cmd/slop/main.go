package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saad039/sloplang/pkg/codegen"
	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

// findModuleRoot walks up from dir looking for go.mod with the sloplang module.
func findModuleRoot(dir string) (string, error) {
	for {
		goMod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goMod); err == nil {
			if strings.Contains(string(data), "module github.com/saad039/sloplang") {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot find sloplang module root (go.mod) from %s", dir)
		}
		dir = parent
	}
}

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

	// Write generated Go source next to the .slop file
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".go"
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	// Find the sloplang module root for the replace directive
	var moduleRoot string
	if envRoot := os.Getenv("SLOP_MODULE_ROOT"); envRoot != "" {
		moduleRoot = envRoot
	} else {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error finding executable path: %v\n", err)
			os.Exit(1)
		}
		moduleRoot, err = findModuleRoot(filepath.Dir(exePath))
		if err != nil {
			// Fallback: try from current working directory
			cwd, _ := os.Getwd()
			moduleRoot, err = findModuleRoot(cwd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Build in a temp directory with a go.mod that references the local module
	tmpDir, err := os.MkdirTemp("", "slop-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	goMod := fmt.Sprintf("module sloprun\n\ngo 1.24\n\nrequire github.com/saad039/sloplang v0.0.0\n\nreplace github.com/saad039/sloplang => %s\n", moduleRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp go.mod: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing temp main.go: %v\n", err)
		os.Exit(1)
	}

	// go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if tidyOut, err := tidyCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go mod tidy failed: %v\n%s\n", err, string(tidyOut))
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
	if err := runCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}
