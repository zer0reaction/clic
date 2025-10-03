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
	// 0-127 are ASCII chars
	tokenError tokenTag = (128 + iota)

	// Keywords
	tokenLet
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

func (p *Parser) pushToken(t token) {
	if (p.writeInd+1)%ringSize == p.readInd {
		panic("ring buffer overflow")
	}

	p.rbuffer[p.writeInd] = t
	p.writeInd = (p.writeInd + 1) % ringSize
}

func (p *Parser) cacheToken() {
	if p.data == "" {
		panic("attempted to cache empty input")
	}

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
		t := token{
			tag:    tokenEOF,
			line:   p.line,
			column: p.column,
		}
		p.pushToken(t)
		return
	}

	for _, pattern := range tokenPatterns {
		match := pattern.pattern.FindString(p.data)
		if match == "" {
			continue
		}

		matched = true

		t := token{
			tag:    pattern.tag,
			line:   p.line,
			column: p.column,
		}

		if pattern.needsData {
			t.data = match
		}

		p.pushToken(t)

		p.data = p.data[len(match):]
		p.column += uint(len(match))
		break
	}

	if !matched {
		report.Report(report.Form{
			Tag:    report.ReportFatal,
			File:   p.fileName,
			Line:   p.line,
			Column: p.column,
			Msg:    "unknown syntax",
		})
	}
}

func (p *Parser) peek(offset uint) token {
	for p.getCachedCount() <= offset {
		p.cacheToken()
	}
	return p.rbuffer[(p.readInd+offset)%ringSize]
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
