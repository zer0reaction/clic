package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
)

type NodeType uint

const (
	nodeError NodeType = iota
	NodeBinOpSum
	NodeInteger
	NodeBlock
)

type Node struct {
	Type    NodeType
	TableId int
	Next    *Node

	BinOpLval *Node
	BinOpRval *Node

	BlockStart *Node
}

func parseList(lx *lexer.Lexer) (*Node, error) {
	err := lx.Match(lexer.TokenRbrOpen)
	if err != nil {
		return nil, err
	}

	lookahead, err := lx.PeekToken(0)
	if err != nil {
		return nil, err
	}

	n := Node{
		TableId: symbol.IdNone,
	}

	switch lookahead.Type {
	case lexer.TokenPlus:
		n.Type = NodeBinOpSum

		err := lx.Match(lexer.TokenPlus)
		if err != nil {
			return nil, err
		}

		// TODO: add checks
		lval, err := parseItem(lx)
		if err != nil {
			return nil, err
		}
		rval, err := parseItem(lx)
		if err != nil {
			return nil, err
		}
		n.BinOpLval = lval
		n.BinOpRval = rval
	case lexer.TokenRbrOpen:
		n.Type = NodeBlock

		var tail *Node

		for lookahead.Type != lexer.TokenRbrClose {
			blockNode, err := parseList(lx)
			if err != nil {
				return nil, err
			}

			if tail == nil && n.BlockStart != nil {
				panic("block start is not nil")
			}

			if tail == nil {
				n.BlockStart = blockNode
				tail = blockNode
			} else {
				tail.Next = blockNode
				tail = tail.Next
			}

			lookahead, err = lx.PeekToken(0)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list head item",
			lookahead.Line, lookahead.Column)
	}

	err = lx.Match(lexer.TokenRbrClose)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func parseItem(lx *lexer.Lexer) (*Node, error) {
	lookahead, err := lx.PeekToken(0)
	if err != nil {
		return nil, err
	}

	n := Node{
		TableId: symbol.IdNone,
	}

	switch lookahead.Type {
	case lexer.TokenInteger:
		err := lx.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		n.Type = NodeInteger
		n.TableId = lookahead.TableId
		symbol.DataToIntegerValue(n.TableId)
	case lexer.TokenRbrOpen:
		return parseList(lx)
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item",
			lookahead.Line, lookahead.Column)
	}

	return &n, nil
}
