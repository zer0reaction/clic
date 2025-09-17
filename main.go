package main

import (
	// "fmt"
	// "github.com/zer0reaction/lisp-go/codegen"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/parser"
	"log"
)

func main() {
	var l lexer.Lexer
	log.SetFlags(0)

	program :=
		`
(
  (+ 3 4)
  (+ (+ 3 4)
     (+ 8 -9))
)
`

	l.LoadString(program)

	_, err := parser.DebugParseList(&l)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Print(codegen.Codegen(root))
}
