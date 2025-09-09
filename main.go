package main

import (
	"log"
	"github.com/zer0reaction/lisp-go/lexer"
)

func main() {
	var lexer lexer.Lexer

	log.SetFlags(0)

	lexer.LoadString("\n  \n  (")
}
