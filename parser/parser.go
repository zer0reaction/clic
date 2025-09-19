package parser

import (
	"errors"
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
		Id    uint
		Start *Node
	}
	Variable struct {
		TableId uint
	}
}

var blkIdStack = []uint{0}

func pushBlkId(id uint) {
	for i := 0; i < len(blkIdStack); i++ {
		if blkIdStack[i] == id {
			panic("id is already on stack")
		}
	}
	blkIdStack = append(blkIdStack, id)
}

func popBlkId() {
	if len(blkIdStack) == 1 {
		panic("trying to pop global block id")
	}
	blkIdStack = blkIdStack[:len(blkIdStack)-1]
}

func resolveVarBlkId(name string) (uint, error) {
	for i := len(blkIdStack) - 1; i >= 0; i-- {
		if symbol.IsVariableInBlock(name, blkIdStack[i]) {
			return blkIdStack[i], nil
		}
	}
	return 0, errors.New("internal: variable not visible")
}

func parseList(lx *lexer.Lexer, curBlkId uint) (*Node, error) {
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
		lval, err := parseItem(lx, curBlkId)
		if err != nil {
			return nil, err
		}
		rval, err := parseItem(lx, curBlkId)
		if err != nil {
			return nil, err
		}
		n.BinOp.Lval = lval
		n.BinOp.Rval = rval
	case lexer.TokenRbrOpen:
		n.Tag = NodeBlock
		n.Block.Id = curBlkId + 1

		pushBlkId(n.Block.Id)

		var tail *Node = nil

		for lookahead.Tag != lexer.TokenRbrClose {
			blockNode, err := parseList(lx, n.Block.Id)
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

		popBlkId()
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
		name := t.Data

		if symbol.IsVariableInBlock(name, curBlkId) {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block",
				t.Line, t.Column)
		}
		tableId := symbol.AddVariable(name, curBlkId)
		n.Variable.TableId = tableId

		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
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

func parseItem(lx *lexer.Lexer, curBlkId uint) (*Node, error) {
	lookahead, err := lx.PeekToken(0)
	if err != nil {
		return nil, err
	}

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenInteger:
		n.Tag = NodeInteger

		value, err := strconv.ParseInt(lookahead.Data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value

		err = lx.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}
	case lexer.TokenIdent:
		n.Tag = NodeVariable
		name := lookahead.Data

		blockId, err := resolveVarBlkId(name)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope",
				lookahead.Line, lookahead.Column)
		}
		n.Variable.TableId = symbol.LookupVariable(name, blockId)

		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
	case lexer.TokenRbrOpen:
		return parseList(lx, curBlkId)
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item",
			lookahead.Line, lookahead.Column)
	}

	return &n, nil
}
