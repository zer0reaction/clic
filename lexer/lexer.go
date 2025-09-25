package lexer

import (
	"fmt"
	"regexp"
)

const lexerRbufferSize uint = 16

type TokenTag uint

const (
	// 0-127 are ASCII chars
	tokenError TokenTag = (128 + iota)

	// Keywords
	TokenLet
	TokenExfun
	TokenColEq

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

type Lexer struct {
	data     string
	line     uint
	column   uint
	writeInd uint
	readInd  uint
	rbuffer  [lexerRbufferSize]Token
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
	{TokenTag('('), regexp.MustCompile(`^\(`), false},
	{TokenTag(')'), regexp.MustCompile(`^\)`), false},

	{TokenInteger, regexp.MustCompile(`^(-?[1-9]+[0-9]*|0)`), true},

	{TokenTag('+'), regexp.MustCompile(`^\+`), false},
	{TokenTag('-'), regexp.MustCompile(`^\-`), false},

	{TokenIdent, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), true},
}
var newlinePattern = regexp.MustCompile(`^\n+`)
var blankPattern = regexp.MustCompile(`^[ \t]+`)

func (t *Token) PrintInfo() {
	fmt.Printf("tag:%d line:%d column:%d data:%d\n",
		t.Tag, t.Line, t.Column, t.Data)
}

func (l *Lexer) LoadString(data string) {
	l.data = data
	l.line = 1
	l.column = 1
	l.writeInd = 0
	l.readInd = 0
}

func (l *Lexer) consumeToken() {
	if l.readInd == l.writeInd {
		panic("ring buffer underflow")
	}

	l.readInd = (l.readInd + 1) % lexerRbufferSize
}

func (l *Lexer) pushToken(t Token) {
	if (l.writeInd+1)%lexerRbufferSize == l.readInd {
		panic("ring buffer overflow")
	}

	l.rbuffer[l.writeInd] = t
	l.writeInd = (l.writeInd + 1) % lexerRbufferSize
}

func (l *Lexer) cacheToken() error {
	for {
		blankFound := false
		newlineFound := false

		if match := blankPattern.FindString(l.data); match != "" {
			l.column += uint(len(match))
			l.data = l.data[len(match):]
			blankFound = true
		}
		if match := newlinePattern.FindString(l.data); match != "" {
			l.line += uint(len(match))
			l.column = 1
			l.data = l.data[len(match):]
			newlineFound = true
		}

		if !blankFound && !newlineFound {
			break
		}
	}

	matched := false

	if l.data == "" {
		t := Token{
			Tag:    TokenEOF,
			Line:   l.line,
			Column: l.column,
		}
		l.pushToken(t)
		return nil
	}

	for _, p := range tokenPatterns {
		match := p.pattern.FindString(l.data)
		if match == "" {
			continue
		}

		matched = true

		t := Token{
			Tag:    p.tag,
			Line:   l.line,
			Column: l.column,
		}

		if p.needsData {
			t.Data = match
		}

		l.pushToken(t)

		l.data = l.data[len(match):]
		l.column += uint(len(match))
		break
	}

	if !matched {
		return fmt.Errorf(":%d:%d: error: unknown syntax",
			l.line, l.column)
	}

	return nil
}

func (l *Lexer) GetCachedCount() uint {
	if l.writeInd >= l.readInd {
		return l.writeInd - l.readInd
	} else {
		return lexerRbufferSize - l.readInd + l.writeInd
	}
}

func (l *Lexer) Peek(offset uint) (Token, error) {
	for l.GetCachedCount() <= offset {
		err := l.cacheToken()
		if err != nil {
			return Token{}, err
		}
	}

	return l.rbuffer[(l.readInd+offset)%lexerRbufferSize], nil
}

func (l *Lexer) Match(tag TokenTag) error {
	token, err := l.Peek(0)
	if err != nil {
		return err
	}

	if token.Tag != tag {
		// TODO: add displaying names
		return fmt.Errorf(":%d:%d: error: expected token [%d]",
			token.Line, token.Column, tag)
	}

	l.consumeToken()
	return nil
}

func (l *Lexer) Chop(tag TokenTag) (Token, error) {
	token, err := l.Peek(0)
	if err != nil {
		return Token{}, err
	}

	if token.Tag != tag {
		// TODO: add displaying names
		return Token{}, fmt.Errorf(":%d:%d: error: expected token [%d]",
			token.Line, token.Column, tag)
	}

	l.consumeToken()
	return token, nil
}
