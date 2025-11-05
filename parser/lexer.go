// This file contains functions that do character grouping, or
// "lexing".

package parser

import (
	"clic/report"
	"fmt"
	"regexp"
)

type lexer struct {
	data   string
	line   uint
	column uint

	writeInd uint
	readInd  uint
	rbuffer  [ringSize]token
}

const ringSize uint = 16

// TODO: Apply some DOD principles in here. Token should only contain
// its tag and index of char (u32) in file. Node should contain the
// index of the major token. Lexer could be reset back to the start
// and calulate line and column of token or node. Minimize the node
// struct, make separate arrays for stuff and keep indexes to them.
type token struct {
	tag    tokenTag
	line   uint
	column uint
	data   string
}

type tokenTag uint

// To start iota from 0 later
const tokenError tokenTag = 0

const (
	// Imagine ASCII chars here

	// Keywords
	tokenKeyword tokenTag = (128 + iota)
	tokenBinOp
	tokenType

	// Other terminals
	tokenInt
	tokenIdent

	tokenEOF
)

var tokenPatterns = []struct {
	tag       tokenTag
	pattern   *regexp.Regexp
	needsData bool
}{
	// Order matters!

	{tokenKeyword, regexp.MustCompile(`^(\blet\b|\bdefun\b|\bexfun\b|\breturn\b|\bif\b|\belse\b|\bwhile\b|\btrue\b|\bfalse\b|\bauto\b|\btypedef\b|\bfor\b)`), true},
	{tokenType, regexp.MustCompile(`^(\bvoid\b|\bs64\b|\bu64\b|\bbool\b|\bstruct\b)`), true},
	{tokenInt, regexp.MustCompile(`^(-?[1-9]+[0-9]*|0)`), true},
	{tokenBinOp, regexp.MustCompile(`^(:=|==|!=|<=|<|>=|>|-|\+|\*|/|%)`), true},

	{tokenTag('('), regexp.MustCompile(`^\(`), false},
	{tokenTag(')'), regexp.MustCompile(`^\)`), false},
	{tokenTag(':'), regexp.MustCompile(`^:`), false},

	{tokenIdent, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), true},
}

func (tag tokenTag) stringify() string {
	if uint(tag) <= 127 {
		return fmt.Sprintf("'%c'", uint(tag))
	}

	switch tag {
	case tokenKeyword:
		return "keyword"

	case tokenBinOp:
		return "binary operator"

	case tokenType:
		return "type"

	case tokenInt:
		return "integer literal"

	case tokenIdent:
		return "identifier"

	case tokenEOF:
		return "EOF"

	default:
		panic("not implemented")
	}
}

func (p *Parser) getCachedCount() uint {
	if p.l.writeInd >= p.l.readInd {
		return p.l.writeInd - p.l.readInd
	} else {
		return ringSize - p.l.readInd + p.l.writeInd
	}
}

func (p *Parser) consumeToken() {
	if p.l.readInd == p.l.writeInd {
		panic("ring buffer underflow")
	}

	p.l.readInd = (p.l.readInd + 1) % ringSize
}

func (p *Parser) pushToken(t token) {
	if (p.l.writeInd+1)%ringSize == p.l.readInd {
		panic("ring buffer overflow")
	}

	p.l.rbuffer[p.l.writeInd] = t
	p.l.writeInd = (p.l.writeInd + 1) % ringSize
}

func (p *Parser) cacheToken() {
	p.skipBlanksAndComments()

	// Multiple EOFs should be emmited when trying to peek beyond the
	// end of file
	if len(p.l.data) == 0 {
		t := token{
			tag:    tokenEOF,
			line:   p.l.line,
			column: p.l.column,
		}
		p.pushToken(t)
		return
	}

	matched := false

	for _, pattern := range tokenPatterns {
		match := pattern.pattern.FindString(p.l.data)
		if match == "" {
			continue
		}

		matched = true

		t := token{
			tag:    pattern.tag,
			line:   p.l.line,
			column: p.l.column,
		}

		if pattern.needsData {
			t.data = match
		}

		p.pushToken(t)

		p.l.data = p.l.data[len(match):]
		p.l.column += uint(len(match))
		break
	}

	if !matched {
		p.r.Report(report.Form{
			Tag:    report.ReportFatal,
			Line:   p.l.line,
			Column: p.l.column,
			Msg:    "unknown syntax",
		})
	}
}

func (p *Parser) skipBlanksAndComments() {
	// Regular expressions are not used here because we need to
	// check for newlines.

	const (
		init = iota
		comment
		// 'goto endSkipping' instead of final state
	)

	state := init

	for len(p.l.data) > 0 {
		switch state {
		case init:
			switch p.l.data[0] {
			case ' ':
			case '\t':
			case '\n':
				// Stay in init
			case ';':
				state = comment
			default:
				goto endSkipping
			}

		case comment:
			switch p.l.data[0] {
			case '\n':
				// To check for blanks after comment
				state = init
			default:
				// Stay in comment
			}

		default:
			panic("unknown state")
		}

		// state != final here, so we need to consume
		// the char.

		if p.l.data[0] == '\n' {
			p.l.column = 1
			p.l.line++
		} else {
			p.l.column++
		}
		p.l.data = p.l.data[1:]
	}

endSkipping:
}

func (p *Parser) peek(offset uint) token {
	for p.getCachedCount() <= offset {
		p.cacheToken()
	}
	return p.l.rbuffer[(p.l.readInd+offset)%ringSize]
}

func (p *Parser) match(tag tokenTag) token {
	token := p.peek(0)

	if token.tag != tag {
		msg := fmt.Sprintf("expected %s, got %s",
			tag.stringify(), token.tag.stringify())
		p.r.Report(report.Form{
			Tag:    report.ReportFatal,
			Line:   token.line,
			Column: token.column,
			Msg:    msg,
		})
	}

	p.consumeToken()
	return token
}

func (p *Parser) consume() token {
	token := p.peek(0)
	p.consumeToken()
	return token
}

func (p *Parser) discard() {
	p.consumeToken()
}
