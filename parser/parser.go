package parser

import (
	"fmt"
	"os"
	"slices"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/token"
)

const (
	ReturnTopLevel = "Can't return from top-level code"
	ReturnFromInit = "Can't return a value from an initializer"

	InheritsSelf       = "A class can't inherit from itself"
	ThisNotInClass     = "Can't use 'this' outside of a class"
	ThisInStatic       = "Can't use 'this' in a static function"
	InitIsStatic       = "Can't use 'init' as a static function"
	StaticNotInClass   = "Can't use 'static' outside of a class"
	StaticNeedsMethod  = "'static' must be before a class method"
	SuperInStatic      = "Can't use 'super' in a static method"
	SuperNotInClass    = "Can't use 'super' outside of a class"
	SuperNotInSubClass = "Can't use 'super' in a class with no superclass"

	// not used yet
	// AlreadyInScope       = "Already a variable with this name in this scope"
	// LocalInitializesSelf = "Can't read local variable in its own initializer"
	// LocalNotRead         = "Local variable is not used"
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
	curFN          ast.FnType
	curErr         error
}

func NewParser(tokens []*token.Token) *Parser {
	return &Parser{
		tokens,
		0,
		0,
		cls_NONE,
		ast.FN_NONE,
		nil,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, error) {
	stmts := make([]ast.Stmt, 0, 16)
	for !p.isAtEnd() {
		// don't need to check the error
		// as parseErr reports it and sets curErr
		stmt, _ := p.declaration()
		stmts = append(stmts, stmt)
	}
	// need to check here because we don't want to give
	// the rest of the pipeline a bad input
	if p.curErr != nil {
		return nil, p.curErr
	}
	return stmts, nil
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.CLASS) {
		return p.classDeclaration()
	}
	// check for lambdas
	if p.check(token.FUN) && p.checkNext(token.IDENTIFIER) {
		_, err := p.consume(token.FUN, "")
		if err != nil {
			return nil, err
		}
		return p.function(ast.FN_FUNC)
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
	prvCLS := p.curClass
	p.curClass = cls_CLASS
	defer func() { p.curClass = prvCLS }()
	name, err := p.consume(token.IDENTIFIER, "Expect class name")
	if err != nil {
		return nil, err
	}
	var supercls *ast.Variable
	if p.match(token.LT) {
		_, err = p.consume(token.IDENTIFIER, "Expect superclass name")
		if err != nil {
			return nil, err
		}
		supercls = &ast.Variable{Name: p.previous()}
		if supercls.Name.Lexeme == name.Lexeme {
			// only report error, this way we don't mess up the state of the parser
			// it also makes parser errors much less noisy
			_ = p.parseErr(supercls.Name, InheritsSelf)
		}
		p.curClass = cls_SUBCLASS
	}
	_, err = p.consume(token.LBRACE, "Expect '{' before class body")
	if err != nil {
		return nil, err
	}
	methods := make([]*ast.Function, 0)
	for !p.check(token.RBRACE) && !p.isAtEnd() {
		isStatic, kind := p.match(token.STATIC), ast.FN_METHOD
		if isStatic {
			kind = ast.FN_STATIC
		}
		method, err := p.function(kind)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}
	_, err = p.consume(token.RBRACE, "Expect '}' after class body")
	if err != nil {
		return nil, err
	}
	return &ast.Class{Name: name, Superclass: supercls, Methods: methods}, nil
}

func (p *Parser) function(kind ast.FnType) (*ast.Function, error) {
	msg := fmt.Sprintf("Expect %s name", kind)
	name, err := p.consume(token.IDENTIFIER, msg)
	if err != nil {
		return nil, err
	}
	if name.Lexeme == "init" {
		if kind == ast.FN_STATIC {
			// only report error, this way we don't mess up the state of the parser
			// it also makes parser errors much less noisy
			_ = p.parseErr(name, InitIsStatic)
		}
		kind = ast.FN_INIT
	}
	body, err := p.lambda(kind)
	if err != nil {
		return nil, err
	}
	fn := &ast.Function{
		Name:   name,
		Params: body.Func.Params,
		Body:   body.Func.Body,
		Kind:   body.Func.Kind,
	}
	return fn, nil
}

func (p *Parser) lambda(kind ast.FnType) (*ast.Lambda, error) {
	prvFn := p.curFN
	p.curFN = kind
	defer func() { p.curFN = prvFn }()
	msg := fmt.Sprintf("Expect '(' after %s name", kind)
	_, err := p.consume(token.LPAREN, msg)
	if err != nil {
		return nil, err
	}

	params := make([]*token.Token, 0)
	if !p.check(token.RPAREN) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			if len(params) >= 255 {
				// only report error, this way we don't mess up the state of the parser
				// it also makes parser errors much less noisy
				_ = p.parseErr(p.peek(), "Can't have more than 255 parameters")
			}
			ident, err := p.consume(token.IDENTIFIER, "Expect parameter name")
			if err != nil {
				return nil, err
			}
			params = append(params, ident)
		}
	}

	_, err = p.consume(token.RPAREN, "Expect ')' after parameters")
	if err != nil {
		return nil, err
	}

	msg = fmt.Sprintf("Expect '{' before %s body", kind)
	_, err = p.consume(token.LBRACE, msg)
	if err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}
	fn := &ast.Function{Name: nil, Params: params, Body: body, Kind: kind}
	return &ast.Lambda{Func: fn}, nil
}

func (p *Parser) varDeclaration() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect variable name")
	if err != nil {
		return nil, err
	}
	var initializer ast.Expr
	if p.match(token.EQ) {
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
	} else if p.match(token.FOR) {
		return p.forStatement()
	} else if p.match(token.IF) {
		return p.ifStatement()
	} else if p.match(token.PRINT) {
		return p.printStatement()
	} else if p.match(token.RETURN) {
		return p.returnStatement()
	} else if p.match(token.WHILE) {
		return p.whileStatement()
	} else if p.match(token.LBRACE) {
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
		// only report error, this way we don't mess up the state of the parser
		// it also makes parser errors much less noisy
		_ = p.parseErr(p.previous(), "Must be in a loop to use 'break'")
	}
	keyword := p.previous()
	_, err := p.consume(token.SEMICOLON, "Expect ';' after 'break'")
	if err != nil {
		return nil, err
	}
	return &ast.Control{Keyword: keyword, Value: nil}, nil
}

func (p *Parser) returnStatement() (ast.Stmt, error) {
	keyword := p.previous()
	// only report error, this way we don't mess up the state of the parser
	// it also makes parser errors much less noisy
	switch p.curFN {
	case ast.FN_NONE:
		_ = p.parseErr(keyword, ReturnTopLevel)
	case ast.FN_INIT:
		_ = p.parseErr(keyword, ReturnFromInit)
	}
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
	return &ast.Control{Keyword: keyword, Value: val}, nil
}

func (p *Parser) forStatement() (ast.Stmt, error) {
	_, err := p.consume(token.LPAREN, "Expect '(' after 'for'")
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
	if !p.check(token.RPAREN) {
		incr, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.RPAREN, "Expect ')' after for clauses")
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
	_, err := p.consume(token.LPAREN, "Expect '(' after 'while'")
	if err != nil {
		return nil, err
	}

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RPAREN, "Expect ')' after condition")
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
	_, err := p.consume(token.LPAREN, "Expect '(' after 'if'")
	if err != nil {
		return nil, err
	}
	cond, err := p.expression()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(token.RPAREN, "Expect ')' after if condition")
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
	for !p.check(token.RBRACE) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	_, err := p.consume(token.RBRACE, "Expect '}' after block")
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
	if p.check(token.STATIC) {
		switch p.curClass {
		case cls_NONE:
			return nil, p.parseErr(p.peek(), StaticNotInClass)
		default:
			return nil, p.parseErr(p.peek(), StaticNeedsMethod)
		}
	}
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}
	if p.match(token.EQ, token.PLUS_EQ, token.MINUS_EQ, token.SLASH_EQ, token.STAR_EQ, token.PERCENT_EQ) {
		opr := p.previous()
		val, err := p.assignment()
		if err != nil {
			return nil, err
		}
		switch n := expr.(type) {
		case *ast.Variable:
			return &ast.Assign{Name: n.Name, Operator: opr, Value: val}, nil
		case *ast.Get:
			return &ast.Set{
				Object: n.Object,
				Name:   n.Name,
				Value:  p.desugarOprEQ(n, opr, val),
			}, nil
		case *ast.IndexedGet:
			if n.Stop != nil {
				_ = p.parseErr(n.Sqr, "Can't use slicing to set values")
			}
			return &ast.IndexedSet{
				Object: n.Object,
				Sqr:    n.Sqr,
				Index:  n.Start,
				Value:  p.desugarOprEQ(n, opr, val),
			}, nil
		}
		// only report error, this way we don't mess up the state of the parser
		// it also makes parser errors much less noisy
		_ = p.parseErr(opr, "Invalid assignment target")
	}
	return expr, nil
}

func (p *Parser) desugarOprEQ(get ast.Expr, opr *token.Token, val ast.Expr) ast.Expr {
	// inst.a += 23;
	// { inst.a = [ (inst.a) + 23 ] }
	// ^Set       ^Bin     ^Get
	oprType := token.NONE
	switch opr.Kind {
	case token.EQ:
		return val
	case token.PLUS_EQ:
		oprType = token.PLUS
	case token.MINUS_EQ:
		oprType = token.MINUS
	case token.SLASH_EQ:
		oprType = token.SLASH
	case token.STAR_EQ:
		oprType = token.STAR
	case token.PERCENT_EQ:
		oprType = token.PERCENT
	}
	// [ (inst.a) + 23 ]
	// ^Bin     ^Get
	return &ast.Binary{
		Left:     get,
		Operator: token.NewToken(oprType, opr.Lexeme, nil, opr.Line),
		Right:    val,
	}
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

	for p.match(token.NEQ, token.EQ_EQ) {
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
	expr, err := p.addition()
	if err != nil {
		return nil, err
	}

	for p.match(token.GT, token.GT_EQ, token.LT, token.LT_EQ) {
		opr := p.previous()
		rhs, err := p.addition()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) addition() (ast.Expr, error) {
	expr, err := p.multiplication()
	if err != nil {
		return nil, err
	}
	for p.match(token.MINUS, token.PLUS) {
		opr := p.previous()
		rhs, err := p.multiplication()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: opr, Right: rhs}
	}
	return expr, nil
}

func (p *Parser) multiplication() (ast.Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(token.SLASH, token.STAR, token.PERCENT) {
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
		if p.match(token.LPAREN) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.LSQR) {
			expr, err = p.finishIndex(expr)
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
	if !p.check(token.RPAREN) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			if len(args) >= 255 {
				// only report error, this way we don't mess up the state of the parser
				// it also makes parser errors much less noisy
				_ = p.parseErr(p.peek(), "Can't have more than 255 arguments")
			}
			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}
	paren, err := p.consume(token.RPAREN, "Expect ')' after arguments")
	if err != nil {
		return nil, err
	}
	return &ast.Call{Callee: callee, Paren: paren, Arguments: args}, nil
}

func (p *Parser) primary() (ast.Expr, error) {
	if p.match(token.FALSE) {
		return &ast.Literal{Value: false}, nil
	} else if p.match(token.TRUE) {
		return &ast.Literal{Value: true}, nil
	} else if p.match(token.NIL) {
		return &ast.Literal{Value: nil}, nil
	} else if p.match(token.NUMBER, token.STRING) {
		return &ast.Literal{Value: p.previous().Literal}, nil
	} else if p.match(token.SUPER) {
		keyword := p.previous()
		// only report error, this way we don't mess up the state of the parser
		// it also makes parser errors much less noisy
		if p.curClass == cls_NONE {
			_ = p.parseErr(keyword, SuperNotInClass)
		} else if p.curClass != cls_SUBCLASS {
			_ = p.parseErr(keyword, SuperNotInSubClass)
		}
		if p.curFN == ast.FN_STATIC {
			_ = p.parseErr(keyword, SuperInStatic)
		}
		_, err := p.consume(token.DOT, "Expect '.' after 'super'")
		if err != nil {
			return nil, err
		}
		method, err := p.consume(token.IDENTIFIER, "Expect superclass method name")
		if err != nil {
			return nil, err
		}
		return &ast.Super{Keyword: keyword, Method: method}, nil
	} else if p.match(token.THIS) {
		keyword := p.previous()
		// only report error, this way we don't mess up the state of the parser
		// it also makes parser errors much less noisy
		if p.curClass == cls_NONE {
			_ = p.parseErr(keyword, ThisNotInClass)
		}
		if p.curFN == ast.FN_STATIC {
			_ = p.parseErr(keyword, ThisInStatic)
		}
		return &ast.This{Keyword: keyword}, nil
	} else if p.match(token.FUN) {
		return p.lambda(ast.FN_LAMBDA)
	} else if p.match(token.IDENTIFIER) {
		return &ast.Variable{Name: p.previous()}, nil
	} else if p.match(token.LPAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(token.RPAREN, "Expect ')' after expression")
		if err != nil {
			return nil, err
		}
		return &ast.Grouping{Expression: expr}, nil
	} else if p.match(token.LSQR) {
		sqr := p.previous()
		elements, err := p.finishArray()
		if err != nil {
			return nil, err
		}
		return &ast.ArrayLiteral{Sqr: sqr, Elements: elements}, nil
	} else if p.match(token.LBRACE) {
		brace := p.previous()
		pairs, err := p.finishHashMap()
		if err != nil {
			return nil, err
		}
		return &ast.HashLiteral{Brace: brace, Pairs: pairs}, nil
	}
	return nil, p.parseErr(p.peek(), "Expect expression")
}

func (p *Parser) finishIndex(iter ast.Expr) (ast.Expr, error) {
	if p.match(token.RSQR) {
		return nil, p.parseErr(p.previous(), "Expect an expression or ':' in an index expression")
	}
	var sqr *token.Token = nil
	var colon *token.Token = nil
	var err error
	var startIdx ast.Expr = nil
	if !p.check(token.COLON) {
		// arr[s:?]
		startIdx, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	// arr[s?:?]
	var stopIdx ast.Expr = nil
	if p.match(token.COLON) {
		colon = p.previous()
		if !p.check(token.RSQR) {
			stopIdx, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
		sqr, err = p.consume(token.RSQR, "Expect ']' after index")
		if err != nil {
			return nil, err
		}
	}
	if sqr == nil {
		sqr, err = p.consume(token.RSQR, "Expect ']' after index")
		if err != nil {
			return nil, err
		}
	}
	return &ast.IndexedGet{
		Object: iter,
		Sqr:    sqr,
		Start:  startIdx,
		Colon:  colon,
		Stop:   stopIdx,
	}, nil
}

func (p *Parser) finishArray() ([]ast.Expr, error) {
	elements := make([]ast.Expr, 0)
	if !p.check(token.RSQR) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			// found trailing comma
			if p.check(token.RSQR) {
				break
			}
			elm, err := p.expression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elm)
		}
	}
	_, err := p.consume(token.RSQR, "Expect ']' after array elements")
	if err != nil {
		return nil, err
	}
	return elements, nil
}

func (p *Parser) finishHashMap() (map[ast.Expr]ast.Expr, error) {
	pairs := make(map[ast.Expr]ast.Expr)
	if !p.check(token.RBRACE) {
		for ok := true; ok; ok = p.match(token.COMMA) {
			// found trailing comma
			if p.check(token.RBRACE) {
				break
			}
			key, err := p.expression()
			if err != nil {
				return nil, err
			}
			_, err = p.consume(token.COLON, "Expect ':' after hashmap key")
			if err != nil {
				return nil, err
			}
			val, err := p.expression()
			if err != nil {
				return nil, err
			}
			pairs[key] = val
		}
	}
	_, err := p.consume(token.RBRACE, "Expect '}' after array elements")
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (p *Parser) consume(kind token.TokenType, msg string) (*token.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	return nil, p.parseErr(p.peek(), msg)
}

func (p *Parser) parseErr(tok *token.Token, msg string) error {
	err := fmt.Errorf("%s", msg)
	if tok.Kind == token.EOF {
		fmt.Fprintf(os.Stderr, "[line %d] [Parser] Error at end: %s\n", tok.Line, err)
	} else {
		fmt.Fprintf(os.Stderr, "[line %d] [Parser] Error at '%s': %s\n", tok.Line, tok.Lexeme, err)
	}
	p.curErr = err
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
		case token.CLASS, token.FUN, token.VAR, token.FOR, token.IF, token.WHILE, token.PRINT, token.RETURN, token.BREAK, token.STATIC:
			return
		}
		p.advance()
	}
}
