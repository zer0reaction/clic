// This file contains AST node defenition.

package ast

import (
	"lisp-go/report"
	sym "lisp-go/symbol"
	"lisp-go/types"
)

type Node struct {
	Tag    NodeTag
	Id     sym.SymbolId
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

func (n *Node) GetType() types.Type {
	switch n.Tag {
	case NodeBinOp:
		switch n.BinOp.Tag {
		case BinOpAssign:
			return n.BinOp.Lval.GetType()

		case BinOpArith:
			return n.BinOp.Rval.GetType()

		case BinOpComp:
			return types.GetBuiltin(types.Bool)

		default:
			panic("invalid binop tag")
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
			panic("invalid integer size")
		}

	case NodeBoolean:
		return types.GetBuiltin(types.Bool)

	case NodeBlock:
		return types.GetBuiltin(types.None)

	case NodeVariableDecl:
		return types.GetBuiltin(types.None)

	case NodeVariable:
		v := sym.GetVariable(n.Id)
		return v.Type

	case NodeFunEx:
		return types.GetBuiltin(types.None)

	case NodeFunCall:
		// TODO: Get return value type
		return types.GetBuiltin(types.None)

	case NodeIf:
		return types.GetBuiltin(types.None)

	case NodeCast:
		return n.Cast.To

	default:
		panic("not implemented")
	}
}

func (n *Node) ReportHere(r *report.Reporter, tag report.ReportTag, msg string) {
	r.Report(report.Form{
		Tag:    tag,
		Line:   n.Line,
		Column: n.Column,
		Msg:    msg,
	})
}
