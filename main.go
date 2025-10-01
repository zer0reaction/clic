package main

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/codegen"
	// "github.com/zer0reaction/lisp-go/symbol"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/parser"
	"log"
)

func main() {
	l := lexer.Lexer{}
	log.SetFlags(0)

	program :=
		`
		(exfun print_s64)

		(let s64 foo)
		(:= foo 1337)
		(
			(print_s64 foo)
			(let s64 foo)
			(:= foo 1234)
			(print_s64 foo)
		)
		(:= foo 456)

		(print_s64 foo)
`

	l.LoadString(program)

	root, err := parser.CreateAST(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
