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

(:= (let x) -5)

( (print x)

  ( (let x)
    (:= x 34)
    (print x) )

  (:= (let x) 7)
  (print x)      )

(print x)
`

	l.LoadString(program)

	root, err := parser.CreateAST(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
