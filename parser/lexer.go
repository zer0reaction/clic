package parser

import (
	"fmt"
	"regexp"
)

const ringSize uint = 16

type TokenTag uint

const (
	// 0-127 are ASCII chars
	tokenError TokenTag = (128 + iota)

	// Keywords
	TokenLet
	TokenExfun
	TokenColEq
	TokenS64
	TokenU64

	// Other terminals
	TokenInteger
	TokenIdent

	TokenEOF
)

type Token struct {
	Tag    TokenTag
	Line   uint
	Column uint
	Data   string
}

var tokenPatterns = []struct {
	tag       TokenTag
	pattern   *regexp.Regexp
	needsData bool
}{
	// Order matters

	{TokenLet, regexp.MustCompile(`^\blet\b`), false},
	{TokenColEq, regexp.MustCompile(`^:=`), false},
	{TokenExfun, regexp.MustCompile(`^\bexfun\b`), false},
	{TokenS64, regexp.MustCompile(`^\bs64\b`), false},
	{TokenU64, regexp.MustCompile(`^\bu64\b`), false},
	{TokenTag('('), regexp.MustCompile(`^\(`), false},
	{TokenTag(')'), regexp.MustCompile(`^\)`), false},

	{TokenInteger, regexp.MustCompile(`^(-?[1-9]+[0-9]*|0)`), true},

	{TokenTag('+'), regexp.MustCompile(`^\+`), false},
	{TokenTag('-'), regexp.MustCompile(`^\-`), false},

	{TokenIdent, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), true},
}
var newlinePattern = regexp.MustCompile(`^\n+`)
var blankPattern = regexp.MustCompile(`^[ \t]+`)

func (p *Parser) getCachedCount() uint {
	if p.writeInd >= p.readInd {
		return p.writeInd - p.readInd
	} else {
		return ringSize - p.readInd + p.writeInd
	}
}

func (p *Parser) consumeToken() {
	if p.readInd == p.writeInd {
		panic("ring buffer underflow")
	}

	p.readInd = (p.readInd + 1) % ringSize
}

func (p *Parser) pushToken(t Token) {
	if (p.writeInd+1)%ringSize == p.readInd {
		panic("ring buffer overflow")
	}

	p.rbuffer[p.writeInd] = t
	p.writeInd = (p.writeInd + 1) % ringSize
}

func (p *Parser) cacheToken() error {
	for {
		blankFound := false
		newlineFound := false

		if match := blankPattern.FindString(p.data); match != "" {
			p.column += uint(len(match))
			p.data = p.data[len(match):]
			blankFound = true
		}
		if match := newlinePattern.FindString(p.data); match != "" {
			p.line += uint(len(match))
			p.column = 1
			p.data = p.data[len(match):]
			newlineFound = true
		}

		if !blankFound && !newlineFound {
			break
		}
	}

	matched := false

	if p.data == "" {
		t := Token{
			Tag:    TokenEOF,
			Line:   p.line,
			Column: p.column,
		}
		p.pushToken(t)
		return nil
	}

	for _, pattern := range tokenPatterns {
		match := pattern.pattern.FindString(p.data)
		if match == "" {
			continue
		}

		matched = true

		t := Token{
			Tag:    pattern.tag,
			Line:   p.line,
			Column: p.column,
		}

		if pattern.needsData {
			t.Data = match
		}

		p.pushToken(t)

		p.data = p.data[len(match):]
		p.column += uint(len(match))
		break
	}

	if !matched {
		return fmt.Errorf(":%d:%d: error: unknown syntax",
			p.line, p.column)
	}

	return nil
}

func (p *Parser) peek(offset uint) (Token, error) {
	for p.getCachedCount() <= offset {
		err := p.cacheToken()
		if err != nil {
			return Token{}, err
		}
	}

	return p.rbuffer[(p.readInd+offset)%ringSize], nil
}

func (p *Parser) match(tag TokenTag) (Token, error) {
	token, err := p.peek(0)
	if err != nil {
		return Token{}, err
	}

	if token.Tag != tag {
		// TODO: add displaying names
		return Token{}, fmt.Errorf(":%d:%d: error: expected token [%d]",
			token.Line, token.Column, tag)
	}

	p.consumeToken()
	return token, nil
}

func (p *Parser) consume() (Token, error) {
	token, err := p.peek(0)
	if err != nil {
		return Token{}, err
	}
	p.consumeToken()
	return token, nil
}

func (p *Parser) discard() {
	p.consumeToken()
}
