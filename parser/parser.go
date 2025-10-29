// This file contains the main parsing functions.

package parser

import (
	"clic/ast"
	"clic/report"
	"clic/symbol"
	"clic/types"
	"fmt"
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

		symbol.PushBlock()

		stmts := p.collectLists()
		n.Block.Stmts = stmts

		symbol.PopBlock()

	case tokenKeyword:
		switch p.consume().data {
		case "let":
			n.Tag = ast.NodeVariable
			n.Variable.IsDecl = true

			name, type_ := p.parseNameWithType()

			if symbol.ExistsInBlock(name, symbol.Variable) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"variable is already declared")
			} else {
				id := symbol.AddToBlock(name, symbol.Variable)

				s := symbol.Get(id)
				s.Variable.Type = type_
				symbol.Set(id, s)

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

			// TODO: Check if the identifier is not a type
			// or something
			rval := p.parseItem()
			n.BinOp.Rval = rval

			if symbol.ExistsInBlock(name, symbol.Variable) {
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
				id := symbol.AddToBlock(name, symbol.Variable)

				s := symbol.Get(id)
				s.Variable.Type = rval.GetTypeShallow()
				symbol.Set(id, s)

				lval.Id = id
				n.BinOp.Lval = &lval
			}

		case "exfun":
			n.Tag = ast.NodeFunEx

			name := p.match(tokenIdent).data

			params := []symbol.TypedIdent{}

			p.match(tokenTag('('))
			for p.peek(0).tag != tokenTag(')') {
				param := symbol.TypedIdent{}
				param.Name, param.Type = p.parseNameWithType()
				params = append(params, param)
			}
			p.match(tokenTag(')'))

			if symbol.ExistsAnywhere(name, symbol.Function) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"function is already declared")
			} else {
				id := symbol.AddToBlock(name, symbol.Function)

				s := symbol.Get(id)
				s.Function.Params = params
				symbol.Set(id, s)

				n.Id = id
			}

		case "defun":
			n.Tag = ast.NodeFunDef

			funName := p.match(tokenIdent).data

			if symbol.ExistsAnywhere(funName, symbol.Function) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"function is already declared")
			} else {
				id := symbol.AddToBlock(funName, symbol.Function)
				n.Id = id
			}

			symbol.PushBlock()

			// Adding parameters as local variables, so they can be
			// seen in the block. Keeping their ids so codegen can
			// assign offsets. Also adding parameters to the symbol
			// table so they can be seen in a call.

			// TODO: Why are parameters TypedIdent in the symbol
			// table?

			params := []symbol.TypedIdent{}

			p.match(tokenTag('('))
			for p.peek(0).tag != tokenTag(')') {
				paramName, type_ := p.parseNameWithType()

				params = append(params, symbol.TypedIdent{
					Name: paramName,
					Type: type_,
				})

				if symbol.ExistsInBlock(paramName, symbol.Variable) {
					n.ReportHere(p.r, report.ReportNonfatal,
						"duplicate parameter names")
				} else {
					id := symbol.AddToBlock(paramName, symbol.Variable)

					s := symbol.Get(id)
					s.Variable.Type = type_
					symbol.Set(id, s)

					n.Function.Params = append(n.Function.Params, id)
				}
			}
			p.match(tokenTag(')'))

			sym := symbol.Get(n.Id)
			sym.Function.Params = params
			symbol.Set(n.Id, sym)

			// n.Function.Type = p.parseType()

			for p.peek(0).tag == tokenTag('(') {
				stmt := p.parseList()
				if stmt.Tag == ast.NodeFunDef {
					stmt.ReportHere(p.r, report.ReportNonfatal,
						"nested functions are not allowed")
				}
				n.Function.Stmts = append(n.Function.Stmts, stmt)
			}

			symbol.PopBlock()

		case "if":
			n.Tag = ast.NodeIf

			n.If.Exp = p.parseItem()

			symbol.PushBlock()
			n.If.IfStmts = p.collectLists()
			symbol.PopBlock()

			{
				t := p.peek(0)
				if t.tag == tokenKeyword && t.data == "else" {
					p.match(tokenKeyword)
					symbol.PushBlock()
					n.If.ElseStmts = p.collectLists()
					symbol.PopBlock()
				}
			}

		case "else":
			n.ReportHere(p.r, report.ReportFatal,
				"unexpected keyword 'else'")

		case "while":
			n.Tag = ast.NodeWhile

			n.While.Exp = p.parseItem()

			symbol.PushBlock()
			n.While.Stmts = p.collectLists()
			symbol.PopBlock()

		case "for":
			n.Tag = ast.NodeFor

			symbol.PushBlock()
			n.For.Init = p.parseList()
			n.For.Cond = p.parseItem()
			n.For.Adv = p.parseList()
			n.For.Stmts = p.collectLists()
			symbol.PopBlock()

		case "typedef":
			n.Tag = ast.NodeTypedef

			name, toDef := p.parseNameWithType()

			if symbol.ExistsAnywhere(name, symbol.Type) {
				n.ReportHere(p.r, report.ReportNonfatal,
					"type is already declared")
			} else {
				size := types.Get(toDef).Size
				align := types.Get(toDef).Align
				typeNode := types.TypeNode{
					Tag:       types.Definition,
					DefinedAs: toDef,
					Size:      size,
					Align:     align,
				}
				def := types.Register(typeNode)

				id := symbol.AddToBlock(name, symbol.Type)

				s := symbol.Get(id)
				s.Type.Id = def
				symbol.Set(id, s)

				n.Id = id
			}

		default:
			// TODO: Throw fatal on unrecognized keyword
			panic("not implemented")
		}

	case tokenIdent:
		name := p.peek(0).data

		switch {
		case symbol.ExistsAnywhere(name, symbol.Type):
			n.Tag = ast.NodeCast
			n.Cast.To = p.parseType()
			n.Cast.What = p.parseItem()

		case symbol.ExistsAnywhere(name, symbol.Function):
			p.match(tokenIdent)
			n.Tag = ast.NodeFunCall

			id := symbol.LookupAnywhere(name, symbol.Function)
			if id == symbol.IdNone {
				panic("unreachable")
			}
			n.Id = id

			n.Function.Args = p.collectItems()

		default:
			n.ReportHere(p.r,
				report.ReportNonfatal,
				fmt.Sprintf("%s is not declared", name))
		}

	case tokenType:
		n.Tag = ast.NodeCast

		n.Cast.To = p.parseType()
		n.Cast.What = p.parseItem()

	default:
		s := lookahead.tag.stringify()
		n.ReportHere(p.r, report.ReportFatal,
			fmt.Sprintf("unexpected list head item: %s", s))
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

		id := symbol.LookupAnywhere(t.data, symbol.Variable)
		if id == symbol.IdNone {
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

		// TODO: Throw fatal on unrecognized keyword
		default:
			panic("not implemented")
		}

	case tokenTag('('):
		return p.parseList()

	default:
		s := lookahead.tag.stringify()
		n.ReportHere(p.r, report.ReportFatal,
			fmt.Sprintf("unexpected list item: %s", s))
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

func (p *Parser) parseType() types.Id {
	if p.peek(0).tag == tokenIdent {
		t := p.match(tokenIdent)

		name := t.data
		symId := symbol.LookupAnywhere(name, symbol.Type)

		typeId := types.IdNone

		if symId == symbol.IdNone {
			p.r.Report(report.Form{
				Tag:    report.ReportNonfatal,
				Line:   t.line,
				Column: t.column,
				Msg:    fmt.Sprintf("type '%s' is not defined", name),
			})
			// Returning IdNone, scary!
		} else {
			typeId = symbol.Get(symId).Type.Id
		}

		return typeId
	}

	t := p.match(tokenType)

	switch t.data {
	case "void":
		return types.GetBuiltin(types.Void)

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

func (p *Parser) parseNameWithType() (string, types.Id) {
	name := p.match(tokenIdent).data
	p.match(tokenTag(':'))
	type_ := p.parseType()
	return name, type_
}

func (p *Parser) collectItems() []*ast.Node {
	var items []*ast.Node

	for p.peek(0).tag != tokenTag(')') {
		items = append(items, p.parseItem())
	}

	return items
}

func (p *Parser) collectLists() []*ast.Node {
	var lists []*ast.Node

	for p.peek(0).tag == tokenTag('(') {
		lists = append(lists, p.parseList())
	}

	return lists
}
