// This file contains the main parsing functions.

package parser

import (
	"fmt"
	sym "github.com/zer0reaction/lisp-go/symbol"
	"strconv"
)

type Parser struct {
	data     string
	line     uint
	column   uint
	writeInd uint
	readInd  uint
	rbuffer  [ringSize]Token
}

func New(data string) *Parser {
	return &Parser{
		data:     data,
		line:     1,
		column:   1,
		writeInd: 0,
		readInd:  0,
	}
}

func (p *Parser) CreateAST() (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		lookahead, err := p.peek(0)
		if err != nil {
			return nil, err
		}
		if lookahead.Tag == TokenEOF {
			break
		}

		n, err := p.parseList()
		if err != nil {
			return nil, err
		}

		if tail == nil {
			head = n
			tail = n
		} else {
			tail.Next = n
			tail = tail.Next
		}
	}

	return head, nil
}

func (p *Parser) parseList() (*Node, error) {
	_, err := p.match(TokenTag('('))
	if err != nil {
		return nil, err
	}

	n := Node{}

	lookahead, err := p.peek(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Tag {
	case TokenTag('+'):
		err = p.parseBinOp(&n, BinOpSum)
		if err != nil {
			return nil, err
		}
	case TokenTag('-'):
		err = p.parseBinOp(&n, BinOpSub)
		if err != nil {
			return nil, err
		}
	case TokenColEq:
		t, err := p.peek(0)
		if err != nil {
			return nil, err
		}

		err = p.parseBinOp(&n, BinOpAssign)
		if err != nil {
			return nil, err
		}

		if n.BinOp.Lval.Tag != NodeVariable {
			return nil, fmt.Errorf(":%d:%d: error: lvalue is not a variable", t.Line, t.Column)
		}
	case TokenTag('('):
		n.Tag = NodeBlock

		sym.PushBlock()

		items, err := p.collectItems()
		if err != nil {
			return nil, err
		}
		n.Block.Start = items

		sym.PopBlock()
	case TokenLet:
		p.discard()

		n.Tag = NodeVariableDecl
		v := sym.Variable{}

		tp, err := p.consume()
		if err != nil {
			return nil, err
		}

		switch tp.Tag {
		case TokenS64:
			v.Type = sym.ValueS64
		case TokenU64:
			v.Type = sym.ValueU64
		default:
			return nil, fmt.Errorf(":%d:%d: error: expected type", tp.Line, tp.Column)
		}

		t, err := p.match(TokenIdent)
		if err != nil {
			return nil, err
		}
		v.Name = t.Data

		if sym.LookupInBlock(v.Name, sym.SymbolVariable) != sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block", t.Line, t.Column)
		}

		id := sym.AddSymbol(v.Name, sym.SymbolVariable)
		sym.SetVariable(id, v)
		n.Id = id
	case TokenExfun:
		p.discard()

		t, err := p.match(TokenIdent)
		if err != nil {
			return nil, err
		}

		n.Tag = NodeFunEx
		name := t.Data

		if sym.LookupGlobal(name, sym.SymbolFunction) != sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: function is already declared", t.Line, t.Column)
		}

		id := sym.AddSymbol(name, sym.SymbolFunction)
		sym.SetFunction(id, sym.Function{
			Name: name,
		})
		n.Id = id
	case TokenIdent:
		t, err := p.peek(0)
		if err != nil {
			return nil, err
		}

		p.discard()

		n.Tag = NodeFunCall

		id := sym.LookupGlobal(t.Data, sym.SymbolFunction)
		if id == sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: function is not declared", t.Line, t.Column)
		}
		n.Id = id

		items, err := p.collectItems()
		if err != nil {
			return nil, err
		}
		n.Function.ArgStart = items
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list head item", lookahead.Line, lookahead.Column)
	}

	_, err = p.match(TokenTag(')'))
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (p *Parser) parseBinOp(n *Node, tag BinOpTag) error {
	t, err := p.consume()
	if err != nil {
		return err
	}

	n.Tag = NodeBinOp
	n.BinOp.Tag = tag

	lval, err := p.parseItem()
	if err != nil {
		return err
	}
	rval, err := p.parseItem()
	if err != nil {
		return err
	}
	n.BinOp.Lval = lval
	n.BinOp.Rval = rval

	lvalType := n.BinOp.Lval.GetType()
	rvalType := n.BinOp.Rval.GetType()
	if lvalType != rvalType {
		return fmt.Errorf(":%d:%d: error: operand type mismatch", t.Line, t.Column)
	}

	return nil
}

func (p *Parser) collectItems() (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		lookahead, err := p.peek(0)
		if err != nil {
			return nil, err
		}
		if lookahead.Tag == TokenTag(')') {
			break
		}

		item, err := p.parseItem()
		if err != nil {
			return nil, err
		}

		if tail == nil {
			head = item
			tail = item
		} else {
			tail.Next = item
			tail = tail.Next
		}
	}

	return head, nil
}

func (p *Parser) parseItem() (*Node, error) {
	n := Node{}

	lookahead, err := p.peek(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Tag {
	case TokenInteger:
		t, err := p.peek(0)
		if err != nil {
			return nil, err
		}

		p.discard()

		n.Tag = NodeInteger
		// TODO: this is not clear, add a cast?
		n.Integer.Type = sym.ValueS64

		value, err := strconv.ParseInt(t.Data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value
	case TokenIdent:
		t, err := p.peek(0)
		if err != nil {
			return nil, err
		}

		p.discard()

		n.Tag = NodeVariable

		id := sym.LookupGlobal(t.Data, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope", t.Line, t.Column)
		}
		n.Id = id
	case TokenTag('('):
		return p.parseList()
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item", lookahead.Line, lookahead.Column)
	}

	return &n, nil
}
