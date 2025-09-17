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

type itemType uint

const (
	itemError itemType = iota
	itemPlus
	itemInteger
	itemList
)

type item struct {
	itemType itemType
	tableId  int
	line     uint
	column   uint
	next     *item

	listPtr *list
}

type list struct {
	head *item
	tail *item
}

type NodeType uint

const (
	nodeError NodeType = iota
	NodeBinOpSum
	NodeInteger
)

type Node struct {
	Type    NodeType
	TableId int
	Next    *Node

	BinOpLval *Node
	BinOpRval *Node
}

func (ls *list) add(it *item) {
	if ls.tail == nil && ls.head != nil {
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

func (ls *list) count() uint {
	if ls.tail == nil && ls.head != nil {
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

func (ls *list) get(index uint) *item {
	current := ls.head

	for i := uint(0); i < index; i++ {
		if current == nil {
			panic("list item unreachable")
		}
		current = current.next
	}

	return current
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

	it := item{
		tableId: symbol.IdNone,
		line:    lookahead.Line,
		column:  lookahead.Column,
	}

	switch lookahead.Type {
	case lexer.TokenRbrClose:
		// chopList() matches ')'

		return nil
	case lexer.TokenPlus:
		_, err := lx.Match(lexer.TokenPlus)
		if err != nil {
			return err
		}

		it.itemType = itemPlus

		ls.add(&it)
		return chopListBody(lx, ls)
	case lexer.TokenInteger:
		_, err := lx.Match(lexer.TokenInteger)
		if err != nil {
			return err
		}

		it.itemType = itemInteger

		ls.add(&it)
		return chopListBody(lx, ls)
	case lexer.TokenRbrOpen:
		lp, err := chopList(lx)
		if err != nil {
			return err
		}

		it.itemType = itemList
		it.listPtr = lp

		ls.add(&it)
		return chopListBody(lx, ls)
	default:
		return fmt.Errorf(":%d:%d: expected list body",
			lookahead.Line, lookahead.Column)
	}
}

func parseList(ls *list) (*Node, error) {
	if ls.head == nil {
		panic("list head is nil")
	}

	n := Node{
		TableId: symbol.IdNone,
	}

	switch ls.head.itemType {
	case itemPlus:
		n.Type = NodeBinOpSum

		if ls.count() != 3 {
			return nil, fmt.Errorf(":%d:%d: expected 3 items",
				ls.head.line, ls.head.column)
		}

		lval, err := parseItem(ls.get(1))
		if err != nil {
			return nil, err
		}
		rval, err := parseItem(ls.get(2))
		if err != nil {
			return nil, err
		}
		n.BinOpLval = lval
		n.BinOpRval = rval

		return &n, nil
	default:
		return nil, fmt.Errorf(":%d:%d: unexpected first item",
			ls.head.line, ls.head.column)
	}
}

func parseItem(it *item) (*Node, error) {
	if it == nil {
		panic("item is nil")
	}

	n := Node{
		TableId: symbol.IdNone,
	}

	switch it.itemType {
	case itemInteger:
		n.Type = NodeInteger
		n.TableId = it.tableId
		return &n, nil
	case itemList:
		return parseList(it.listPtr)
	default:
		// TODO: add displaying names
		return nil, fmt.Errorf(":%d:%d: unexpected item, got [%d]",
			it.line, it.column, it.itemType)
	}
}
