package parser

import (
	"github.com/zer0reaction/lisp-go/lexer"
)

func (ls *list) DebugCount() uint {
	if ls.tail == nil && ls.head != ls.tail {
		panic("list head is not nil")
	}

	cnt := uint(0)
	cur := ls.head

	for cur != nil {
		cur = cur.next
		cnt += 1
	}

	return cnt
}

func DebugChopList(lx *lexer.Lexer) (*list, error) {
	return chopList(lx)
}
