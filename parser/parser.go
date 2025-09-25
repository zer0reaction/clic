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
	NodeBinOp
	NodeInteger
	NodeBlock
	NodeVariableDecl
	NodeVariable
	NodeFunEx
	NodeFunCall
)

type BinOpTag uint

const (
	binOpError BinOpTag = iota
	BinOpSum
	BinOpSub
	BinOpAssign
)

type Node struct {
	Tag  NodeTag
	Next *Node

	Integer struct {
		Value int64
	}
	BinOp struct {
		Tag  BinOpTag
		Lval *Node
		Rval *Node
	}
	Block struct {
		Id    symbol.BlockId
		Start *Node
	}
	Variable struct {
		Id symbol.SymbolId
	}
	Function struct {
		Id       symbol.SymbolId
		ArgStart *Node
	}
}

var blockStack = []symbol.BlockId{0}

func pushBlock(id symbol.BlockId) {
	for i := 0; i < len(blockStack); i++ {
		if blockStack[i] == id {
			panic("id is already on stack")
		}
	}
	blockStack = append(blockStack, id)
}

func popBlock() {
	if len(blockStack) == 1 {
		panic("trying to pop global block id")
	}
	blockStack = blockStack[:len(blockStack)-1]
}

func resolveVar(name string) (symbol.SymbolId, error) {
	for i := len(blockStack) - 1; i >= 0; i-- {
		id, err := symbol.LookupVariable(name, blockStack[i])
		if err == nil {
			return id, nil
		}
	}
	return symbol.SymbolIdNone, errors.New("internal: variable not visible")
}

func CreateAST(lx *lexer.Lexer) (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		t, err := lx.Peek(0)
		if err != nil {
			return nil, err
		}
		if t.Tag == lexer.TokenEOF {
			break
		}

		n, err := parseList(lx, symbol.BlockId(0))
		if err != nil {
			return nil, err
		}

		if tail == nil {
			head = n
			tail = n
		} else {
			tail.Next = n
			tail = tail.Next
		}
	}

	return head, nil
}

func parseList(lx *lexer.Lexer, blockId symbol.BlockId) (*Node, error) {
	err := lx.Match(lexer.TokenTag('('))
	if err != nil {
		return nil, err
	}

	lookahead, err := lx.Peek(0)
	if err != nil {
		return nil, err
	}

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenTag('+'):
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpSum

		err := lx.Match(lexer.TokenTag('+'))
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx, blockId)
		if err != nil {
			return nil, err
		}
	case lexer.TokenTag('-'):
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpSub

		err := lx.Match(lexer.TokenTag('-'))
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx, blockId)
		if err != nil {
			return nil, err
		}
	case lexer.TokenColEq:
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpAssign

		err := lx.Match(lexer.TokenColEq)
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx, blockId)
		if err != nil {
			return nil, err
		}
	case lexer.TokenTag('('):
		n.Tag = NodeBlock
		n.Block.Id = blockId + 1

		pushBlock(n.Block.Id)

		items, err := collectItems(lx, n.Block.Id)
		if err != nil {
			return nil, err
		}
		n.Block.Start = items

		popBlock()
	case lexer.TokenLet:
		n.Tag = NodeVariableDecl

		err := lx.Match(lexer.TokenLet)
		if err != nil {
			return nil, err
		}

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := t.Data

		_, err = symbol.LookupVariable(name, blockId)
		if err == nil {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block",
				t.Line, t.Column)
		}

		id := symbol.AddSymbol(symbol.SymbolVariable)
		symbol.SetVariable(id, symbol.Variable{
			Name:    name,
			BlockId: blockId,
		})
		n.Variable.Id = id
	case lexer.TokenExfun:
		n.Tag = NodeFunEx

		err := lx.Match(lexer.TokenExfun)
		if err != nil {
			return nil, err
		}

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := t.Data

		_, err = symbol.LookupFunction(name)
		if err == nil {
			return nil, fmt.Errorf(":%d:%d: error: function is already declared",
				t.Line, t.Column)
		}

		id := symbol.AddSymbol(symbol.SymbolFunction)
		symbol.SetFunction(id, symbol.Function{
			Name: name,
		})
		n.Function.Id = id
	case lexer.TokenIdent:
		n.Tag = NodeFunCall

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := t.Data

		id, err := symbol.LookupFunction(name)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: function is not declared",
				t.Line, t.Column)
		}
		n.Function.Id = id

		items, err := collectItems(lx, blockId)
		if err != nil {
			return nil, err
		}
		n.Function.ArgStart = items
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list head item",
			lookahead.Line, lookahead.Column)
	}

	err = lx.Match(lexer.TokenTag(')'))
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func parseBinOp(n *Node, lx *lexer.Lexer, blockId symbol.BlockId) error {
	lval, err := parseItem(lx, blockId)
	if err != nil {
		return err
	}
	rval, err := parseItem(lx, blockId)
	if err != nil {
		return err
	}
	n.BinOp.Lval = lval
	n.BinOp.Rval = rval

	return nil
}

func collectItems(lx *lexer.Lexer, blockId symbol.BlockId) (*Node, error) {
	lookahead, err := lx.Peek(0)
	if err != nil {
		return nil, err
	}

	var head *Node = nil
	var tail *Node = nil

	for lookahead.Tag != lexer.TokenTag(')') {
		item, err := parseItem(lx, blockId)
		if err != nil {
			return nil, err
		}

		if tail == nil {
			head = item
			tail = item
		} else {
			tail.Next = item
			tail = tail.Next
		}

		lookahead, err = lx.Peek(0)
		if err != nil {
			return nil, err
		}
	}

	return head, nil
}

func parseItem(lx *lexer.Lexer, blockId symbol.BlockId) (*Node, error) {
	lookahead, err := lx.Peek(0)
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

		id, err := resolveVar(name)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope",
				lookahead.Line, lookahead.Column)
		}
		n.Variable.Id = id

		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
	case lexer.TokenTag('('):
		return parseList(lx, blockId)
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item",
			lookahead.Line, lookahead.Column)
	}

	return &n, nil
}
