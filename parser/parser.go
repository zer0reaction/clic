// This file contains the main parsing functions.

package parser

import (
	"github.com/zer0reaction/lisp-go/report"
	sym "github.com/zer0reaction/lisp-go/symbol"
	"strconv"
)

type lexer struct {
	data     string
	line     uint
	column   uint
	writeInd uint
	readInd  uint
	rbuffer  [ringSize]token
}

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

func (p *Parser) CreateAST() *Node {
	var head *Node = nil
	var tail *Node = nil

	for {
		lookahead := p.peek(0)
		if lookahead.tag == tokenEOF {
			break
		}

		n := p.parseList()

		if tail == nil {
			head = n
			tail = n
		} else {
			tail.Next = n
			tail = tail.Next
		}
	}

	return head
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
	case tokenTag('+'):
		p.parseBinOp(&n, BinOpSum)
	case tokenTag('-'):
		p.parseBinOp(&n, BinOpSub)
	case tokenColEq:
		p.parseBinOp(&n, BinOpAssign)

		if n.BinOp.Lval.Tag != NodeVariable {
			p.reportHere(&n,
				report.ReportNonfatal,
				"lvalue is not a variable")
		}
	case tokenTag('('):
		n.Tag = NodeBlock

		sym.PushBlock()

		items := p.collectItems()
		n.Block.Start = items

		sym.PopBlock()
	case tokenLet:
		p.discard()

		n.Tag = NodeVariableDecl
		v := sym.Variable{}

		switch p.consume().tag {
		case tokenS64:
			v.Type = sym.ValueS64
		case tokenU64:
			v.Type = sym.ValueU64
		default:
			p.reportHere(&n,
				report.ReportFatal,
				"expected type here")
		}

		t := p.match(tokenIdent)
		v.Name = t.data

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
		t := p.peek(0)

		p.discard()

		n.Tag = NodeFunCall

		id := sym.LookupGlobal(t.data, sym.SymbolFunction)
		if id == sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"function is not declared")
		}
		n.Id = id

		items := p.collectItems()
		n.Function.ArgStart = items
	default:
		p.reportHere(&n,
			report.ReportFatal,
			"incorrect list head item")
	}

	p.match(tokenTag(')'))

	return &n
}

func (p *Parser) parseBinOp(n *Node, tag BinOpTag) {
	p.discard()

	n.Tag = NodeBinOp
	n.BinOp.Tag = tag

	n.BinOp.Lval = p.parseItem()
	n.BinOp.Rval = p.parseItem()

	lvalType := n.BinOp.Lval.GetType()
	rvalType := n.BinOp.Rval.GetType()
	if lvalType != rvalType {
		p.reportHere(n,
			report.ReportNonfatal,
			"operand type mismatch")
	}
}

func (p *Parser) collectItems() *Node {
	var head *Node = nil
	var tail *Node = nil

	for {
		lookahead := p.peek(0)
		if lookahead.tag == tokenTag(')') {
			break
		}

		item := p.parseItem()

		if tail == nil {
			head = item
			tail = item
		} else {
			tail.Next = item
			tail = tail.Next
		}
	}

	return head
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
		n.Integer.Type = sym.ValueS64

		value, err := strconv.ParseInt(t.data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value
	case tokenIdent:
		t := p.peek(0)

		p.discard()

		n.Tag = NodeVariable

		id := sym.LookupGlobal(t.data, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			p.reportHere(&n,
				report.ReportNonfatal,
				"variable does not exist")
		}
		n.Id = id
	case tokenTag('('):
		return p.parseList()
	default:
		p.reportHere(&n,
			report.ReportFatal,
			"incorrect list item")
	}

	return &n
}
