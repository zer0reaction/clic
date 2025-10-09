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
	NodeWhile
)

type BinOpTag uint

const (
	binOpError BinOpTag = iota
	BinOpAssign
	BinOpArith
	BinOpComp
)

type BinOpArithTag uint

const (
	binOpArithError BinOpArithTag = iota
	BinOpSum
	BinOpSub
)

type BinOpCompTag uint

const (
	binOpCompError BinOpCompTag = iota
	BinOpEq
	BinOpNeq
	BinOpLessEq
	BinOpLess
	BinOpGreatEq
	BinOpGreat
)

type Node struct {
	Tag    NodeTag
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
		Tag      BinOpTag
		ArithTag BinOpArithTag
		CompTag  BinOpCompTag
		Lval     *Node
		Rval     *Node
	}
	Block struct {
		Stmts []*Node
	}
	Function struct {
		Args []*Node
	}
	If struct {
		Exp      *Node
		IfBody   *Node
		ElseBody *Node
	}
	While struct {
		Exp  *Node
		Body *Node
	}
}

func (n *Node) GetType() types.Type {
	switch n.Tag {
	case NodeBinOp:
		switch n.BinOp.Tag {
		case BinOpAssign:
			return n.BinOp.Lval.GetType()
		case BinOpArith:
			return n.BinOp.Rval.GetType()
		case BinOpComp:
			return types.Bool
		default:
			panic("invalid binop tag")
		}
	case NodeInteger:
		return n.Integer.Type
	case NodeBoolean:
		return types.Bool
	case NodeBlock:
		return types.None
	case NodeVariableDecl:
		return types.None
	case NodeVariable:
		v := sym.GetVariable(n.Id)
		return v.Type
	case NodeFunEx:
		return types.None
	case NodeFunCall:
		// TODO: add checking types
		return types.None
	case NodeIf:
		return types.None
	default:
		panic("not implemented")
	}
}
