// This file contains implementation of the type checker.

package types

// TODO: move type enum here

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/parser"
	"github.com/zer0reaction/lisp-go/report"
	sym "github.com/zer0reaction/lisp-go/symbol"
)

func TypeCheck(p *parser.Parser, n *parser.Node) {
	if n == nil {
		return
	}

	switch n.Tag {
	case parser.NodeBinOp:
		checkBinOp(p, n)
	case parser.NodeInteger:
		// do nothing
	case parser.NodeBoolean:
		// do nothing
	case parser.NodeBlock:
		TypeCheck(p, n.Block.Start)
	case parser.NodeVariableDecl:
		// do nothing
	case parser.NodeVariable:
		// do nothing
	case parser.NodeFunEx:
		// do nothing
	case parser.NodeFunCall:
		TypeCheck(p, n.Function.ArgStart)
	case parser.NodeIf:
		checkIf(p, n)
	default:
		panic(fmt.Sprintf("node not implemented [%d]", n.Tag))
	}

	TypeCheck(p, n.Next)
}

func checkBinOp(p *parser.Parser, n *parser.Node) {
	if n.Tag != parser.NodeBinOp {
		panic("incorrect node tag")
	}

	TypeCheck(p, n.BinOp.Lval)
	TypeCheck(p, n.BinOp.Rval)

	lvalType := n.BinOp.Lval.GetType()
	rvalType := n.BinOp.Rval.GetType()
	if lvalType != rvalType {
		p.ReportHere(n,
			report.ReportNonfatal,
			"operand type mismatch")
	}

	isAssign := (n.BinOp.Tag == parser.BinOpAssign)
	isStorage := (n.BinOp.Lval.Tag == parser.NodeVariable)
	if isAssign && !isStorage {
		p.ReportHere(n,
			report.ReportNonfatal,
			"lvalue is not a storage location")
	}
}

func checkIf(p *parser.Parser, n *parser.Node) {
	if n.Tag != parser.NodeIf {
		panic("incorrect node tag")
	}

	expType := n.If.Exp.GetType()
	if expType != sym.ValueBoolean {
		p.ReportHere(n,
			report.ReportNonfatal,
			"expected boolean type in expression")
	}

	TypeCheck(p, n.If.Body)
}
