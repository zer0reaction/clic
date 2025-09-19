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
((let foo)
 (set foo 34)

 (let bar)
 (set bar 35)

 (let res)
 (set res (+ foo bar)))
`

	l.LoadString(program)

	root, err := parser.DebugParseList(&l, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
