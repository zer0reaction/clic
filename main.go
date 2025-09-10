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
	_, _ = lexer.DebugReadToken()
	lexer.DebugCacheToken()
	lexer.DebugCacheToken()
	_, _ = lexer.DebugReadToken()
	lexer.DebugCacheToken()

	fmt.Println(lexer.GetCachedCount())
}
