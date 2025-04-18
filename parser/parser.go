package parser

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

// TODO: even if we error we should still return the AST
type Parser struct {
	ER             errs.ErrorReporter
	tokens         []*token.Token
	cur, loopDepth int
}

func NewParser(tokens []*token.Token, ER errs.ErrorReporter) *Parser {
	return &Parser{ER, tokens, 0, 0}
}

func (p *Parser) Parse() ([]ast.Stmt, []error) {
	stmts := make([]ast.Stmt, 0, 16)
	errList := make([]error, 0)
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			errList = append(errList, err)
		}
		stmts = append(stmts, stmt)
	}
	if len(errList) != 0 {
		return nil, errList
	}
	return stmts, errList
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.CLASS) {
		return p.classDeclaration()
	}
	if p.check(token.FUN) && p.checkNext(token.IDENTIFIER) {
		p.consume(token.FUN, "")
		return p.function("function")
	}
	if p.match(token.VAR) {
		val, err := p.varDeclaration()
		if errors.Is(err, errs.ErrParse) {
			p.synchronise()
			return nil, err
		}
		return val, nil
	}
	val, err := p.statement()
	if errors.Is(err, errs.ErrParse) {
		p.synchronise()
		return nil, err
	}
	return val, err
}

func (p *Parser) classDeclaration() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect class name")
	if err != nil {
		return nil, err
	}
	var supercls *ast.Variable
	if p.match(token.LESS) {
		_, err = p.consume(token.IDENTIFIER, "Expect superclass name")
		if err != nil {
			return nil, err
		}
		supercls = &ast.Variable{Name: p.previous()}
	}
	_, err = p.consume(token.LEFT_BRACE, "Expect '{' before class body")
	if err != nil {
		return nil, err
	}
	methods := make([]*ast.Function, 0)
	statics := make([]*ast.Function, 0)
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		isStatic := p.match(token.STATIC)
		method, err := p.function("method")
		if err != nil {
			return nil, err
		}
		if isStatic {
			statics = append(statics, method)
		} else {
			methods = append(methods, method)
		}
	}
	_, err = p.consume(token.RIGHT_BRACE, "Expect '}' after class body")
	if err != nil {
		return nil, err
	}
	return &ast.Class{Name: name, Superclass: supercls, Methods: methods, Statics: statics}, nil
}

func (p *Parser) function(kind string) (*ast.Function, error) {
	msg := fmt.Sprintf("Expect %s name", kind)
	name, err := p.consume(token.IDENTIFIER, msg)
	if err != nil {
		return nil, err
	}
	body, err := p.lambda(kind)
	if err != nil {
		return nil, err
	}
	return &ast.Function{Name: name, Func: body}, nil
}

func (p *Parser) lambda(kind string) (*ast.Lambda, error) {
	msg := fmt.Sprintf("Expect '(' after %s name", kind)
	_, err := p.consume(token.LEFT_PAREN, msg)
	if err != nil {
		return nil, err
	}

	params := make([]*token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(params) >= 255 {
				p.parseErr(p.peek(), "Can't have more than 255 parameters")
			}
			ident, err := p.consume(token.IDENTIFIER, "Expect parameter name")
			if err != nil {
				return nil, err
			}
			params = append(params, ident)
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after parameters")
	if err != nil {
		return nil, err
	}

	msg = fmt.Sprintf("Expect '{' before %s body", kind)
	_, err = p.consume(token.LEFT_BRACE, msg)
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &ast.Lambda{Params: params, Body: body}, nil
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
	if p.match(token.BREAK) {
		return p.breakStatement()
	}
	if p.match(token.FOR) {
		return p.forStatement()
	}
	if p.match(token.IF) {
		return p.ifStatement()
	}
	if p.match(token.PRINT) {
		return p.printStatement()
	}
	if p.match(token.RETURN) {
		return p.returnStatement()
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

func (p *Parser) breakStatement() (ast.Stmt, error) {
	if p.loopDepth == 0 {
		p.parseErr(p.previous(), "Must be in a loop to use 'break'")
	}
	_, err := p.consume(token.SEMICOLON, "Expect ';' after 'break'")
	if err != nil {
		return nil, err
	}
	return &ast.Break{}, nil
}

func (p *Parser) returnStatement() (ast.Stmt, error) {
	keyword := p.previous()
	var val ast.Expr
	var err error
	if !p.check(token.SEMICOLON) {
		val, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after return value")
	if err != nil {
		return nil, err
	}
	return &ast.Return{Keyword: keyword, Value: val}, nil
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

	p.loopDepth++
	defer func() { p.loopDepth-- }()
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

	p.loopDepth++
	defer func() { p.loopDepth-- }()
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
		} else if get, ok := expr.(*ast.Get); ok {
			return &ast.Set{
				Object: get.Object,
				Name:   get.Name,
				Value:  val,
			}, nil
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
	return p.call()
}

func (p *Parser) call() (ast.Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(token.LEFT_PAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.DOT) {
			name, err := p.consume(token.IDENTIFIER, "Expect property name after '.'")
			if err != nil {
				return nil, err
			}
			expr = &ast.Get{Object: expr, Name: name}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	args := make([]ast.Expr, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(args) >= 255 {
				p.parseErr(p.peek(), "Can't have more than 255 arguments")
			}
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	paren, err := p.consume(token.RIGHT_PAREN, "Expect ')' after arguments")
	if err != nil {
		return nil, err
	}
	return &ast.Call{Callee: callee, Paren: paren, Arguments: args}, nil
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

	if p.match(token.SUPER) {
		keyword := p.previous()
		_, err := p.consume(token.DOT, "Expect '.' after 'super'")
		if err != nil {
			return nil, err
		}
		method, err := p.consume(token.IDENTIFIER, "Expect superclass method name")
		if err != nil {
			return nil, err
		}
		return &ast.Super{Keyword: keyword, Method: method}, nil
	}

	if p.match(token.THIS) {
		return &ast.This{Keyword: p.previous()}, nil
	}

	if p.match(token.FUN) {
		return p.lambda("function")
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
	if slices.ContainsFunc(kinds, p.check) {
		p.advance()
		return true
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

func (p *Parser) checkNext(kind token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	if p.tokens[p.cur+1].Kind == token.EOF {
		return false
	}
	return p.tokens[p.cur+1].Kind == kind
}

func (p *Parser) peek() *token.Token {
	return p.tokens[p.cur]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Kind == token.EOF
}

func (p *Parser) synchronise() {
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
