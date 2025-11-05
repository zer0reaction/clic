package parser

import (
	"clic/ast"
	"clic/report"
	"clic/symbol"
	"clic/types"
	"fmt"
	"strconv"
)

type parserState uint

const (
	stateError parserState = iota
	inGlobal
	inFunction
)

type Parser struct {
	l lexer
	t *symbol.Table
	r *report.Reporter

	state    parserState
	function symbol.Id
}

func New(data string, t *symbol.Table, r *report.Reporter) *Parser {
	return &Parser{
		l: lexer{
			data:     data,
			line:     1,
			column:   1,
			writeInd: 0,
			readInd:  0,
		},
		t:     t,
		r:     r,
		state: inGlobal,
	}
}

func (p *Parser) CreateASTs() []*ast.Node {
	var roots []*ast.Node

	p.t.PushScope()
	for {
		lookahead := p.peek(0)
		if lookahead.tag == tokenEOF {
			break
		}

		node := p.parseList()
		roots = append(roots, node)
	}
	p.t.PopScope()

	return roots
}

func (p *Parser) parseList() *ast.Node {
	p.match(tokenTag('('))

	lookahead := p.peek(0)
	n := ast.Node{
		Id:     symbol.IdNone,
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
		n.Tag = ast.NodeScope

		p.t.PushScope()

		for p.peek(0).tag == tokenTag('(') {
			n.Scope.Stmts = append(n.Scope.Stmts, p.parseList())
		}

		p.t.PopScope()

	case tokenKeyword:
		switch p.consume().data {
		case "let":
			n.Tag = ast.NodeLVarDecl

			if p.state == inGlobal {
				panic("not implemented")
			}

			name, typ := p.parseNameWithType()

			id, added := p.t.Add(name, symbol.LVar)

			if added {
				n.Id = id
				sym := p.t.Get(id)
				sym.Type = typ
				p.t.Set(id, sym)
			} else {
				n.ReportHere(p.r, report.ReportNonfatal,
					"local variable is already declared in the current scope")
			}

		case "auto":
			// Type is inferred.
			//              n
			//             / \
			// new variable   item

			if p.state == inGlobal {
				panic("not implemented")
			}

			n.Tag = ast.NodeBinOp
			n.BinOp.Tag = ast.BinOpAssign

			ident := p.match(tokenIdent)
			name := ident.data

			// TODO: Check if the identifier is not a type
			// or something
			rval := p.parseItem()
			n.BinOp.Rval = rval

			id, added := p.t.Add(name, symbol.LVar)

			if added {
				lval := ast.Node{
					Tag:    ast.NodeLVarDecl,
					Id:     id,
					Line:   ident.line,
					Column: ident.column,
				}

				sym := p.t.Get(id)
				// Type is determined in the parser, so it should be
				// known by now.
				sym.Type = rval.GetTypeShallow(p.t)
				p.t.Set(id, sym)

				n.BinOp.Lval = &lval
			} else {
				n.ReportHere(p.r, report.ReportNonfatal,
					"local variable is already declared in the current scope")
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

			typ := p.parseType()

			id, added := p.t.Add(name, symbol.Fun)

			if added {
				n.Id = id
				sym := p.t.Get(id)
				sym.Type = typ
				sym.Fun.Params = params
				p.t.Set(id, sym)
			} else {
				n.ReportHere(p.r, report.ReportNonfatal,
					"function is already declared")
			}

		case "defun":
			n.Tag = ast.NodeFunDef

			funName := p.match(tokenIdent).data

			funId, funAdded := p.t.Add(funName, symbol.Fun)

			if funAdded {
				p.state = inFunction
				p.function = funId
			} else {
				n.ReportHere(p.r, report.ReportNonfatal,
					"function is already declared")
			}

			p.t.PushScope()

			// Adding parameters as local variables, so they can be
			// seen in the scope. Keeping their ids so codegen can
			// assign offsets. Also adding parameters to the symbol
			// table so they can be seen in a call.

			params := []symbol.TypedIdent{}
			p.match(tokenTag('('))
			for p.peek(0).tag != tokenTag(')') {
				paramName, paramType := p.parseNameWithType()

				params = append(params, symbol.TypedIdent{
					Name: paramName,
					Type: paramType,
				})

				paramId, paramAdded := p.t.Add(paramName, symbol.LVar)
				if paramAdded {
					sym := p.t.Get(paramId)
					sym.Type = paramType
					p.t.Set(paramId, sym)

					n.Fun.Params = append(n.Fun.Params, paramId)
				} else {
					n.ReportHere(p.r, report.ReportNonfatal,
						"duplicate parameter names")
				}
			}
			p.match(tokenTag(')'))

			funType := p.parseType()

			if funAdded {
				n.Id = funId

				sym := p.t.Get(funId)
				sym.Type = funType
				sym.Fun.Params = params
				p.t.Set(funId, sym)
			}

			for p.peek(0).tag == tokenTag('(') {
				stmt := p.parseList()
				if stmt.Tag == ast.NodeFunDef {
					stmt.ReportHere(p.r, report.ReportNonfatal,
						"nested functions are not allowed")
				}
				n.Fun.Stmts = append(n.Fun.Stmts, stmt)
			}

			p.t.PopScope()
			p.state = inGlobal

		case "return":
			n.Tag = ast.NodeReturn

			if p.state != inFunction {
				n.ReportHere(p.r, report.ReportNonfatal,
					"return found outside function")
			}

			n.Return.Val = p.parseItem()
			n.Return.Fun = p.function

		case "if":
			n.Tag = ast.NodeIf

			n.If.Exp = p.parseItem()

			p.t.PushScope()
			for p.peek(0).tag == tokenTag('(') {
				n.If.IfStmts = append(n.If.IfStmts, p.parseList())
			}
			p.t.PopScope()

			{
				t := p.peek(0)
				if t.tag == tokenKeyword && t.data == "else" {
					p.match(tokenKeyword)
					p.t.PushScope()
					for p.peek(0).tag == tokenTag('(') {
						n.If.ElseStmts = append(n.If.ElseStmts, p.parseList())
					}
					p.t.PopScope()
				}
			}

		case "else":
			n.ReportHere(p.r, report.ReportFatal,
				"unexpected keyword 'else'")

		case "while":
			n.Tag = ast.NodeWhile

			n.While.Exp = p.parseItem()

			p.t.PushScope()
			for p.peek(0).tag == tokenTag('(') {
				n.While.Stmts = append(n.While.Stmts, p.parseList())
			}
			p.t.PopScope()

		case "for":
			n.Tag = ast.NodeFor

			p.t.PushScope()
			n.For.Init = p.parseList()
			n.For.Cond = p.parseItem()
			n.For.Adv = p.parseList()
			for p.peek(0).tag == tokenTag('(') {
				n.For.Stmts = append(n.For.Stmts, p.parseList())
			}
			p.t.PopScope()

		case "typedef":
			n.Tag = ast.NodeTypedef

			name, toDef := p.parseNameWithType()

			id, added := p.t.Add(name, symbol.Type)

			if added {
				n.Id = id

				toDefNode := types.Get(toDef)
				defNode := types.TypeNode{
					Tag:       types.Definition,
					DefinedAs: toDef,
					Size:      toDefNode.Size,
					Align:     toDefNode.Align,
				}
				def := types.Register(defNode)

				sym := p.t.Get(id)
				sym.Type = def
				p.t.Set(id, sym)
			} else {
				n.ReportHere(p.r, report.ReportNonfatal,
					"type is already declared in the current scope")
			}

		default:
			// TODO: Throw fatal on unrecognized keyword
			panic("not implemented")
		}

	case tokenIdent:
		name := p.match(tokenIdent).data
		id, exists := p.t.Resolve(name)

		if exists {
			sym := p.t.Get(id)

			switch sym.Tag {
			case symbol.Type:
				n.Tag = ast.NodeCast
				n.Cast.To = sym.Type
				n.Cast.What = p.parseItem()

			case symbol.Fun:
				n.Tag = ast.NodeFunCall
				n.Id = id
				for p.peek(0).tag != tokenTag(')') {
					n.Fun.Args = append(n.Fun.Args, p.parseItem())
				}

			default:
				n.ReportHere(p.r, report.ReportNonfatal,
					"unexpected identifier")
			}
		} else {
			n.ReportHere(p.r, report.ReportNonfatal,
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
		Id:     symbol.IdNone,
		Line:   lookahead.line,
		Column: lookahead.column,
	}

	switch lookahead.tag {
	case tokenInt:
		t := p.match(tokenInt)

		n.Tag = ast.NodeInt

		// TODO: There is currently no way to type u64 literals
		value, err := strconv.ParseInt(t.data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}

		// By default all integer literals are s64
		n.Int.Size = 64
		n.Int.Signed = true
		n.Int.SValue = value

	case tokenIdent:
		t := p.consume()

		// TODO: Add global variables
		n.Tag = ast.NodeLVar

		id, exists := p.t.ResolveWithTag(t.data, symbol.LVar)

		if exists {
			n.Id = id
		} else {
			n.ReportHere(p.r, report.ReportNonfatal,
				"local variable does not exist")
		}

	case tokenKeyword:
		switch p.consume().data {
		case "true":
			n.Tag = ast.NodeBool
			n.Bool.Value = true

		case "false":
			n.Tag = ast.NodeBool
			n.Bool.Value = false

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

		id, exists := p.t.ResolveWithTag(name, symbol.Type)

		if exists {
			sym := p.t.Get(id)
			return sym.Type
		} else {
			p.r.Report(report.Form{
				Tag:    report.ReportNonfatal,
				Line:   t.line,
				Column: t.column,
				Msg:    fmt.Sprintf("type '%s' does not exist in the current scope", name),
			})
			// TODO: Returning IdNone looks scary
			return types.IdNone
		}
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
