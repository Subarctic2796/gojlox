package scanner

import (
	"strconv"

	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

type Scanner struct {
	src              []rune
	ER               errs.ErrorReporter
	Tokens           []*token.Token
	start, cur, Line int
}

func NewScanner(src string, er errs.ErrorReporter) *Scanner {
	return &Scanner{[]rune(src), er, make([]*token.Token, 0, 16), 0, 0, 1}
}

func (s *Scanner) ScanTokens() []*token.Token {
	for !s.isAtEnd() {
		s.start = s.cur
		s.scanToken()
	}
	s.Tokens = append(s.Tokens, &token.Token{
		Kind:    token.EOF,
		Lexeme:  "",
		Literal: nil,
		Line:    s.Line,
	})
	return s.Tokens
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case ',':
		s.addToken(token.COMMA)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '*':
		s.addToken(token.STAR)
	case '!':
		s.addMatchToken('=', token.BANG_EQUAL, token.BANG)
	case '=':
		s.addMatchToken('=', token.EQUAL_EQUAL, token.EQUAL)
	case '<':
		s.addMatchToken('=', token.LESS_EQUAL, token.LESS)
	case '>':
		s.addMatchToken('=', token.GREATER_EQUAL, token.GREATER)
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			s.multiLineComment()
		} else {
			s.addToken(token.SLASH)
		}
	case ' ', '\r', '\t':
	case '\n':
		s.Line++
	case '"':
		s.addString()
	default:
		if s.isDigit(c) {
			s.addNumber()
		} else if s.isAlpha(c) {
			s.identifier()
		} else {
			s.ER.ReportErr(s.Line, errs.ErrUnexpectedChar)
		}
	}
}

func (s *Scanner) multiLineComment() {
	nesting := 1
	for nesting > 0 && !s.isAtEnd() {
		p, pn := s.peek(), s.peekNext()
		if p == '\n' || s.src[s.cur] == '\n' {
			s.Line++
		}
		if p == '/' && pn == '*' {
			s.advance()
			s.advance()
			nesting++
			continue
		}
		if p == '*' && pn == '/' {
			s.advance()
			s.advance()
			nesting--
			continue
		}
		s.advance()
	}
	if s.isAtEnd() {
		s.ER.ReportErr(s.Line, errs.ErrUnterminatedComment)
		return
	}
	s.advance()
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}
	txt := string(s.src[s.start:s.cur])
	s.addToken(token.LookUpKeyWord(txt))
}

func (s *Scanner) isAlphaNumeric(c rune) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

func (s *Scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func (s *Scanner) addNumber() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	n, err := strconv.ParseFloat(string(s.src[s.start:s.cur]), 64)
	if err != nil {
		panic(err)
	}
	s.addTokenWithLit(token.NUMBER, n)
}

func (s *Scanner) peekNext() rune {
	if s.cur+1 >= len(s.src) {
		return 0
	}
	return s.src[s.cur+1]
}

func (s *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) addString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.Line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.ER.ReportErr(s.Line, errs.ErrUnterminatedStr)
		return
	}

	s.advance()
	val := string(s.src[s.start+1 : s.cur-1])
	s.addTokenWithLit(token.STRING, val)
}

func (s *Scanner) addMatchToken(expected rune, t1, t2 token.TokenType) {
	if s.match(expected) {
		s.addToken(t1)
	} else {
		s.addToken(t2)
	}
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}
	if s.src[s.cur] != expected {
		return false
	}
	s.cur++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}
	return s.src[s.cur]
}

func (s *Scanner) isAtEnd() bool {
	return s.cur >= len(s.src)
}

func (s *Scanner) advance() rune {
	s.cur++
	return s.src[s.cur-1]
}

func (s *Scanner) addToken(kind token.TokenType) {
	s.addTokenWithLit(kind, nil)
}

func (s *Scanner) addTokenWithLit(kind token.TokenType, lit any) {
	txt := string(s.src[s.start:s.cur])
	s.Tokens = append(s.Tokens, &token.Token{Kind: kind, Lexeme: txt, Literal: lit, Line: s.Line})
}
