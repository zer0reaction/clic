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

	program := "((+ 3 4) (+ 3 45))"

	l.LoadString(program)

	list, err := parser.DebugChopList(&l)
	if err != nil {
		log.Fatal(err)
	}

	_, err = parser.DebugParseList(list)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Print(codegen.Codegen(root))
}
