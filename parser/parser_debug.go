package parser

import (
	"github.com/zer0reaction/lisp-go/lexer"
)

func (ls *list) DebugCount() uint {
	return ls.count()
}

func DebugChopList(lx *lexer.Lexer) (*list, error) {
	return chopList(lx)
}

func DebugParseList(ls *list) (*Node, error) {
	return parseList(ls)
}
