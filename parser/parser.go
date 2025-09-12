package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/lexer"
	"github.com/zer0reaction/lisp-go/symbol"
)

type NodeType uint

const (
	NodeBinOpSum NodeType = iota
	NodeInteger
)

type Node struct {
	Type    NodeType
	TableId int

	BinOpLval *Node
	BinOpRval *Node
}

func List(l *lexer.Lexer) (*Node, error) {
	lookahead, err := l.PeekToken(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Type {
	case lexer.TokenRbrOpen:
		_, err := l.Match(lexer.TokenRbrOpen)
		if err != nil {
			return nil, err
		}

		n, err := expr(l)
		if err != nil {
			return nil, err
		}

		_, err = l.Match(lexer.TokenRbrClose)
		if err != nil {
			return nil, err
		}

		return n, nil
	default:
		return nil, fmt.Errorf(":%d:%d: expected list",
			lookahead.Line, lookahead.Column)
	}
}

func expr(l *lexer.Lexer) (*Node, error) {
	lookahead, err := l.PeekToken(0)
	if err != nil {
		return nil, err
	}

	switch lookahead.Type {
	case lexer.TokenPlus:
		_, err := l.Match(lexer.TokenPlus)
		if err != nil {
			return nil, err
		}

		n := Node{
			Type:    NodeBinOpSum,
			TableId: symbol.IdNone,
		}

		n.BinOpLval, err = expr(l)
		if err != nil {
			return nil, err
		}
		n.BinOpRval, err = expr(l)
		if err != nil {
			return nil, err
		}

		return &n, nil
	case lexer.TokenInteger:
		t, err := l.Match(lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		n := Node{
			Type:    NodeInteger,
			TableId: t.TableId,
		}

		return &n, nil
	case lexer.TokenRbrOpen:
		return List(l)
	default:
		return nil, fmt.Errorf(":%d:%d: expected expression",
			lookahead.Line, lookahead.Column)
	}
}
