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
	NodeBinOpAssign
	NodeInteger
	NodeBlock
	NodeVariableDecl
	NodeVariable
	NodeFunEx
	NodeFunCall
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
		Id    symbol.BlockId
		Start *Node
	}
	Variable struct {
		TableId symbol.SymbolId
	}
	Function struct {
		TableId  symbol.SymbolId
		ArgStart *Node
	}
}

var blkIdStack = []symbol.BlockId{0}

func pushBlkId(id symbol.BlockId) {
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

func resolveVar(name string) (symbol.SymbolId, error) {
	for i := len(blkIdStack) - 1; i >= 0; i-- {
		tableId, err := symbol.LookupVariable(name, blkIdStack[i])
		if err == nil {
			return tableId, nil
		}
	}
	return symbol.SymbolIdNone, errors.New("internal: variable not visible")
}

func CreateAST(lx *lexer.Lexer) (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		t, err := lx.PeekToken(0)
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

func parseList(lx *lexer.Lexer, curBlkId symbol.BlockId) (*Node, error) {
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

		err = parseBinOp(&n, lx, curBlkId)
		if err != nil {
			return nil, err
		}
	case lexer.TokenSet:
		n.Tag = NodeBinOpAssign

		err := lx.Match(lexer.TokenSet)
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx, curBlkId)
		if err != nil {
			return nil, err
		}
	case lexer.TokenRbrOpen:
		n.Tag = NodeBlock
		n.Block.Id = curBlkId + 1

		pushBlkId(n.Block.Id)

		var tail *Node = nil

		// TODO: refactor in a separate function
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
		n.Tag = NodeVariableDecl

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
		name := t.Data

		_, err = symbol.LookupVariable(name, curBlkId)
		if err == nil {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block",
				t.Line, t.Column)
		}

		tableId := symbol.AddSymbol(symbol.SymbolVariable)
		symbol.SetVariable(tableId, symbol.Variable{
			Name:    name,
			BlockId: curBlkId,
		})
		n.Variable.TableId = tableId
	case lexer.TokenExfun:
		n.Tag = NodeFunEx

		err := lx.Match(lexer.TokenExfun)
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

		name := t.Data

		_, err = symbol.LookupFunction(name)
		if err == nil {
			return nil, fmt.Errorf(":%d:%d: error: function is already declared",
				t.Line, t.Column)
		}

		tableId := symbol.AddSymbol(symbol.SymbolFunction)
		symbol.SetFunction(tableId, symbol.Function{
			Name: name,
		})
		n.Function.TableId = tableId
	case lexer.TokenIdent:
		n.Tag = NodeFunCall

		t, err := lx.PeekToken(0)
		if err != nil {
			return nil, err
		}
		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}

		name := t.Data

		tableId, err := symbol.LookupFunction(name)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: function is not declared",
				t.Line, t.Column)
		}

		n.Function.TableId = tableId

		var tail *Node = nil

		// TODO: refactor in a separate function
		for lookahead.Tag != lexer.TokenRbrClose {
			argNode, err := parseItem(lx, n.Block.Id)
			if err != nil {
				return nil, err
			}

			if tail == nil && n.Function.ArgStart != nil {
				panic("arg start is not nil")
			}

			if tail == nil {
				n.Function.ArgStart = argNode
				tail = argNode
			} else {
				tail.Next = argNode
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

func parseBinOp(n *Node, lx *lexer.Lexer, curBlkId symbol.BlockId) error {
	// TODO: add checks
	lval, err := parseItem(lx, curBlkId)
	if err != nil {
		return err
	}
	rval, err := parseItem(lx, curBlkId)
	if err != nil {
		return err
	}
	n.BinOp.Lval = lval
	n.BinOp.Rval = rval

	return nil
}

func parseItem(lx *lexer.Lexer, curBlkId symbol.BlockId) (*Node, error) {
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

		tableId, err := resolveVar(name)
		if err != nil {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope",
				lookahead.Line, lookahead.Column)
		}
		n.Variable.TableId = tableId

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
