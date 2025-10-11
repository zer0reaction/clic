// This file contains the main parsing functions.

package parser

import (
	"lisp-go/report"
	sym "lisp-go/symbol"
	"lisp-go/types"
	"strconv"
)

type Parser struct {
	fileName string

	l lexer
}

func New(fileName string, data string) *Parser {
	return &Parser{
		fileName: fileName,
		l: lexer{
			data:     data,
			line:     1,
			column:   1,
			writeInd: 0,
			readInd:  0,
		},
	}
}

func (p *Parser) CreateASTs() []*Node {
	var roots []*Node

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

func (p *Parser) reportHere(n *Node, tag report.ReportTag, msg string) {
	report.Report(report.Form{
		Tag:    tag,
		File:   p.fileName,
		Line:   n.Line,
		Column: n.Column,
		Msg:    msg,
	})
}

func (p *Parser) parseList() *Node {
	p.match(tokenTag('('))

	lookahead := p.peek(0)
	n := Node{
		Line:   lookahead.line,
		Column: lookahead.column,
	}

	switch lookahead.tag {
	case tokenBinOp:
		p.parseBinOp(&n)

	case tokenTag('('):
		n.Tag = NodeBlock

		sym.PushBlock()

		items := p.collectItems()
		n.Block.Stmts = items

		sym.PopBlock()

	case tokenLet:
		p.discard()

		n.Tag = NodeVariableDecl

		v := sym.Variable{}

		v.Name = p.match(tokenIdent).data
		p.match(tokenTag(':'))
		v.Type = p.parseType()

		if sym.LookupInBlock(v.Name, sym.SymbolVariable) != sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"variable is already declared")
		}

		id := sym.AddSymbol(v.Name, sym.SymbolVariable)
		sym.SetVariable(id, v)
		n.Id = id

	case tokenExfun:
		p.discard()

		t := p.match(tokenIdent)

		n.Tag = NodeFunEx
		name := t.data

		if sym.LookupGlobal(name, sym.SymbolFunction) != sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"function is already declared")
		}

		id := sym.AddSymbol(name, sym.SymbolFunction)
		sym.SetFunction(id, sym.Function{
			Name: name,
		})
		n.Id = id

	case tokenIdent:
		t := p.consume()

		n.Tag = NodeFunCall

		id := sym.LookupGlobal(t.data, sym.SymbolFunction)
		if id == sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"function is not declared")
		}
		n.Id = id

		items := p.collectItems()
		n.FunCall.Args = items

	case tokenIf:
		n.Tag = NodeIf

		p.discard()
		n.If.Exp = p.parseItem()
		n.If.IfBody = p.parseList()

		if p.peek(0).tag == tokenElse {
			p.discard()
			n.If.ElseBody = p.parseList()
		}

	case tokenWhile:
		n.Tag = NodeWhile

		p.discard()
		n.While.Exp = p.parseItem()
		n.While.Body = p.parseList()

	default:
		p.reportHere(&n,
			report.ReportFatal,
			"incorrect list head item")
	}

	p.match(tokenTag(')'))

	return &n
}

func (p *Parser) parseItem() *Node {
	lookahead := p.peek(0)
	n := Node{
		Line:   lookahead.line,
		Column: lookahead.column,
	}

	switch lookahead.tag {
	case tokenInteger:
		t := p.peek(0)

		p.discard()

		n.Tag = NodeInteger
		// TODO: this is not clear, add a cast?
		n.Integer.Type = types.S64

		value, err := strconv.ParseInt(t.data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value

	case tokenIdent:
		t := p.consume()

		n.Tag = NodeVariable

		id := sym.LookupGlobal(t.data, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"variable does not exist")
		}
		n.Id = id

	case tokenTrue:
		p.discard()
		n.Tag = NodeBoolean
		n.Boolean.Value = true

	case tokenFalse:
		p.discard()
		n.Tag = NodeBoolean
		n.Boolean.Value = false

	case tokenTag('('):
		return p.parseList()

	default:
		p.reportHere(&n,
			report.ReportFatal,
			"incorrect list item")
	}

	return &n
}

func (p *Parser) parseBinOp(n *Node) {
	n.Tag = NodeBinOp

	switch p.consume().data {
	case ":=":
		n.BinOp.Tag = BinOpAssign

	case "+":
		n.BinOp.Tag = BinOpArith
		n.BinOp.ArithTag = BinOpSum

	case "-":
		n.BinOp.Tag = BinOpArith
		n.BinOp.ArithTag = BinOpSub

	case "==":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpEq

	case "!=":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpNeq

	case "<=":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpLessEq

	case "<":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpLess

	case ">=":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpGreatEq

	case ">":
		n.BinOp.Tag = BinOpComp
		n.BinOp.CompTag = BinOpGreat

	default:
		panic("not implemented")
	}

	n.BinOp.Lval = p.parseItem()
	n.BinOp.Rval = p.parseItem()
}

func (p *Parser) parseType() types.Type {
	t := p.consume()
	if t.tag != tokenType {
		report.Report(report.Form{
			Tag:    report.ReportFatal,
			File:   p.fileName,
			Line:   t.line,
			Column: t.column,
			Msg:    "expected type",
		})
	}

	switch t.data {
	case "s64":
		return types.S64

	case "u64":
		return types.U64

	default:
		panic("not implemented")
	}
}

func (p *Parser) collectItems() []*Node {
	var items []*Node

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
