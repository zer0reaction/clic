// This file contains node type defenition and method implementation.

package parser

import (
	sym "lisp-go/symbol"
	"lisp-go/types"
)

type NodeTag uint

const (
	nodeError NodeTag = iota
	NodeBinOp
	NodeInteger
	NodeBoolean
	NodeBlock
	NodeVariableDecl
	NodeVariable
	NodeFunEx
	NodeFunCall
	NodeIf
)

type BinOpTag uint

const (
	binOpError BinOpTag = iota
	BinOpSum
	BinOpSub
	BinOpAssign
	BinOpEq
	BinOpLessEq
	BinOpLess
	BinOpGreatEq
	BinOpGreat
)

type Node struct {
	Tag    NodeTag
	Next   *Node
	Id     sym.SymbolId
	Line   uint
	Column uint

	Integer struct {
		Value int64
		Type  types.Type
	}
	Boolean struct {
		Value bool
	}
	BinOp struct {
		Tag  BinOpTag
		Lval *Node
		Rval *Node
	}
	Block struct {
		Start *Node
	}
	Function struct {
		ArgStart *Node
	}
	If struct {
		Exp      *Node
		IfBody   *Node
		ElseBody *Node
	}
}

func (n *Node) GetType() types.Type {
	switch n.Tag {
	case NodeInteger:
		return n.Integer.Type
	case NodeVariable:
		v := sym.GetVariable(n.Id)
		return v.Type
	case NodeBoolean:
		return types.Bool
	case NodeBinOp:
		switch n.BinOp.Tag {
		case BinOpAssign:
			return n.BinOp.Rval.GetType()
		case BinOpSum:
			return n.BinOp.Rval.GetType()
		case BinOpSub:
			return n.BinOp.Rval.GetType()
		case BinOpEq:
			return types.Bool
		case BinOpLessEq:
			return types.Bool
		case BinOpLess:
			return types.Bool
		case BinOpGreatEq:
			return types.Bool
		case BinOpGreat:
			return types.Bool
		default:
			panic("invalid binop tag")
		}
	case NodeFunCall:
		panic("no support for returning values from functions yet")
	default:
		return types.None
	}
}
