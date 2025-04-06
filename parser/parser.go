package parser

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

// TODO: even if we error we should still return the AST
type Parser struct {
	ER     errs.ErrorReporter
	tokens []*token.Token
	cur    int
}

func NewParser(tokens []*token.Token, ER errs.ErrorReporter) *Parser {
	return &Parser{ER, tokens, 0}
}

func (p *Parser) Parse() ast.Expr {
	expr, err := p.expression()
	if err != nil && errors.Is(err, errs.ErrParse) {
		return nil
	}
	return expr
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.equality()
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		opr := p.previous()
		rhs, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		opr := p.previous()
		rhs, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.match(token.MINUS, token.PLUS) {
		opr := p.previous()
		rhs, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(token.SLASH, token.STAR) {
		opr := p.previous()
		rhs, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		opr := p.previous()
		rhs, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{Operator: opr, Right: rhs}, nil
	}
	return p.primary()
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.FALSE) {
		return &ast.Literal{Value: false}, nil
	}
	if p.match(token.TRUE) {
		return &ast.Literal{Value: true}, nil
	}
	if p.match(token.NIL) {
		return &ast.Literal{Value: nil}, nil
	}

	if p.match(token.NUMBER, token.STRING) {
		return &ast.Literal{Value: p.previous().Literal}, nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression")
		if err != nil {
			return nil, err
		}
		return &ast.Grouping{Expression: expr}, nil
	}
	return nil, p.parseErr(p.peek(), "Expect expression")
}

func (p *Parser) consume(kind token.TokenType, msg string) (*token.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	return nil, p.parseErr(p.peek(), msg)
}

func (p *Parser) parseErr(token *token.Token, msg string) error {
	p.ER.ReportTok(token, fmt.Errorf(msg))
	return errs.ErrParse
}

func (p *Parser) previous() *token.Token {
	return p.tokens[p.cur-1]
}

func (p *Parser) match(kinds ...token.TokenType) bool {
	for _, t := range kinds {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) advance() *token.Token {
	if !p.isAtEnd() {
		p.cur++
	}
	return p.previous()
}

func (p *Parser) check(kind token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Kind == kind
}

func (p *Parser) peek() *token.Token {
	return p.tokens[p.cur]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Kind == token.EOF
}

func (p *Parser) Synchronise() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().Kind == token.SEMICOLON {
			return
		}
		switch p.peek().Kind {
		case token.CLASS, token.FUN, token.VAR, token.FOR, token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}
		p.advance()
	}
}
