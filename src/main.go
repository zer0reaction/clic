package main

import (
	"flag"
	"fmt"
	"lisp-go/src/codegen"
	"lisp-go/src/parser"
	"lisp-go/src/report"
	"os"
	"os/exec"
	"strings"
)

func main() {
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

	roots := p.CreateASTs()
	report.ExitOnErrors(1)

	p.TypeCheck(roots)
	report.ExitOnErrors(1)

	asm := codegen.Codegen(roots)
	asmPath := "/tmp/cli.s"
	err = os.WriteFile(asmPath, []byte(asm), 0666)
	if err != nil {
		fmt.Printf("error: failed to write to %s\n", asmPath)
		os.Exit(1)
	}

	var cmdFlags []string
	for _, el := range strings.Split(*backendFlags, " ") {
		cmdFlags = append(cmdFlags, el)
	}
	cmdFlags = append(cmdFlags, asmPath)

	cmd := exec.Command(*backend, cmdFlags...)
	fmt.Printf("[backend] executing %s\n", cmd)

	backendOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", backendOutput)
		os.Exit(1)
	}
}
