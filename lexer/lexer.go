package lexer

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/symbol"
	"regexp"
)

const lexerRbufferSize uint = 16

type TokenType uint

const (
	tokenError TokenType = iota
	TokenRbrOpen
	TokenRbrClose
	TokenPlus
	TokenInteger
	TokenEOF
	tokenCount
)

type Token struct {
	Type    TokenType
	Line    uint
	Column  uint
	TableId int
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
	tokenType TokenType
	pattern   *regexp.Regexp
	needsData bool
}{
	{TokenRbrOpen, regexp.MustCompile(`^\(`), false},
	{TokenRbrClose, regexp.MustCompile(`^\)`), false},
	{TokenPlus, regexp.MustCompile(`^\+`), false},
	{TokenInteger, regexp.MustCompile(`^-?[1-9]+[0-9]*`), true},
}
var newlinePattern = regexp.MustCompile(`^\n+`)
var blankPattern = regexp.MustCompile(`^[ \t]+`)

func (t *Token) PrintInfo() {
	fmt.Printf("type:%d line:%d column:%d id:%d\n",
		t.Type, t.Line, t.Column, t.TableId)
}

func (l *Lexer) LoadString(data string) {
	l.data = data
	l.line = 1
	l.column = 1
	l.writeInd = 0
	l.readInd = 0
}

func (l *Lexer) popToken() *Token {
	if l.readInd == l.writeInd {
		panic("ring buffer underflow")
	}

	t := &l.rbuffer[l.readInd]
	l.readInd = (l.readInd + 1) % lexerRbufferSize

	return t
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
			Type:    TokenEOF,
			Line:    l.line,
			Column:  l.column,
			TableId: symbol.IdNone,
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
			Type:    p.tokenType,
			Line:    l.line,
			Column:  l.column,
			TableId: symbol.IdNone,
		}

		if p.needsData {
			id := symbol.Create()
			symbol.SetData(id, match)
			t.TableId = id
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

func (l *Lexer) PeekToken(offset uint) (*Token, error) {
	for l.GetCachedCount() <= offset {
		err := l.cacheToken()
		if err != nil {
			return nil, err
		}
	}

	return &l.rbuffer[(l.readInd+offset)%lexerRbufferSize], nil
}

func (l *Lexer) Match(tokenType TokenType) error {
	token, err := l.PeekToken(0)
	if err != nil {
		return err
	}

	if token.Type != tokenType {
		// TODO: add displaying names
		return fmt.Errorf(":%d:%d: error: expected token [%d]",
			token.Line, token.Column, tokenType)
	}

	l.consumeToken()
	return nil
}
