package main

import (
	// "fmt"
	// "github.com/zer0reaction/lisp-go/codegen"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/parser"
	"log"
)

func main() {
	l := lexer.Lexer{}
	log.SetFlags(0)

	program :=
		`
(
  (let foo)
)
(let foo)
`

	l.LoadString(program)

	_, err := parser.DebugParseList(&l)
	if err != nil {
		log.Fatal(err)
	}
	_, err = parser.DebugParseList(&l)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Print(codegen.Codegen(root))
}
