package scanner

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/Subarctic2796/gojlox/token"
)

var (
	ErrUnexpectedChar      = errors.New("unexpected character")
	ErrUnterminatedStr     = errors.New("unterminated string")
	ErrUnterminatedComment = errors.New("unterminated comment")
)

type Lexer struct {
	src              []rune
	Tokens           []token.Token
	start, cur, Line int
	curErr           error
}

func NewLexer(src string) *Lexer {
	return &Lexer{[]rune(src), make([]token.Token, 0, 16), 0, 0, 1, nil}
}

func (l *Lexer) ScanTokens() ([]token.Token, error) {
	for !l.isAtEnd() {
		l.start = l.cur
		l.scanToken()
	}
	l.Tokens = append(l.Tokens, token.NewToken(token.EOF, "", nil, l.Line))
	if l.curErr != nil {
		return nil, l.curErr
	}
	return l.Tokens, nil
}

func (l *Lexer) scanToken() {
	switch c := l.advance(); c {
	case '(':
		l.addToken(token.LPAREN)
	case ')':
		l.addToken(token.RPAREN)
	case '{':
		l.addToken(token.LBRACE)
	case '}':
		l.addToken(token.RBRACE)
	case '[':
		l.addToken(token.LSQR)
	case ']':
		l.addToken(token.RSQR)
	case ',':
		l.addToken(token.COMMA)
	case '.':
		l.addToken(token.DOT)
	case ';':
		l.addToken(token.SEMICOLON)
	case ':':
		l.addToken(token.COLON)
	case '%':
		l.addMatchToken('=', token.PERCENT_EQ, token.PERCENT)
	case '*':
		l.addMatchToken('=', token.STAR_EQ, token.STAR)
	case '+':
		l.addMatchToken('=', token.PLUS_EQ, token.PLUS)
	case '-':
		l.addMatchToken('=', token.MINUS_EQ, token.MINUS)
	case '!':
		l.addMatchToken('=', token.NEQ, token.BANG)
	case '=':
		l.addMatchToken('=', token.EQ_EQ, token.EQ)
	case '<':
		l.addMatchToken('=', token.LT_EQ, token.LT)
	case '>':
		l.addMatchToken('=', token.GT_EQ, token.GT)
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else if l.match('*') {
			l.multiLineComment()
		} else {
			l.addMatchToken('=', token.SLASH_EQ, token.SLASH)
		}
	case ' ', '\r', '\t':
	case '\n':
		l.Line++
	case '"':
		l.addString()
	default:
		if isDigit(c) {
			l.addNumber()
		} else if isAlpha(c) {
			l.identifier()
		} else {
			l.report(ErrUnexpectedChar)
		}
	}
}

func (l *Lexer) multiLineComment() {
	nesting := 1
	for nesting > 0 && !l.isAtEnd() {
		p, pn := l.peek(), l.peekNext()
		if p == '\n' || l.src[l.cur] == '\n' {
			l.Line++
		}
		if p == '/' && pn == '*' {
			l.advance()
			l.advance()
			nesting++
			continue
		}
		if p == '*' && pn == '/' {
			l.advance()
			l.advance()
			nesting--
			continue
		}
		l.advance()
	}
	if l.isAtEnd() {
		l.report(ErrUnterminatedComment)
		return
	}
	l.advance()
}

func (l *Lexer) identifier() {
	for isAlpha(l.peek()) || isDigit(l.peek()) {
		l.advance()
	}
	txt := string(l.src[l.start:l.cur])
	l.addToken(token.LookUpKeyWord(txt))
}

func isAlpha(c rune) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' }
func isDigit(c rune) bool { return c >= '0' && c <= '9' }

func (l *Lexer) addNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && isDigit(l.peekNext()) {
		// consume '.'
		l.advance()

		for isDigit(l.peek()) {
			l.advance()
		}
	}

	n, err := strconv.ParseFloat(string(l.src[l.start:l.cur]), 64)
	if err != nil {
		panic(err)
	}
	l.addTokenWithLit(token.NUMBER, n)
}

func (l *Lexer) peekNext() rune {
	if l.cur+1 >= len(l.src) {
		return 0
	}
	return l.src[l.cur+1]
}

func (l *Lexer) addString() {
	for l.peek() != '"' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.Line++
		}
		l.advance()
	}

	if l.isAtEnd() {
		l.report(ErrUnterminatedStr)
		return
	}

	l.advance()
	val := string(l.src[l.start+1 : l.cur-1])
	l.addTokenWithLit(token.STRING, val)
}

func (l *Lexer) addMatchToken(expected rune, t1, t2 token.TokenType) {
	if l.match(expected) {
		l.addToken(t1)
	} else {
		l.addToken(t2)
	}
}

func (l *Lexer) match(expected rune) bool {
	if l.isAtEnd() {
		return false
	}
	if l.src[l.cur] != expected {
		return false
	}
	l.cur++
	return true
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.src[l.cur]
}

func (l *Lexer) isAtEnd() bool { return l.cur >= len(l.src) }

func (l *Lexer) advance() rune {
	l.cur++
	return l.src[l.cur-1]
}

func (l *Lexer) addToken(kind token.TokenType) { l.addTokenWithLit(kind, nil) }

func (l *Lexer) addTokenWithLit(kind token.TokenType, lit any) {
	txt := string(l.src[l.start:l.cur])
	l.Tokens = append(l.Tokens, token.NewToken(kind, txt, lit, l.Line))
}

func (l *Lexer) report(msg error) {
	fullMsg := fmt.Sprintf("[line %d] [Lexer] Error: %s", l.Line, msg)
	if errors.Is(msg, ErrUnexpectedChar) {
		fmt.Fprintf(os.Stderr, "%s '%c'\n", fullMsg, l.src[l.cur-1])
	} else {
		fmt.Fprintln(os.Stderr, fullMsg)
	}
	l.curErr = msg
}
