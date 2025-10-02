package parser

import (
	"fmt"
	lex "github.com/zer0reaction/lisp-go/lexer"
	sym "github.com/zer0reaction/lisp-go/symbol"
	"strconv"
)

type Parser struct {
	lx *lex.Lexer
}

func New(lx *lex.Lexer) *Parser {
	p := Parser{
		lx: lx,
	}
	return &p
}

func (p *Parser) CreateAST() (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		if p.look().Tag == lex.TokenEOF {
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

func (p *Parser) look() lex.Token {
	t, err := p.lx.Peek(0)
	if err != nil {
		panic(err)
	}
	return t
}

func (p *Parser) parseList() (*Node, error) {
	_, err := p.lx.Match(lex.TokenTag('('))
	if err != nil {
		return nil, err
	}

	n := Node{}

	switch p.look().Tag {
	case lex.TokenTag('+'):
		err = p.parseBinOp(&n, BinOpSum)
		if err != nil {
			return nil, err
		}
	case lex.TokenTag('-'):
		err = p.parseBinOp(&n, BinOpSub)
		if err != nil {
			return nil, err
		}
	case lex.TokenColEq:
		t := p.look()

		err = p.parseBinOp(&n, BinOpAssign)
		if err != nil {
			return nil, err
		}

		if n.BinOp.Lval.Tag != NodeVariable {
			return nil, fmt.Errorf(":%d:%d: error: lvalue is not a variable", t.Line, t.Column)
		}
	case lex.TokenTag('('):
		n.Tag = NodeBlock

		sym.PushBlock()

		items, err := p.collectItems()
		if err != nil {
			return nil, err
		}
		n.Block.Start = items

		sym.PopBlock()
	case lex.TokenLet:
		p.lx.Discard()

		n.Tag = NodeVariableDecl
		v := sym.Variable{}

		tp, err := p.lx.Consume()
		if err != nil {
			return nil, err
		}

		switch tp.Tag {
		case lex.TokenS64:
			v.Type = sym.ValueS64
		case lex.TokenU64:
			v.Type = sym.ValueU64
		default:
			return nil, fmt.Errorf(":%d:%d: error: expected type", tp.Line, tp.Column)
		}

		t, err := p.lx.Match(lex.TokenIdent)
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
	case lex.TokenExfun:
		p.lx.Discard()

		t, err := p.lx.Match(lex.TokenIdent)
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
	case lex.TokenIdent:
		t := p.look()

		p.lx.Discard()

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
		return nil, fmt.Errorf(":%d:%d: error: incorrect list head item", p.look().Line, p.look().Column)
	}

	_, err = p.lx.Match(lex.TokenTag(')'))
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (p *Parser) parseBinOp(n *Node, tag BinOpTag) error {
	t, err := p.lx.Consume()
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

	for p.look().Tag != lex.TokenTag(')') {
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

	switch p.look().Tag {
	case lex.TokenInteger:
		t := p.look()

		p.lx.Discard()

		n.Tag = NodeInteger
		// TODO: this is not clear, add a cast?
		n.Integer.Type = sym.ValueS64

		value, err := strconv.ParseInt(t.Data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value
	case lex.TokenIdent:
		t := p.look()

		p.lx.Discard()

		n.Tag = NodeVariable

		id := sym.LookupGlobal(t.Data, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope", t.Line, t.Column)
		}
		n.Id = id
	case lex.TokenTag('('):
		return p.parseList()
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item", p.look().Line, p.look().Column)
	}

	return &n, nil
}
