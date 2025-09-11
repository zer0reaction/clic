package main

import (
	_ "fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	_ "github.com/zer0reaction/lisp-go/symbol"
	"github.com/zer0reaction/lisp-go/parser"
	"log"
)

func main() {
	var l lexer.Lexer
	log.SetFlags(0)

	l.LoadString("(+ 34 (+ 34 -33))")
	_, err := parser.List(&l)
	if err != nil {
		log.Fatal(err)
	}
}
