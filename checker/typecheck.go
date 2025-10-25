// This file contains the type checking functions.

package checker

import (
	"clic/ast"
	"clic/report"
	sym "clic/symbol"
	"clic/types"
	"fmt"
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

		lvalType := n.BinOp.Lval.GetTypeShallow()
		rvalType := n.BinOp.Rval.GetTypeShallow()
		voidType := types.GetBuiltin(types.Void)

		lvalStr := lvalType.Stringify()
		rvalStr := rvalType.Stringify()
		voidStr := voidType.Stringify()

		if lvalType == voidType {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("lvalue is of type %s", voidStr))
		}
		if rvalType == voidType {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("rvalue is of type %s", voidStr))
		}

		if lvalType != rvalType {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("operand type mismatch\n\tlval: %s\n\trval: %s",
					lvalStr, rvalStr))
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

		fun := sym.Get(n.Id).Function

		if len(n.FunCall.Args) != len(fun.Params) {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected %d arguments, got %d", len(fun.Params), len(n.FunCall.Args)))
		}

		mismatch := false
		for i, arg := range n.FunCall.Args {
			if arg.GetTypeShallow() != fun.Params[i].Type {
				mismatch = true
				break
			}
		}
		if mismatch {
			got := ""
			expected := ""
			for _, arg := range n.FunCall.Args {
				got += arg.GetTypeShallow().Stringify() + " "
			}
			for _, param := range fun.Params {
				expected += param.Type.Stringify() + " "
			}

			msg := fmt.Sprintf("mismatched types in function call\n\tgot %s\n\texpected %s",
				got, expected)
			n.ReportHere(r, report.ReportNonfatal, msg)
		}

	case ast.NodeIf:
		checkNode(n.If.Exp, r)

		expType := n.If.Exp.GetTypeShallow()
		boolType := types.GetBuiltin(types.Bool)
		if expType != boolType {
			n.If.Exp.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), expType.Stringify()))
		}

		checkNode(n.If.IfBody, r)
		checkNode(n.If.ElseBody, r)

	case ast.NodeWhile:
		checkNode(n.While.Exp, r)
		checkNode(n.While.Body, r)

		expType := n.While.Exp.GetTypeShallow()
		boolType := types.GetBuiltin(types.Bool)
		if expType != boolType {
			n.While.Exp.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), expType.Stringify()))
		}

	case ast.NodeFor:
		checkNode(n.For.Init, r)
		checkNode(n.For.Cond, r)
		checkNode(n.For.Adv, r)
		checkNode(n.For.Body, r)

		condType := n.For.Cond.GetTypeShallow()
		boolType := types.GetBuiltin(types.Bool)
		if condType != boolType {
			n.For.Cond.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), condType.Stringify()))
		}

	case ast.NodeBlock:
		for _, node := range n.Block.Stmts {
			checkNode(node, r)
		}

	case ast.NodeCast:
		// Right now there are only integer types, so we can
		// convert between all of them. This code only checks
		// for new and unsupported types.

		from := n.Cast.What.GetTypeDeep()

		switch from {
		// Do nothing
		case types.GetBuiltin(types.S64):
		case types.GetBuiltin(types.U64):
		case types.GetBuiltin(types.Bool):

		case types.GetBuiltin(types.Void):
			voidType := types.GetBuiltin(types.Void)
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("can't cast from type %s", voidType.Stringify()))

		default:
			panic("not implemented")
		}

	// Do nothing
	case ast.NodeInteger:
	case ast.NodeBoolean:
	case ast.NodeVariable:
	case ast.NodeFunEx:
	case ast.NodeTypedef:
	case ast.NodeEmpty:

	default:
		panic("not implemented")
	}
}
