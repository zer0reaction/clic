package main

import (
	"fmt"
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

(() ())
`
	l.LoadString(program)

	list1, err := parser.DebugChopList(&l)
	if err != nil {
		log.Fatal(err)
	}

	list2, err := parser.DebugChopList(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(list1.DebugCount())
	fmt.Println(list2.DebugCount())
}
