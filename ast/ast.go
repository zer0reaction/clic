// This file contains AST node defenition.

package ast

import (
	"clic/report"
	"clic/symbol"
	"clic/types"
)

type Node struct {
	Tag    tag
	Id     symbol.Id
	Line   uint
	Column uint

	// Union could really help here... Sigh.

	Int struct {
		SValue int64
		UValue uint64
		Signed bool
		Size   uint8
	}

	Bool struct {
		Value bool
	}

	BinOp struct {
		Tag      BinOpTag
		ArithTag BinOpArithTag
		CompTag  BinOpCompTag
		Lval     *Node
		Rval     *Node
	}

	Scope struct {
		Stmts []*Node
	}

	Fun struct {
		// Function definiton
		Params []symbol.Id
		Stmts  []*Node

		// Function call
		Args []*Node
	}

	// TODO: This is a dirty hack.
	Return struct {
		Fun symbol.Id
		Val *Node
	}

	If struct {
		Exp       *Node
		IfStmts   []*Node
		ElseStmts []*Node
	}

	While struct {
		Exp   *Node
		Stmts []*Node
	}

	For struct {
		Init  *Node
		Cond  *Node
		Adv   *Node
		Stmts []*Node
	}

	Cast struct {
		To   types.Id
		What *Node
	}
}

type tag uint

const (
	nodeError tag = iota
	NodeBinOp
	NodeInt
	NodeBool
	NodeScope
	NodeLVarDecl
	NodeLVar
	NodeFunEx
	NodeFunDef
	NodeFunCall
	NodeReturn
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
func (n *Node) GetTypeShallow(t *symbol.Table) types.Id {
	switch n.Tag {
	case NodeBinOp:
		switch n.BinOp.Tag {
		case BinOpAssign:
			return n.BinOp.Lval.GetTypeShallow(t)

		case BinOpArith:
			return n.BinOp.Rval.GetTypeShallow(t)

		case BinOpComp:
			return types.GetBuiltin(types.Bool)

		default:
			panic("not implemented")
		}

	case NodeInt:
		switch n.Int.Size {
		case 64:
			if n.Int.Signed {
				return types.GetBuiltin(types.S64)
			} else {
				return types.GetBuiltin(types.U64)
			}
		default:
			panic("not implemented")
		}

	case NodeBool:
		return types.GetBuiltin(types.Bool)

	case NodeScope:
		return types.GetBuiltin(types.Void)

	case NodeLVar:
		return t.Get(n.Id).Type

	case NodeLVarDecl:
		return t.Get(n.Id).Type

	case NodeFunEx:
		return types.GetBuiltin(types.Void)

	case NodeFunCall:
		return t.Get(n.Id).Type

	case NodeReturn:
		return n.Return.Val.GetTypeShallow(t)

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
func (n *Node) GetTypeDeep(t *symbol.Table) types.Id {
	typeId := n.GetTypeShallow(t)
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
