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
	Tag    NodeTag
	Next   *Node
	Id     sym.SymbolId
	Line   uint
	Column uint

	Integer struct {
		Value int64
		Type  sym.ValueType
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
}

func (n *Node) GetType() sym.ValueType {
	switch n.Tag {
	case NodeInteger:
		return n.Integer.Type
	case NodeVariable:
		v := sym.GetVariable(n.Id)
		return v.Type
	case NodeBinOp:
		return n.BinOp.Rval.GetType()
	default:
		panic("node does not have a type")
	}
}
