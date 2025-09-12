package main

import (
	_ "fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/parser"
	_ "github.com/zer0reaction/lisp-go/symbol"
	"log"
)

func main() {
	var l lexer.Lexer
	log.SetFlags(0)

	program :=
		`
(+ (+ 33 1)
   (+ 34 1))
`

	l.LoadString(program)
	_, err := parser.List(&l)
	if err != nil {
		log.Fatal(err)
	}
}
