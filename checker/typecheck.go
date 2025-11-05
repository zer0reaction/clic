// This file contains the type checking functions.

package checker

import (
	"clic/ast"
	"clic/report"
	"clic/symbol"
	"clic/types"
	"fmt"
)

func TypeCheck(roots []*ast.Node, t *symbol.Table, r *report.Reporter) {
	for _, node := range roots {
		checkNode(node, t, r)
	}
}

func checkNode(n *ast.Node, t *symbol.Table, r *report.Reporter) {
	if n == nil {
		return
	}

	switch n.Tag {
	case ast.NodeBinOp:
		checkNode(n.BinOp.Lval, t, r)
		checkNode(n.BinOp.Rval, t, r)

		lvalType := n.BinOp.Lval.GetTypeShallow(t)
		rvalType := n.BinOp.Rval.GetTypeShallow(t)
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
		isStorage := ((n.BinOp.Lval.Tag == ast.NodeLVar) || (n.BinOp.Lval.Tag == ast.NodeLVarDecl))
		if isAssign && !isStorage {
			n.ReportHere(r, report.ReportNonfatal,
				"lvalue is not a storage location")
		}

	case ast.NodeFunCall:
		for _, node := range n.Fun.Args {
			checkNode(node, t, r)
		}

		fun := t.Get(n.Id).Fun

		if len(n.Fun.Args) != len(fun.Params) {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected %d arguments, got %d", len(fun.Params), len(n.Fun.Args)))
		}

		mismatch := false
		for i, arg := range n.Fun.Args {
			if arg.GetTypeShallow(t) != fun.Params[i].Type {
				mismatch = true
				break
			}
		}
		if mismatch {
			got := ""
			expected := ""
			for _, arg := range n.Fun.Args {
				got += arg.GetTypeShallow(t).Stringify() + " "
			}
			for _, param := range fun.Params {
				expected += param.Type.Stringify() + " "
			}

			msg := fmt.Sprintf("mismatched types in function call\n\tgot %s\n\texpected %s",
				got, expected)
			n.ReportHere(r, report.ReportNonfatal, msg)
		}

	case ast.NodeIf:
		checkNode(n.If.Exp, t, r)

		expType := n.If.Exp.GetTypeShallow(t)
		boolType := types.GetBuiltin(types.Bool)
		if expType != boolType {
			n.If.Exp.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), expType.Stringify()))
		}

		for _, stmt := range n.If.IfStmts {
			checkNode(stmt, t, r)
		}
		for _, stmt := range n.If.ElseStmts {
			checkNode(stmt, t, r)
		}

	case ast.NodeWhile:
		checkNode(n.While.Exp, t, r)
		for _, stmt := range n.While.Stmts {
			checkNode(stmt, t, r)
		}

		expType := n.While.Exp.GetTypeShallow(t)
		boolType := types.GetBuiltin(types.Bool)
		if expType != boolType {
			n.While.Exp.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), expType.Stringify()))
		}

	case ast.NodeFor:
		checkNode(n.For.Init, t, r)
		checkNode(n.For.Cond, t, r)
		checkNode(n.For.Adv, t, r)
		for _, stmt := range n.For.Stmts {
			checkNode(stmt, t, r)
		}

		condType := n.For.Cond.GetTypeShallow(t)
		boolType := types.GetBuiltin(types.Bool)
		if condType != boolType {
			n.For.Cond.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					boolType.Stringify(), condType.Stringify()))
		}

	case ast.NodeScope:
		for _, node := range n.Scope.Stmts {
			checkNode(node, t, r)
		}

	case ast.NodeCast:
		// Right now there are only integer types, so we can
		// convert between all of them. This code only checks
		// for new and unsupported types.

		from := n.Cast.What.GetTypeDeep(t)

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

	case ast.NodeLVarDecl:
		voidType := types.GetBuiltin(types.Void)
		varType := n.GetTypeDeep(t)

		if varType == voidType {
			n.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("variable of type %s", voidType.Stringify()))
		}

	case ast.NodeFunDef:
		// TODO: Add check for void
		for _, stmt := range n.Fun.Stmts {
			checkNode(stmt, t, r)
		}

	// TODO: Add check for void
	case ast.NodeReturn:
		checkNode(n.Return.Val, t, r)

		funType := t.Get(n.Return.Fun).Type
		valType := n.Return.Val.GetTypeShallow(t)

		if funType != valType {
			n.Return.Val.ReportHere(r, report.ReportNonfatal,
				fmt.Sprintf("expected type %s, got %s",
					funType.Stringify(), valType.Stringify()))
		}

	// Do nothing
	case ast.NodeInt:
	case ast.NodeLVar:
	case ast.NodeBool:
	case ast.NodeFunEx: // TODO: Add check for 'void' params
	case ast.NodeTypedef: // TODO: Add check for 'void'
	case ast.NodeEmpty:

	default:
		panic("not implemented")
	}
}
