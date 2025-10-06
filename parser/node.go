// This file contains node type defenition and method implementation.

package parser

import (
	sym "github.com/zer0reaction/lisp-go/symbol"
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
		Type  sym.ValueType
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

func (n *Node) GetType() sym.ValueType {
	switch n.Tag {
	case NodeInteger:
		return n.Integer.Type
	case NodeVariable:
		v := sym.GetVariable(n.Id)
		return v.Type
	case NodeBoolean:
		return sym.ValueBoolean
	case NodeBinOp:
		return n.BinOp.Rval.GetType()
	case NodeIf:
		return sym.ValueNone
	default:
		panic("node does not have a type")
	}
}
