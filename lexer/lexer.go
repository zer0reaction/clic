package lexer

import (
	"fmt"
	"errors"
	"regexp"
	"github.com/zer0reaction/lisp-go/symbol"
)

const lexerRbufferSize uint = 16

type TokenType uint

const (
	TokenRbrOpen TokenType = iota
	TokenRbrClose
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
}{
	{TokenRbrOpen, regexp.MustCompile(`^\(`)},
	{TokenRbrClose, regexp.MustCompile(`^\)`)},
}
var newlinePattern = regexp.MustCompile(`^\n+`)
var blankPattern = regexp.MustCompile(`^[ \t]+`)

func (t *Token) PrintInfo() {
	fmt.Printf("type:%d line:%d col:%d id:%d\n",
			t.Type, t.Line, t.Column, t.TableId)
}

func (l *Lexer) LoadString(data string) {
	l.data = data
	l.line = 1
	l.column = 1
	l.writeInd = 0
	l.readInd = 0
}

func (l *Lexer) readToken() (*Token, error) {
	if l.readInd == l.writeInd {
		return nil, errors.New("readToken: ring buffer overflow")
	}

	t := &l.rbuffer[l.readInd]
	l.readInd = (l.readInd + 1) % lexerRbufferSize

	return t, nil
}

func (l *Lexer) writeToken(t Token) error {
	if (l.writeInd+1)%lexerRbufferSize == l.readInd {
		return errors.New("writeToken: ring buffer underflow")
	}

	l.rbuffer[l.writeInd] = t
	l.writeInd = (l.writeInd + 1) % lexerRbufferSize

	return nil
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
			TableId: symbol.SymbolIdNone,
		}

		err := l.writeToken(t)
		if err != nil {
			return err
		}

		l.data = l.data[len(match):]
		l.column += uint(len(match))
		break
	}

	if !matched {
		return fmt.Errorf("cacheToken: no tokens matched at line %d, column %d",
					l.line, l.column)
	}


	return nil
}

func (l *Lexer) DebugCacheToken() error {
	return l.cacheToken()
}

func (l *Lexer) DebugReadToken() (*Token, error) {
	return l.readToken()
}

func (l *Lexer) GetCachedCount() uint {
	if l.writeInd >= l.readInd {
		return l.writeInd - l.readInd;
	} else {
		return lexerRbufferSize - l.readInd + l.writeInd;
	}
}
