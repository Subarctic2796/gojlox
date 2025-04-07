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

func (p *Parser) Parse() ([]ast.Stmt, error) {
	stmts := make([]ast.Stmt, 0, 16)
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	return stmts, nil
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.VAR) {
		val, err := p.varDeclaration()
		if errors.Is(err, errs.ErrParse) {
			p.Synchronise()
			return nil, err
		}
		return val, nil
	}
	val, err := p.statement()
	if errors.Is(err, errs.ErrParse) {
		p.Synchronise()
		return nil, err
	}
	return val, err
}

func (p *Parser) varDeclaration() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect variable name")
	if err != nil {
		return nil, err
	}
	var initializer ast.Expr
	if p.match(token.EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after variable declaration")
	if err != nil {
		return nil, err
	}
	return &ast.Var{Name: name, Initializer: initializer}, nil
}

func (p *Parser) statement() (ast.Stmt, error) {
	if p.match(token.FOR) {
		return p.forStatement()
	}
	if p.match(token.IF) {
		return p.ifStatement()
	}
	if p.match(token.PRINT) {
		return p.printStatement()
	}
	if p.match(token.WHILE) {
		return p.whileStatement()
	}
	if p.match(token.LEFT_BRACE) {
		block, err := p.block()
		if err != nil {
			return nil, err
		}
		return &ast.Block{Statements: block}, nil
	}
	return p.expressionStatement()
}

func (p *Parser) forStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'for'")
	if err != nil {
		return nil, err
	}
	var init ast.Stmt
	if p.match(token.SEMICOLON) {
		init = nil
	} else if p.match(token.VAR) {
		init, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	} else {
		init, err = p.expressionStatement()
		if err != nil {
			return nil, err
		}
	}

	var cond ast.Expr
	if !p.check(token.SEMICOLON) {
		cond, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect  ';' after loop condition")
	if err != nil {
		return nil, err
	}

	var incr ast.Expr
	if !p.check(token.RIGHT_PAREN) {
		incr, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses")
	if err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if incr != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{
				body,
				&ast.Expression{Expression: incr},
			},
		}
	}

	if cond == nil {
		cond = &ast.Literal{Value: true}
	}
	body = &ast.While{Condition: cond, Body: body}

	if init != nil {
		body = &ast.Block{
			Statements: []ast.Stmt{init, body},
		}
	}

	return body, nil
}

func (p *Parser) whileStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'while'")
	if err != nil {
		return nil, err
	}
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after condition")
	if err != nil {
		return nil, err
	}
	body, err := p.statement()
	if err != nil {
		return nil, err
	}
	return &ast.While{Condition: cond, Body: body}, nil
}

func (p *Parser) ifStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'if'")
	if err != nil {
		return nil, err
	}
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after if condition")
	if err != nil {
		return nil, err
	}
	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}
	var elseBranch ast.Stmt = nil
	if p.match(token.ELSE) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}
	return &ast.If{
		Condition:  cond,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) block() ([]ast.Stmt, error) {
	stmts := make([]ast.Stmt, 0, 8)
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	_, err := p.consume(token.RIGHT_BRACE, "Expect '}' after block")
	if err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) expressionStatement() (ast.Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after expression")
	if err != nil {
		return nil, err
	}
	return &ast.Expression{Expression: expr}, nil
}

func (p *Parser) printStatement() (ast.Stmt, error) {
	val, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after value")
	if err != nil {
		return nil, err
	}
	return &ast.Print{Expression: val}, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}
	if p.match(token.EQUAL) {
		equals := p.previous()
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}
		if vexpr, ok := expr.(*ast.Variable); ok {
			name := vexpr.Name
			return &ast.Assign{Name: name, Value: val}, nil
		}
		p.parseErr(equals, "Invalid assignment target")
	}
	return expr, nil
}

func (p *Parser) or() (ast.Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}
	for p.match(token.OR) {
		opr := p.previous()
		rhs, err := p.and()
		if err != nil {
			return nil, err
		}
		expr = &ast.Logical{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) and() (ast.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(token.AND) {
		opr := p.previous()
		rhs, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = &ast.Logical{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
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

	if p.match(token.IDENTIFIER) {
		return &ast.Variable{Name: p.previous()}, nil
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
