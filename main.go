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
(exfun print)

(:= (let x) (- 1000 -7))

(
	(print x)
	(:= x (- x 6))

	(
		(let x)
		(:= x 34)
		(:= x (- x 6))
		(print x)
	)

	(:= (let x) 7)
	(print x)
	(:= x (- x 10000))
)

(print x)
`

	l.LoadString(program)

	root, err := parser.CreateAST(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
