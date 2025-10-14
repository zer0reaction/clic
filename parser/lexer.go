// This file contains functions that do character grouping, or
// "lexing".

package parser

import (
	"fmt"
	"lisp-go/report"
	"regexp"
)

type lexer struct {
	data     string
	line     uint
	column   uint
	writeInd uint
	readInd  uint
	rbuffer  [ringSize]token
}

const ringSize uint = 16

type tokenTag uint

const (
	tokenError tokenTag = 0

	// imagine ASCII chars here

	// Keywords
	tokenLet = (128 + iota)
	tokenExfun
	tokenIf
	tokenWhile
	tokenBinOp
	tokenType
	tokenTrue
	tokenFalse

	// Other terminals
	tokenInteger
	tokenIdent

	tokenEOF
)

type token struct {
	tag    tokenTag
	line   uint
	column uint
	data   string
}

// TODO: probably need to refactor to tokenKeyword
var tokenPatterns = []struct {
	tag       tokenTag
	pattern   *regexp.Regexp
	needsData bool
}{
	// Order matters!

	{tokenLet, regexp.MustCompile(`^\blet\b`), false},
	{tokenExfun, regexp.MustCompile(`^\bexfun\b`), false},
	{tokenIf, regexp.MustCompile(`^\bif\b`), false},
	{tokenWhile, regexp.MustCompile(`^\bwhile\b`), false},
	{tokenType, regexp.MustCompile(`^(\bs64\b|\bu64\b|\bbool\b)`), true},
	{tokenTrue, regexp.MustCompile(`^\btrue\b`), false},
	{tokenFalse, regexp.MustCompile(`^\bfalse\b`), false},

	{tokenInteger, regexp.MustCompile(`^(-?[1-9]+[0-9]*|0)`), true},

	{tokenBinOp, regexp.MustCompile(`^(:=|==|!=|<=|<|>=|>|-|\+|\*|/|%)`), true},

	{tokenTag('('), regexp.MustCompile(`^\(`), false},
	{tokenTag(')'), regexp.MustCompile(`^\)`), false},
	{tokenTag(':'), regexp.MustCompile(`^:`), false},

	{tokenIdent, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), true},
}

func (tag tokenTag) toString() string {
	if uint(tag) <= 127 {
		return fmt.Sprintf("'%c'", uint(tag))
	}

	switch tag {
	case tokenLet:
		return "'let'"

	case tokenExfun:
		return "'exfun'"

	case tokenIf:
		return "'if'"

	case tokenWhile:
		return "'while'"

	case tokenBinOp:
		return "binary operator"

	case tokenType:
		return "type"

	case tokenTrue:
		return "'true'"

	case tokenFalse:
		return "'false'"

	case tokenInteger:
		return "integer literal"

	case tokenIdent:
		return "identifier"

	case tokenEOF:
		return "EOF"

	default:
		panic("unrecognized token")
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
	if p.l.data == "" {
		panic("attempted to cache empty input")
	}

blankLoop:
	for {
		if len(p.l.data) == 0 {
			break blankLoop
		}

		switch p.l.data[0] {
		case ' ':
			p.l.column += 1

		case '\t':
			p.l.column += 1

		case '\n':
			p.l.column = 1
			p.l.line += 1

		default:
			break blankLoop
		}

		p.l.data = p.l.data[1:]
	}

	matched := false

	if p.l.data == "" {
		t := token{
			tag:    tokenEOF,
			line:   p.l.line,
			column: p.l.column,
		}
		p.pushToken(t)
		return
	}

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
		report.Report(report.Form{
			Tag:    report.ReportFatal,
			File:   p.fileName,
			Line:   p.l.line,
			Column: p.l.column,
			Msg:    "unknown syntax",
		})
	}
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
			tag.toString(), token.tag.toString())
		report.Report(report.Form{
			Tag:    report.ReportFatal,
			File:   p.fileName,
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
