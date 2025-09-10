package main

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"log"
)

func main() {
	var lexer lexer.Lexer

	log.SetFlags(0)

	lexer.LoadString("\n  ))  )\t)\n)   \n")

	lexer.DebugCacheToken()
	lexer.DebugCacheToken()
	lexer.DebugCacheToken()
	lexer.DebugCacheToken()
	lexer.DebugCacheToken()

	for i := uint(0); i < lexer.GetCachedCount(); i++ {
		tp, err := lexer.PeekToken(i)
		if err != nil {
			log.Fatal(err)
		}
		tp.PrintInfo()
	}

	fmt.Println(lexer.GetCachedCount())

	tp, err := lexer.PeekToken(5)
	if err != nil {
		log.Fatal(err)
	}
	tp.PrintInfo()
}
