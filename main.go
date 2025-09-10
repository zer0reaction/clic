package main

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"log"
)

func main() {
	var lexer lexer.Lexer

	log.SetFlags(0)

	lexer.LoadString("  ()\n\n(  )\n))))")

	for {
		fmt.Println("--------------------------------------------------------------------------------")
		err := lexer.DebugCacheToken()
		if err != nil {
			log.Fatal(err)
		}

		token, err := lexer.DebugReadToken()
		if err != nil {
			log.Fatal(err)
		}

		token.PrintInfo()
	}
}
