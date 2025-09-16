/*
   Parsing is done in two steps:

   1. Construct a list of items
   2. Convert a list to a tree
*/

package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
)

type list struct {
	head *item
	tail *item
}

type item struct {
	tp   itemType
	id   int
	lp   *list
	next *item
}

type itemType uint

const (
	itemError itemType = iota
	itemPlus
	itemInteger
	itemList
)

func (ls *list) add(it *item) {
	if ls.tail == nil && ls.head != ls.tail {
		panic("list head is not nil")
	}

	if ls.tail == nil {
		ls.head = it
		ls.tail = it
	} else {
		ls.tail.next = it
		ls.tail = ls.tail.next
	}
}

func chopList(lx *lexer.Lexer) (*list, error) {
	var ls list

	_, err := lx.Match(lexer.TokenRbrOpen)
	if err != nil {
		return nil, err
	}

	err = chopListBody(lx, &ls)
	if err != nil {
		return nil, err
	}

	_, err = lx.Match(lexer.TokenRbrClose)
	if err != nil {
		return nil, err
	}

	return &ls, nil
}

func chopListBody(lx *lexer.Lexer, ls *list) error {
	lookahead, err := lx.PeekToken(0)
	if err != nil {
		return err
	}

	switch lookahead.Type {
	case lexer.TokenRbrClose:
		// chopList() matches ')'
		return nil
	case lexer.TokenPlus:
		it := item{id: symbol.IdNone}

		_, err := lx.Match(lexer.TokenPlus)
		if err != nil {
			return err
		}
		it.tp = itemPlus

		ls.add(&it)
		return chopListBody(lx, ls)
	case lexer.TokenInteger:
		it := item{id: symbol.IdNone}

		_, err := lx.Match(lexer.TokenInteger)
		if err != nil {
			return err
		}
		it.tp = itemPlus

		ls.add(&it)
		return chopListBody(lx, ls)
	case lexer.TokenRbrOpen:
		it := item{id: symbol.IdNone}

		lp, err := chopList(lx)
		if err != nil {
			return err
		}
		it.lp = lp

		ls.add(&it)
		return chopListBody(lx, ls)
	default:
		return fmt.Errorf(":%d:%d: expected list body",
			lookahead.Line, lookahead.Column)
	}
}

/*
type NodeType uint

const (
	NodeBinOpSum NodeType = iota
	NodeInteger
)

type Node struct {
	Type    NodeType
	TableId int
	Next    *Node

	BinOpLval *Node
	BinOpRval *Node
}

func Parse(l *lexer.Lexer) (*Node, error) {
	var head *Node
	var tail *Node

	for {
		t, err := l.PeekToken(0)
		if err != nil {
			return nil, err
		}

		if t.Type == lexer.TokenEOF {
			break
		}

		n, err := list(l)
		if err != nil {
			return nil, err
		}

		if tail == nil {
			if head != nil {
				panic("head is not nil")
			}
			head = n
			tail = n
		} else {
			tail.Next = n
			tail = n
		}
	}

	return head, nil
}

func list(l *lexer.Lexer) (*Node, error) {
	lookahead, err := l.PeekToken(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Type {
	case lexer.TokenRbrOpen:
		_, err := l.Match(lexer.TokenRbrOpen)
		if err != nil {
			return nil, err
		}

		n, err := expr(l)
		if err != nil {
			return nil, err
		}

		_, err = l.Match(lexer.TokenRbrClose)
		if err != nil {
			return nil, err
		}

		return n, nil
	default:
		return nil, fmt.Errorf(":%d:%d: expected list",
			lookahead.Line, lookahead.Column)
	}
}

func expr(l *lexer.Lexer) (*Node, error) {
	lookahead, err := l.PeekToken(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Type {
	case lexer.TokenPlus:
		_, err := l.Match(lexer.TokenPlus)
		if err != nil {
			return nil, err
		}

		n := Node{
			Type:    NodeBinOpSum,
			TableId: symbol.IdNone,
		}

		n.BinOpLval, err = expr(l)
		if err != nil {
			return nil, err
		}
		n.BinOpRval, err = expr(l)
		if err != nil {
			return nil, err
		}

		return &n, nil
	case lexer.TokenInteger:
		t, err := l.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		symbol.DataToIntegerValue(t.TableId)

		n := Node{
			Type:    NodeInteger,
			TableId: t.TableId,
		}

		return &n, nil
	case lexer.TokenRbrOpen:
		return list(l)
	default:
		return nil, fmt.Errorf(":%d:%d: expected expression",
			lookahead.Line, lookahead.Column)
	}
}
*/
