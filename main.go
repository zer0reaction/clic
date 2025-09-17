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
(+ 3 4)
(+ 3 (+ 34 9))
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

	_, err = parser.DebugParseList(list1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = parser.DebugParseList(list2)
	if err != nil {
		log.Fatal(err)
	}
}
