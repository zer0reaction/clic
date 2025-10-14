// This file contains the type checking functions.

package checker

import (
	"fmt"
	"lisp-go/ast"
	"lisp-go/report"
	sym "lisp-go/symbol"
	"lisp-go/types"
)

func TypeCheck(roots []*ast.Node, r *report.Reporter) {
	for _, node := range roots {
		checkNode(node, r)
	}
}

func checkNode(n *ast.Node, r *report.Reporter) {
	if n == nil {
		return
	}

	switch n.Tag {
	case ast.NodeBinOp:
		checkNode(n.BinOp.Lval, r)
		checkNode(n.BinOp.Rval, r)

		lvalType := n.BinOp.Lval.GetType()
		rvalType := n.BinOp.Rval.GetType()
		if lvalType != rvalType {
			n.ReportHere(r, report.ReportNonfatal,
				"operand type mismatch")
		}

		isAssign := (n.BinOp.Tag == ast.BinOpAssign)
		isStorage := (n.BinOp.Lval.Tag == ast.NodeVariable)
		if isAssign && !isStorage {
			n.ReportHere(r, report.ReportNonfatal,
				"lvalue is not a storage location")
		}

	case ast.NodeFunCall:
		for _, node := range n.FunCall.Args {
			checkNode(node, r)
		}

		fun := sym.GetFunction(n.Id)

		if len(n.FunCall.Args) != len(fun.Params) {
			var where *ast.Node

			if len(n.FunCall.Args) > 0 {
				where = n.FunCall.Args[0]
			} else {
				where = n
			}

			where.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected %d arguments, got %d", len(fun.Params), len(n.FunCall.Args)))
		}

		for i, arg := range n.FunCall.Args {
			if arg.GetType() != fun.Params[i].Type {
				n.ReportHere(r, report.ReportNonfatal,
					"mismatched types in function call")
			}
		}

	case ast.NodeIf:
		checkNode(n.If.Exp, r)

		expType := n.If.Exp.GetType()
		if expType != types.Bool {
			n.If.Exp.ReportHere(r, report.ReportNonfatal,
				"expected boolean type")
		}

		checkNode(n.If.IfBody, r)
		checkNode(n.If.ElseBody, r)

	case ast.NodeWhile:
		checkNode(n.While.Exp, r)

		expType := n.While.Exp.GetType()
		if expType != types.Bool {
			n.While.Exp.ReportHere(r, report.ReportNonfatal,
				"expected boolean type")
		}

		checkNode(n.While.Body, r)

	case ast.NodeBlock:
		for _, node := range n.Block.Stmts {
			checkNode(node, r)
		}

	case ast.NodeCast:
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
			n.Cast.What.ReportHere(r, report.ReportNonfatal,
				"can not cast from type None")

		default:
			panic("not implemented")
		}

	// Do nothing
	case ast.NodeInteger:
	case ast.NodeBoolean:
	case ast.NodeVariableDecl:
	case ast.NodeVariable:
	case ast.NodeFunEx:

	default:
		panic("node not implemented")
	}
}
