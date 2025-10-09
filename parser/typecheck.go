// This file contains the type checking functions.

// Type checking probably shouldn't be done in the parser package, but
// it is convenient to place it here.

package parser

import (
	"lisp-go/report"
	"lisp-go/types"
)

func (p *Parser) TypeCheck(roots []*Node) {
	for _, node := range roots {
		p.checkNode(node)
	}
}

func (p *Parser) checkNode(n *Node) {
	if n == nil {
		return
	}

	switch n.Tag {
	case NodeBinOp:
		p.checkNode(n.BinOp.Lval)
		p.checkNode(n.BinOp.Rval)

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

	case NodeFunCall:
		for _, node := range n.Function.Args {
			p.checkNode(node)
		}

	case NodeIf:
		p.checkNode(n.If.Exp)

		expType := n.If.Exp.GetType()
		if expType != types.Bool {
			p.reportHere(n.If.Exp,
				report.ReportNonfatal,
				"expected boolean type")
		}

		p.checkNode(n.If.IfBody)
		p.checkNode(n.If.ElseBody)

	case NodeWhile:
		p.checkNode(n.While.Exp)

		expType := n.While.Exp.GetType()
		if expType != types.Bool {
			p.reportHere(n.While.Exp,
				report.ReportNonfatal,
				"expected boolean type")
		}

		p.checkNode(n.While.Body)

	case NodeBlock:
		for _, node := range n.Block.Stmts {
			p.checkNode(node)
		}

	// do nothing
	case NodeInteger:
	case NodeBoolean:
	case NodeVariableDecl:
	case NodeVariable:
	case NodeFunEx:

	default:
		panic("node not implemented")
	}
}
