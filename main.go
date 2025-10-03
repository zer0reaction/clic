package main

import (
	"flag"
	"fmt"
	"github.com/zer0reaction/lisp-go/codegen"
	"github.com/zer0reaction/lisp-go/parser"
	"github.com/zer0reaction/lisp-go/report"
	"os"
	"os/exec"
)

func main() {
	output := flag.String("o", "out", "Output executable file path")
	backend := flag.String("b", "gcc", "Compiler backend")
	backendFlags := flag.String("bf", "", "Backend flags")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("error: expected one file to compile\n")
		os.Exit(1)
	}
	input := flag.Args()[0]

	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("error: failed to open file %s\n", input)
		os.Exit(1)
	}

	p := parser.New(input, string(data))
	root := p.CreateAST()

	if report.ErrorsOccured() {
		fmt.Printf("errors occured, exiting\n")
		os.Exit(1)
	}

	asm := codegen.Codegen(root)
	asmPath := "/tmp/cli.s"
	err = os.WriteFile(asmPath, []byte(asm), 0666)
	if err != nil {
		fmt.Printf("error: failed to write to %s\n", asmPath)
		os.Exit(1)
	}

	var cmdFlags []string
	if len(*backendFlags) > 0 {
		cmdFlags = []string{"-o", *output, *backendFlags, asmPath}
	} else {
		cmdFlags = []string{"-o", *output, asmPath}
	}

	cmd := exec.Command(*backend, cmdFlags...)
	fmt.Printf("[backend] executing %s\n", cmd)

	backendOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", backendOutput)
		os.Exit(1)
	}
}
