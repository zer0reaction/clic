package parser

import (
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
)

func DebugParseList(lx *lexer.Lexer, blockId symbol.BlockId) (*Node, error) {
	return parseList(lx, blockId)
}
