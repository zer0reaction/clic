// This file contains the type checking functions.

// Type checking probably shouldn't be done in the parser package, but
// it is convenient to place it here.

package parser

import (
	"lisp-go/report"
	"lisp-go/types"
)

func (p *Parser) TypeCheck(n *Node) {
	if n == nil {
		return
	}

	switch n.Tag {
	case NodeBinOp:
		p.TypeCheck(n.BinOp.Lval)
		p.TypeCheck(n.BinOp.Rval)

		lvalType := n.BinOp.Lval.GetType()
		rvalType := n.BinOp.Rval.GetType()
		if lvalType != rvalType {
			p.reportHere(n,
				report.ReportNonfatal,
				"operand type mismatch")
		}

		isAssign := (n.BinOp.Tag == BinOpAssign)
		isStorage := (n.BinOp.Lval.Tag == NodeVariable)
		if isAssign && !isStorage {
			p.reportHere(n,
				report.ReportNonfatal,
				"lvalue is not a storage location")
		}
		p.TypeCheck(n.Block.Start)

	case NodeFunCall:
		p.TypeCheck(n.Function.ArgStart)

	case NodeIf:
		p.TypeCheck(n.If.Exp)

		expType := n.If.Exp.GetType()
		if expType != types.Bool {
			p.reportHere(n,
				report.ReportNonfatal,
				"expected boolean type in expression")
		}

		p.TypeCheck(n.If.IfBody)
		p.TypeCheck(n.If.ElseBody)

	// do nothing
	case NodeInteger:
	case NodeBoolean:
	case NodeBlock:
	case NodeVariableDecl:
	case NodeVariable:
	case NodeFunEx:

	default:
		panic("node not implemented")
	}

	p.TypeCheck(n.Next)
}
