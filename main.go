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
(exfun print_u64)

(set (let x) 5)

((print_u64 x)

 ((let x)
  (set x 34)
  (print_u64 x))

 (set (let x) 7)
 (print_u64 x))

(print_u64 x)
`

	l.LoadString(program)

	root, err := parser.CreateAST(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
