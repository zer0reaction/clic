// This file contains the main parsing functions.

package parser

import (
	"lisp-go/ast"
	"lisp-go/report"
	sym "lisp-go/symbol"
	"lisp-go/types"
	"strconv"
)

type Parser struct {
	l lexer
	r *report.Reporter
}

func New(data string, r *report.Reporter) *Parser {
	return &Parser{
		l: lexer{
			data:     data,
			line:     1,
			column:   1,
			writeInd: 0,
			readInd:  0,
		},
		r: r,
	}
}

func (p *Parser) CreateASTs() []*ast.Node {
	var roots []*ast.Node

	for {
		lookahead := p.peek(0)
		if lookahead.tag == tokenEOF {
			break
		}

		node := p.parseList()
		roots = append(roots, node)
	}

	return roots
}

func (p *Parser) parseList() *ast.Node {
	p.match(tokenTag('('))

	lookahead := p.peek(0)
	n := ast.Node{
		Line:   lookahead.line,
		Column: lookahead.column,
	}

	switch lookahead.tag {
	case tokenBinOp:
		p.parseBinOp(&n)

	case tokenTag(')'):
		n.Tag = ast.NodeEmpty
		// Matched at the end of function

	case tokenTag('('):
		n.Tag = ast.NodeBlock

		sym.PushBlock()

		items := p.collectItems()
		n.Block.Stmts = items

		sym.PopBlock()

	case tokenKeyword:
		switch p.consume().data {
		case "let":
			n.Tag = ast.NodeVariable
			n.Variable.IsDecl = true

			name, type_ := p.parseNameWithType()

			if sym.ExistsInBlock(name, sym.SymbolVariable) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"variable is already declared")
			} else {
				id := sym.AddSymbol(name, sym.SymbolVariable)
				sym.SetVariable(id, sym.Variable{Type: type_})
				n.Id = id
			}

		case "auto":
			// Type is inferred.
			//              n
			//             / \
			// new variable   item

			n.Tag = ast.NodeBinOp
			n.BinOp.Tag = ast.BinOpAssign

			ident := p.match(tokenIdent)
			name := ident.data

			p.match(tokenTag(':'))

			rval := p.parseItem()
			n.BinOp.Rval = rval

			if sym.ExistsInBlock(name, sym.SymbolVariable) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"variable is already declared")
			} else {
				lval := ast.Node{
					Tag:    ast.NodeVariable,
					Line:   ident.line,
					Column: ident.column,
				}
				lval.Variable.IsDecl = true

				// Type is determined in the parser,
				// so it should be known by now.
				id := sym.AddSymbol(name, sym.SymbolVariable)
				sym.SetVariable(id, sym.Variable{Type: rval.GetType()})
				lval.Id = id

				n.BinOp.Lval = &lval
			}

		case "exfun":
			n.Tag = ast.NodeFunEx
			fun := sym.Function{}

			name := p.match(tokenIdent).data

			p.match(tokenTag('('))

			for p.peek(0).tag != tokenTag(')') {
				param := sym.TypedIdent{}
				param.Name, param.Type = p.parseNameWithType()
				fun.Params = append(fun.Params, param)
			}

			p.match(tokenTag(')'))

			if sym.ExistsAnywhere(name, sym.SymbolFunction) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"function is already declared")
			} else {
				id := sym.AddSymbol(name, sym.SymbolFunction)
				sym.SetFunction(id, fun)
				n.Id = id
			}

		case "if":
			n.Tag = ast.NodeIf

			n.If.Exp = p.parseItem()
			n.If.IfBody = p.parseList()

			if p.peek(0).tag != tokenTag(')') {
				n.If.ElseBody = p.parseList()
			}

		case "while":
			n.Tag = ast.NodeWhile

			n.While.Exp = p.parseItem()
			n.While.Body = p.parseList()

		default:
			panic("not implemented")
		}

	case tokenIdent:
		t := p.consume()

		n.Tag = ast.NodeFunCall

		id := sym.LookupAnywhere(t.data, sym.SymbolFunction)
		if id == sym.SymbolIdNone {
			n.ReportHere(p.r,
				report.ReportNonfatal,
				"function is not declared")
		}
		n.Id = id

		n.FunCall.Args = p.collectItems()

	case tokenType:
		n.Tag = ast.NodeCast

		n.Cast.To = p.parseType()
		n.Cast.What = p.parseItem()

	default:
		n.ReportHere(p.r, report.ReportFatal,
			"incorrect list head item")
	}

	p.match(tokenTag(')'))

	return &n
}

func (p *Parser) parseItem() *ast.Node {
	lookahead := p.peek(0)
	n := ast.Node{
		Line:   lookahead.line,
		Column: lookahead.column,
	}

	switch lookahead.tag {
	case tokenInteger:
		t := p.peek(0)

		p.discard()

		n.Tag = ast.NodeInteger

		// TODO: There is currently no way to type u64 literals
		value, err := strconv.ParseInt(t.data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}

		// By default all integer literals are s64
		n.Integer.Size = 64
		n.Integer.Signed = true
		n.Integer.SValue = value

	case tokenIdent:
		t := p.consume()

		n.Tag = ast.NodeVariable

		id := sym.LookupAnywhere(t.data, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			n.ReportHere(p.r, report.ReportNonfatal,
				"variable does not exist")
		}
		n.Id = id

	case tokenKeyword:
		switch p.consume().data {
		case "true":
			n.Tag = ast.NodeBoolean
			n.Boolean.Value = true

		case "false":
			n.Tag = ast.NodeBoolean
			n.Boolean.Value = false

		default:
			panic("not implemented")
		}

	case tokenTag('('):
		return p.parseList()

	default:
		n.ReportHere(p.r, report.ReportFatal,
			"incorrect list item")
	}

	return &n
}

func (p *Parser) parseBinOp(n *ast.Node) {
	n.Tag = ast.NodeBinOp

	switch p.consume().data {
	case ":=":
		n.BinOp.Tag = ast.BinOpAssign

	case "+":
		n.BinOp.Tag = ast.BinOpArith
		n.BinOp.ArithTag = ast.BinOpSum

	case "-":
		n.BinOp.Tag = ast.BinOpArith
		n.BinOp.ArithTag = ast.BinOpSub

	case "*":
		n.BinOp.Tag = ast.BinOpArith
		n.BinOp.ArithTag = ast.BinOpMult

	case "/":
		n.BinOp.Tag = ast.BinOpArith
		n.BinOp.ArithTag = ast.BinOpDiv

	case "%":
		n.BinOp.Tag = ast.BinOpArith
		n.BinOp.ArithTag = ast.BinOpMod

	case "==":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpEq

	case "!=":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpNeq

	case "<=":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpLessEq

	case "<":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpLess

	case ">=":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpGreatEq

	case ">":
		n.BinOp.Tag = ast.BinOpComp
		n.BinOp.CompTag = ast.BinOpGreat

	default:
		panic("not implemented")
	}

	n.BinOp.Lval = p.parseItem()
	n.BinOp.Rval = p.parseItem()
}

func (p *Parser) parseType() types.TypeId {
	t := p.match(tokenType)

	switch t.data {
	case "s64":
		return types.GetBuiltin(types.S64)

	case "u64":
		return types.GetBuiltin(types.U64)

	case "bool":
		return types.GetBuiltin(types.Bool)

	case "struct":
		struct_ := types.TypeNode{Tag: types.Struct}

		p.match(tokenTag('('))

		for p.peek(0).tag != tokenTag(')') {
			name, type_ := p.parseNameWithType()
			field := types.Field{Type: type_, Name: name}
			struct_.Fields = append(struct_.Fields, field)
		}

		p.match(tokenTag(')'))

		return types.Register(struct_)

	default:
		panic("not implemented")
	}
}

func (p *Parser) parseNameWithType() (string, types.TypeId) {
	name := p.match(tokenIdent).data
	p.match(tokenTag(':'))
	type_ := p.parseType()
	return name, type_
}

func (p *Parser) collectItems() []*ast.Node {
	var items []*ast.Node

	for {
		lookahead := p.peek(0)
		if lookahead.tag == tokenTag(')') {
			break
		}

		item := p.parseItem()
		items = append(items, item)
	}

	return items
}
