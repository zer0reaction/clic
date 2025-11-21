package main

import (
	"clic/checker"
	"clic/codegen"
	"clic/parser"
	"clic/report"
	"clic/symbol"
	"fmt"
	"flag"
	"os"
)

func main() {
	outFlag := flag.String("o", "out.s", "Assembly output path")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Usage: clic [-o outfile] infile")
		flag.PrintDefaults()
		os.Exit(1)
	}

	input := flag.Args()[0]
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t := &symbol.Table{}
	r := &report.Reporter{FileName: input}
	p := parser.New(string(data), t, r)

	asts := p.CreateASTs()
	r.ExitOnErrors(1)

	checker.TypeCheck(asts, t, r)
	r.ExitOnErrors(1)

	asm := codegen.Codegen(asts, t)
	err = os.WriteFile(*outFlag, []byte(asm), 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
