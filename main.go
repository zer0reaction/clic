package main

import (
	_ "fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	_ "github.com/zer0reaction/lisp-go/symbol"
	"log"
)

func main() {
	var l lexer.Lexer

	log.SetFlags(0)
	l.LoadString("(+ 34 35)")

	t, err := l.Match(lexer.TokenRbrOpen)
	if err != nil {
		log.Fatal(err)
	}
	t.PrintInfo()

	t, err = l.Match(lexer.TokenPlus)
	if err != nil {
		log.Fatal(err)
	}
	t.PrintInfo()
}
