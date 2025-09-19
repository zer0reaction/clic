package parser

import (
	"github.com/zer0reaction/lisp-go/lexer"
)

func DebugParseList(lx *lexer.Lexer, curBlkId uint) (*Node, error) {
	return parseList(lx, curBlkId)
}
