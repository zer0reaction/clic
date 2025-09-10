package main

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
	"log"
)

func main() {
	var lexer lexer.Lexer

	log.SetFlags(0)
	lexer.LoadString("(((((((((((((((123 -345((((((((((((((((((((((((((((((((((")

	for i := uint(0); i < 5; i++ {
		fmt.Println("----------------------------------------")

		tp, err := lexer.PeekToken(i)
		if err != nil {
			log.Fatal(err)
		}

		tp.PrintInfo()

		if tp.TableId != symbol.IdNone {
			err := symbol.PrintInfo(tp.TableId)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	_, _ = lexer.DebugReadToken()
	_, _ = lexer.DebugReadToken()
	_, _ = lexer.DebugReadToken()
	_, _ = lexer.DebugReadToken()
	_, _ = lexer.DebugReadToken()

	for i := uint(0); ; i++ {
		fmt.Println("----------------------------------------")

		tp, err := lexer.PeekToken(i)
		if err != nil {
			log.Fatal(err)
		}

		tp.PrintInfo()

		if tp.TableId != symbol.IdNone {
			err := symbol.PrintInfo(tp.TableId)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
