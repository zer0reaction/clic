package main

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/codegen"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/parser"
	"log"
)

func main() {
	var l lexer.Lexer
	log.SetFlags(0)

	program :=
		`
(+ (+ 33 1)
   (+ 34 1))

(+ (+ 32 2)
   (+ 33 2))
`

	l.LoadString(program)
	root, err := parser.Parse(&l)
	if err != nil {
		log.Fatal(err)
	}

	code := codegen.Codegen(root)
	fmt.Println(code)
}
