package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
	"strconv"
)

type NodeTag uint

const (
	nodeError NodeTag = iota
	NodeBinOpSum
	NodeInteger
	NodeBlock
	NodeVariable
)

type Node struct {
	Tag  NodeTag
	Next *Node

	Integer struct {
		Value int64
	}
	BinOp struct {
		Lval *Node
		Rval *Node
	}
	Block struct {
		Start *Node
	}
	Variable struct {
		TableId uint
	}
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

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenPlus:
		n.Tag = NodeBinOpSum

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
		n.BinOp.Lval = lval
		n.BinOp.Rval = rval
	case lexer.TokenRbrOpen:
		n.Tag = NodeBlock

		var tail *Node = nil

		for lookahead.Tag != lexer.TokenRbrClose {
			blockNode, err := parseList(lx)
			if err != nil {
				return nil, err
			}

			if tail == nil && n.Block.Start != nil {
				panic("block start is not nil")
			}

			if tail == nil {
				n.Block.Start = blockNode
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
	case lexer.TokenLet:
		n.Tag = NodeVariable

		err := lx.Match(lexer.TokenLet)
		if err != nil {
			return nil, err
		}

		t, err := lx.PeekToken(0)
		if err != nil {
			return nil, err
		}
		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}

		// TODO: add block ids
		tableId, err := symbol.AddVariable(t.Data, 0)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block",
				t.Line, t.Column)
		}
		n.Variable.TableId = tableId
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

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenInteger:
		n.Tag = NodeInteger

		err := lx.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		value, err := strconv.ParseInt(lookahead.Data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}

		n.Integer.Value = value
	case lexer.TokenRbrOpen:
		return parseList(lx)
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item",
			lookahead.Line, lookahead.Column)
	}

	return &n, nil
}
