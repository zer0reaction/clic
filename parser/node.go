// Node type defenition and method implementation

package parser

import (
	"github.com/zer0reaction/lisp-go/symbol"
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
		Type  symbol.ValueType
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

func (n *Node) GetType() symbol.ValueType {
	switch n.Tag {
	case NodeInteger:
		return n.Integer.Type
	case NodeVariable:
		v := symbol.GetVariable(n.Variable.Id)
		return v.Type
	case NodeBinOp:
		return n.BinOp.Rval.GetType()
	default:
		panic("node does not have a type")
	}
}
