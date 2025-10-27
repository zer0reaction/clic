// This file contains AST node defenition.

package ast

import (
	"clic/report"
	"clic/symbol"
	"clic/types"
)

type Node struct {
	Tag    NodeTag
	Id     symbol.Id
	Line   uint
	Column uint

	// Union could really help here... Sigh.

	Integer struct {
		SValue int64
		UValue uint64
		Signed bool
		Size   uint8
	}

	Boolean struct {
		Value bool
	}

	Variable struct {
		IsDecl bool
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

	For struct {
		Init *Node
		Cond *Node
		Adv  *Node
		Body *Node
	}

	Cast struct {
		To   types.TypeId
		What *Node
	}
}

type NodeTag uint

const (
	nodeError NodeTag = iota
	NodeBinOp
	NodeInteger
	NodeBoolean
	NodeBlock
	NodeVariable
	NodeFunEx
	NodeFunCall
	NodeIf
	NodeWhile
	NodeFor
	NodeCast
	NodeTypedef
	NodeEmpty
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

// Does not recurse if type is a defenition
func (n *Node) GetTypeShallow() types.TypeId {
	switch n.Tag {
	case NodeBinOp:
		switch n.BinOp.Tag {
		case BinOpAssign:
			return n.BinOp.Lval.GetTypeShallow()

		case BinOpArith:
			return n.BinOp.Rval.GetTypeShallow()

		case BinOpComp:
			return types.GetBuiltin(types.Bool)

		default:
			panic("not implemented")
		}

	case NodeInteger:
		switch n.Integer.Size {
		case 64:
			if n.Integer.Signed {
				return types.GetBuiltin(types.S64)
			} else {
				return types.GetBuiltin(types.U64)
			}
		default:
			panic("not implemented")
		}

	case NodeBoolean:
		return types.GetBuiltin(types.Bool)

	case NodeBlock:
		return types.GetBuiltin(types.Void)

	case NodeVariable:
		v := symbol.Get(n.Id).Variable
		return v.Type

	case NodeFunEx:
		return types.GetBuiltin(types.Void)

	case NodeFunCall:
		// TODO: Get return value type
		return types.GetBuiltin(types.Void)

	case NodeIf:
		return types.GetBuiltin(types.Void)

	case NodeCast:
		return n.Cast.To

	case NodeEmpty:
		return types.GetBuiltin(types.Void)

	default:
		panic("not implemented")
	}
}

// If the type is a defenition, recurses to get the actual type
func (n *Node) GetTypeDeep() types.TypeId {
	typeId := n.GetTypeShallow()
	{
		typeNode := types.Get(typeId)
		for typeNode.Tag == types.Definition {
			typeId = typeNode.DefinedAs
			typeNode = types.Get(typeId)
		}
	}
	return typeId
}

func (n *Node) ReportHere(r *report.Reporter, tag report.ReportTag, msg string) {
	r.Report(report.Form{
		Tag:    tag,
		Line:   n.Line,
		Column: n.Column,
		Msg:    msg,
	})
}
