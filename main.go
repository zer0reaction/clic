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
(exfun print)

(let s64 foo)
(:= foo 25)

(let s64 bar)
(:= bar (+ foo 234))

(let s64 i)
(:= i 0)

(print i)
(:= i (+ i 1))
(print i)
(:= i (+ i 1))
(print i)
(:= i (+ i 1))
(print i)
`

	l.LoadString(program)

	root, err := parser.CreateAST(&l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(codegen.Codegen(root))
}
