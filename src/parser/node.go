// This file contains node type defenition and method implementation.

package parser

import (
	sym "lisp-go/src/symbol"
	"lisp-go/src/types"
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
	NodeCast
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
	BinOpMult
	BinOpDiv
	BinOpMod
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

	FunCall struct {
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

	Cast struct {
		To   types.Type
		What *Node
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

	case NodeCast:
		return n.Cast.To

	default:
		panic("not implemented")
	}
}
