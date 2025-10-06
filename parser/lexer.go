// This file contains functions that do character grouping, or
// "lexing".

package parser

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/report"
	"regexp"
)

const ringSize uint = 16

type tokenTag uint

const (
	tokenError tokenTag = 0

	// imagine ASCII chars here

	// Keywords
	tokenLet = (128 + iota)
	tokenExfun
	tokenColEq
	tokenS64
	tokenU64

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

var tokenPatterns = []struct {
	tag       tokenTag
	pattern   *regexp.Regexp
	needsData bool
}{
	// Order matters

	{tokenLet, regexp.MustCompile(`^\blet\b`), false},
	{tokenColEq, regexp.MustCompile(`^:=`), false},
	{tokenExfun, regexp.MustCompile(`^\bexfun\b`), false},
	{tokenS64, regexp.MustCompile(`^\bs64\b`), false},
	{tokenU64, regexp.MustCompile(`^\bu64\b`), false},
	{tokenTag('('), regexp.MustCompile(`^\(`), false},
	{tokenTag(')'), regexp.MustCompile(`^\)`), false},

	{tokenInteger, regexp.MustCompile(`^(-?[1-9]+[0-9]*|0)`), true},

	{tokenTag('+'), regexp.MustCompile(`^\+`), false},
	{tokenTag('-'), regexp.MustCompile(`^\-`), false},

	{tokenIdent, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), true},
}
var newlinePattern = regexp.MustCompile(`^\n+`)
var blankPattern = regexp.MustCompile(`^[ \t]+`)

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

	for {
		blankFound := false
		newlineFound := false

		if match := blankPattern.FindString(p.l.data); match != "" {
			p.l.column += uint(len(match))
			p.l.data = p.l.data[len(match):]
			blankFound = true
		}
		if match := newlinePattern.FindString(p.l.data); match != "" {
			p.l.line += uint(len(match))
			p.l.column = 1
			p.l.data = p.l.data[len(match):]
			newlineFound = true
		}

		if !blankFound && !newlineFound {
			break
		}
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
		// TODO: add displaying names
		report.Report(report.Form{
			Tag:    report.ReportFatal,
			File:   p.fileName,
			Line:   token.line,
			Column: token.column,
			Msg:    fmt.Sprintf("expected token [%d]", tag),
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
