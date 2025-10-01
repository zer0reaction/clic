package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	sym "github.com/zer0reaction/lisp-go/symbol"
	"strconv"
)

func CreateAST(lx *lexer.Lexer) (*Node, error) {
	var head *Node = nil
	var tail *Node = nil

	for {
		t, err := lx.Peek(0)
		if err != nil {
			return nil, err
		}
		if t.Tag == lexer.TokenEOF {
			break
		}

		n, err := parseList(lx)
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

func parseList(lx *lexer.Lexer) (*Node, error) {
	err := lx.Match(lexer.TokenTag('('))
	if err != nil {
		return nil, err
	}

	lookahead, err := lx.Peek(0)
	if err != nil {
		return nil, err
	}

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenTag('+'):
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpSum

		t, err := lx.Chop(lexer.TokenTag('+'))
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx)
		if err != nil {
			return nil, err
		}

		if !binOpTypeCheck(&n) {
			return nil, fmt.Errorf(":%d:%d: error: operand type mismatch",
				t.Line, t.Column)
		}
	case lexer.TokenTag('-'):
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpSub

		t, err := lx.Chop(lexer.TokenTag('-'))
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx)
		if err != nil {
			return nil, err
		}

		if !binOpTypeCheck(&n) {
			return nil, fmt.Errorf(":%d:%d: error: operand type mismatch",
				t.Line, t.Column)
		}
	case lexer.TokenColEq:
		n.Tag = NodeBinOp
		n.BinOp.Tag = BinOpAssign

		t, err := lx.Chop(lexer.TokenColEq)
		if err != nil {
			return nil, err
		}

		err = parseBinOp(&n, lx)
		if err != nil {
			return nil, err
		}

		if n.BinOp.Lval.Tag != NodeVariable {
			return nil, fmt.Errorf(":%d:%d: error: lvalue is not a variable",
				t.Line, t.Column)
		}

		if !binOpTypeCheck(&n) {
			return nil, fmt.Errorf(":%d:%d: error: operand type mismatch",
				t.Line, t.Column)
		}
	case lexer.TokenTag('('):
		n.Tag = NodeBlock

		sym.PushBlock()

		items, err := collectItems(lx)
		if err != nil {
			return nil, err
		}
		n.Block.Start = items

		sym.PopBlock()
	case lexer.TokenLet:
		n.Tag = NodeVariableDecl
		v := sym.Variable{}

		err := lx.Match(lexer.TokenLet)
		if err != nil {
			return nil, err
		}

		tp, err := lx.Peek(0)
		if err != nil {
			return nil, err
		}

		switch tp.Tag {
		case lexer.TokenS64:
			v.Type = sym.ValueS64
			err := lx.Match(lexer.TokenS64)
			if err != nil {
				return nil, err
			}
		case lexer.TokenU64:
			v.Type = sym.ValueU64
			err := lx.Match(lexer.TokenU64)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf(":%d:%d: error: expected type",
				tp.Line, tp.Column)
		}

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		v.Name = t.Data

		if sym.LookupInBlock(v.Name, sym.SymbolVariable) != sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: variable is already declared in the current block",
				t.Line, t.Column)
		}

		id := sym.AddSymbol(v.Name, sym.SymbolVariable)
		sym.SetVariable(id, v)
		n.Variable.Id = id
	case lexer.TokenExfun:
		n.Tag = NodeFunEx

		err := lx.Match(lexer.TokenExfun)
		if err != nil {
			return nil, err
		}

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := t.Data

		if sym.LookupGlobal(name, sym.SymbolFunction) != sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: function is already declared",
				t.Line, t.Column)
		}

		id := sym.AddSymbol(name, sym.SymbolFunction)
		sym.SetFunction(id, sym.Function{
			Name: name,
		})
		n.Function.Id = id
	case lexer.TokenIdent:
		n.Tag = NodeFunCall

		t, err := lx.Chop(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := t.Data

		id := sym.LookupGlobal(name, sym.SymbolFunction)
		if id == sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: function is not declared",
				t.Line, t.Column)
		}
		n.Function.Id = id

		items, err := collectItems(lx)
		if err != nil {
			return nil, err
		}
		n.Function.ArgStart = items
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list head item",
			lookahead.Line, lookahead.Column)
	}

	err = lx.Match(lexer.TokenTag(')'))
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func parseBinOp(n *Node, lx *lexer.Lexer) error {
	lval, err := parseItem(lx)
	if err != nil {
		return err
	}
	rval, err := parseItem(lx)
	if err != nil {
		return err
	}
	n.BinOp.Lval = lval
	n.BinOp.Rval = rval

	return nil
}

func collectItems(lx *lexer.Lexer) (*Node, error) {
	lookahead, err := lx.Peek(0)
	if err != nil {
		return nil, err
	}

	var head *Node = nil
	var tail *Node = nil

	for lookahead.Tag != lexer.TokenTag(')') {
		item, err := parseItem(lx)
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

		lookahead, err = lx.Peek(0)
		if err != nil {
			return nil, err
		}
	}

	return head, nil
}

func parseItem(lx *lexer.Lexer) (*Node, error) {
	lookahead, err := lx.Peek(0)
	if err != nil {
		return nil, err
	}

	n := Node{}

	switch lookahead.Tag {
	case lexer.TokenInteger:
		n.Tag = NodeInteger
		// TODO: this is not clear, add a cast?
		n.Integer.Type = sym.ValueS64

		value, err := strconv.ParseInt(lookahead.Data, 0, 64)
		if err != nil {
			panic("incorrect integer data")
		}
		n.Integer.Value = value

		err = lx.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}
	case lexer.TokenIdent:
		n.Tag = NodeVariable
		name := lookahead.Data

		id := sym.LookupGlobal(name, sym.SymbolVariable)
		if id == sym.SymbolIdNone {
			return nil, fmt.Errorf(":%d:%d: error: variable does not exist in the current scope",
				lookahead.Line, lookahead.Column)
		}
		n.Variable.Id = id

		err = lx.Match(lexer.TokenIdent)
		if err != nil {
			return nil, err
		}
	case lexer.TokenTag('('):
		return parseList(lx)
	default:
		return nil, fmt.Errorf(":%d:%d: error: incorrect list item",
			lookahead.Line, lookahead.Column)
	}

	return &n, nil
}

func binOpTypeCheck(n *Node) bool {
	if n.Tag != NodeBinOp {
		panic("node is not a binary operator")
	}

	lvalType := n.BinOp.Lval.GetType()
	rvalType := n.BinOp.Rval.GetType()

	return (lvalType == rvalType)
}
