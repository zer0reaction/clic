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
		Exp  *Node
		Body *Node
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
		return n.BinOp.Rval.GetType()
	case NodeFunCall:
		panic("no support for returning values from functions yet")
	default:
		// TODO: seems like a hack
		return types.None
	}
}
