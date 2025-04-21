package parser

import (
	"fmt"
	"os"
	"slices"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/token"
)

const (
	ThisOutSideClass  = "Can't use 'this' outside of a class"
	SuperOutSideClass = "Can't use 'super' outside of a class"
	ReturnTopLevel    = "Can't return from top-level code"
	SelfInheritance   = "A class can't inherit from itself"
)

type fnType int

const (
	fn_NONE fnType = iota
	fn_FUNC
	fn_INIT
	fn_METHOD
	fn_STATIC
)

type clsType int

const (
	cls_NONE clsType = iota
	cls_CLASS
	cls_SUBCLASS
)

// TODO: even if we error we should still return the AST
type Parser struct {
	tokens         []*token.Token
	cur, loopDepth int
	curClass       clsType
	curFN          fnType
	curErr         error
}

func NewParser(tokens []*token.Token) *Parser {
	return &Parser{
		tokens,
		0,
		0,
		cls_NONE,
		fn_NONE,
		nil,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, error) {
	stmts := make([]ast.Stmt, 0, 16)
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.curErr = err
		}
		stmts = append(stmts, stmt)
	}
	if p.curErr != nil {
		return nil, p.curErr
	}
	return stmts, nil
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
		if err != nil {
			p.synchronise()
			return nil, err
		}
		return val, nil
	}
	val, err := p.statement()
	if err != nil {
		p.synchronise()
		return nil, err
	}
	return val, err
}

func (p *Parser) classDeclaration() (ast.Stmt, error) {
	p.curClass = cls_CLASS
	defer func() { p.curClass = cls_NONE }()
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
		if supercls.Name.Lexeme == name.Lexeme {
			return nil, p.parseErr(supercls.Name, SelfInheritance)
		}
	}
	_, err = p.consume(token.LEFT_BRACE, "Expect '{' before class body")
	if err != nil {
		return nil, err
	}
	methods := make([]*ast.Function, 0)
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		isStatic, kind := p.match(token.STATIC), "method"
		if isStatic {
			kind = "static"
		}
		method, err := p.function(kind)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	_, err = p.consume(token.RIGHT_BRACE, "Expect '}' after class body")
	if err != nil {
		return nil, err
	}
	return &ast.Class{Name: name, Superclass: supercls, Methods: methods}, nil
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
	p.curFN = fn_FUNC
	defer func() { p.curFN = fn_NONE }()
	msg := fmt.Sprintf("Expect '(' after %s name", kind)
	_, err := p.consume(token.LEFT_PAREN, msg)
	if err != nil {
		return nil, err
	}

	params := make([]*token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			if len(params) >= 255 {
				p.parseErr(p.peek(), "Can't have more than 255 parameters")
			}
			ident, err := p.consume(token.IDENTIFIER, "Expect parameter name")
			if err != nil {
				return nil, err
			}
			params = append(params, ident)
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
	fnKind := ast.NONE
	switch kind {
	case "function":
		fnKind = ast.FUNC
	case "method":
		fnKind = ast.METHOD
	case "static":
		fnKind = ast.STATIC
	case "lambda":
		fnKind = ast.LAMDA
	}
	return &ast.Lambda{Params: params, Body: body, Kind: fnKind}, nil
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
	if p.curFN == fn_NONE {
		return nil, p.parseErr(p.previous(), ReturnTopLevel)
	}
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
	if p.match(token.EQUAL, token.PLUS_EQUAL, token.MINUS_EQUAL, token.SLASH_EQUAL, token.STAR_EQUAL) {
		opr := p.previous()
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}
		if vexpr, ok := expr.(*ast.Variable); ok {
			name := vexpr.Name
			return &ast.Assign{Name: name, Operator: opr, Value: val}, nil
		} else if get, ok := expr.(*ast.Get); ok {
			return &ast.Set{
				Object: get.Object,
				Name:   get.Name,
				Value:  val,
			}, nil
		}
		p.parseErr(opr, "Invalid assignment target")
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
		for ok := true; ok; ok = p.match(token.COMMA) {
			if len(args) >= 255 {
				p.parseErr(p.peek(), "Can't have more than 255 arguments")
			}
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
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
		if p.curClass == cls_NONE {
			return nil, p.parseErr(p.previous(), SuperOutSideClass)
		}
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
		if p.curClass == cls_NONE {
			return nil, p.parseErr(p.previous(), ThisOutSideClass)
		}
		return &ast.This{Keyword: p.previous()}, nil
	}

	if p.match(token.FUN) {
		return p.lambda("lambda")
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
	err := fmt.Errorf(msg)
	p.reportTok(token, err)
	return err
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

func (p *Parser) reportTok(tok *token.Token, msg error) {
	if tok.Kind == token.EOF {
		fmt.Fprintf(os.Stderr, "[line %d] [Parser] Error at end: %s\n", tok.Line, msg)
	} else {
		fmt.Fprintf(os.Stderr, "[line %d] [Parser] Error at '%s': %s\n", tok.Line, tok.Lexeme, msg)
	}
	p.curErr = msg
}
