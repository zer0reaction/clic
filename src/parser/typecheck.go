// This file contains the type checking functions.

// Type checking probably shouldn't be done in the parser package, but
// it is convenient to place it here.

package parser

import (
	"fmt"
	"lisp-go/src/report"
	sym "lisp-go/src/symbol"
	"lisp-go/src/types"
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
		for _, node := range n.FunCall.Args {
			p.checkNode(node)
		}

		fun := sym.GetFunction(n.Id)

		if len(n.FunCall.Args) != len(fun.Params) {
			var where *Node

			if len(n.FunCall.Args) > 0 {
				where = n.FunCall.Args[0]
			} else {
				where = n
			}

			p.reportHere(where,
				report.ReportNonfatal,
				fmt.Sprintf("expected %d arguments, got %d",
					len(fun.Params), len(n.FunCall.Args)))
		}

		for i, arg := range n.FunCall.Args {
			if arg.GetType() != fun.Params[i].Type {
				p.reportHere(arg,
					report.ReportNonfatal,
					"mismatched types in function call")
			}
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

	case NodeCast:
		// Right now there are only integer types, so we can
		// convert between all of them. This code only checks
		// for new and unsupported types.

		from := n.Cast.What.GetType()

		switch from {
		// Do nothing
		case types.S64:
		case types.U64:
		case types.Bool:

		case types.None:
			p.reportHere(n.Cast.What,
				report.ReportNonfatal,
				"can not cast from type None")

		default:
			panic("not implemented")
		}

	// Do nothing
	case NodeInteger:
	case NodeBoolean:
	case NodeVariableDecl:
	case NodeVariable:
	case NodeFunEx:

	default:
		panic("node not implemented")
	}
}
