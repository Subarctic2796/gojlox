package resolver

import (
	"errors"
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/interpreter"
	"github.com/Subarctic2796/gojlox/token"
)

var (
	ErrAlreadyInScope       = errors.New("Already a variable with this name in this scope")
	ErrLocalInitializesSelf = errors.New("Can't read local variable in its own initializer")
	ErrLocalNotRead         = errors.New("Local variable is not used")
)

type varInfo struct {
	name   *token.Token
	status varStatus
}

type varStatus int

const (
	vs_DECLARED varStatus = iota
	vs_DEFINED
	vs_READ
	vs_IMPLICIT
)

type Resolver struct {
	intprt *interpreter.Interpreter
	scopes []map[string]*varInfo
	curErr error
}

func NewResolver(intptr *interpreter.Interpreter) *Resolver {
	return &Resolver{
		intptr,
		make([]map[string]*varInfo, 0),
		nil,
	}
}

func (r *Resolver) ResolveStmts(stmts []ast.Stmt) error {
	for _, s := range stmts {
		r.resolveStmt(s)
	}
	return r.curErr
}

func (r *Resolver) resolveLocal(expr ast.Expr, name *token.Token, isRead bool) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		scope := r.scopes[i]
		if _, ok := scope[name.Lexeme]; ok {
			r.intprt.Resolve(expr, len(r.scopes)-1-i)
			if isRead {
				scope[name.Lexeme].status = vs_READ
			}
			return
		}
	}
}

func (r *Resolver) resolveFunction(fn *ast.Function) {
	r.beginScope()
	for _, param := range fn.Params {
		r.declare(param)
		r.define(param)
	}
	_ = r.ResolveStmts(fn.Body)
	r.endScope()
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]*varInfo))
}

func (r *Resolver) endScope() {
	for _, vi := range r.scopes[len(r.scopes)-1] {
		if vi.status == vs_DEFINED {
			r.reportTok(vi.name, ErrLocalNotRead)
		}
	}
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name *token.Token) {
	if len(r.scopes) == 0 {
		return
	}
	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope[name.Lexeme]; ok {
		r.reportTok(name, ErrAlreadyInScope)
	}
	scope[name.Lexeme] = &varInfo{name, vs_DECLARED}
}

func (r *Resolver) define(name *token.Token) {
	if len(r.scopes) == 0 {
		return
	}
	r.scopes[len(r.scopes)-1][name.Lexeme].status = vs_DEFINED
}

func (r *Resolver) reportTok(tok *token.Token, msg error) {
	errfmt := fmt.Sprintf("[line %d] [Resolver] Error at", tok.Line)
	if tok.Kind == token.EOF {
		fmt.Fprintf(os.Stderr, "%s end: %s\n", errfmt, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s '%s': %s\n", errfmt, tok.Lexeme, msg)
	}
	r.curErr = msg
}

func (r *Resolver) resolveExpr(exprNode ast.Expr) {
	switch expr := exprNode.(type) {
	case *ast.ArrayLiteral:
		for _, elm := range expr.Elements {
			r.resolveExpr(elm)
		}
	case *ast.Assign:
		r.resolveExpr(expr.Value)
		r.resolveLocal(expr, expr.Name, false)
	case *ast.Binary:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case *ast.Call:
		r.resolveExpr(expr.Callee)
		for _, arg := range expr.Arguments {
			r.resolveExpr(arg)
		}
	case *ast.Get:
		r.resolveExpr(expr.Object)
	case *ast.Grouping:
		r.resolveExpr(expr.Expression)
	case *ast.HashLiteral:
		for k, v := range expr.Pairs {
			r.resolveExpr(k)
			r.resolveExpr(v)
		}
	case *ast.IndexedGet:
		r.resolveExpr(expr.Object)
		if expr.Start != nil {
			r.resolveExpr(expr.Start)
		}
		if expr.Stop != nil {
			r.resolveExpr(expr.Stop)
		}
	case *ast.IndexedSet:
		r.resolveExpr(expr.Object)
		r.resolveExpr(expr.Index)
		r.resolveExpr(expr.Value)
	case *ast.Lambda:
		r.resolveFunction(expr.Func)
	case *ast.Literal:
		return
	case *ast.Logical:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case *ast.Set:
		r.resolveExpr(expr.Value)
		r.resolveExpr(expr.Object)
	case *ast.Super:
		r.resolveLocal(expr, expr.Keyword, true)
	case *ast.This:
		r.resolveLocal(expr, expr.Keyword, true)
	case *ast.Unary:
		r.resolveExpr(expr.Right)
	case *ast.Variable:
		if len(r.scopes) != 0 {
			state, ok := r.scopes[len(r.scopes)-1][expr.Name.Lexeme]
			if ok && state.status == vs_DECLARED {
				r.reportTok(expr.Name, ErrLocalInitializesSelf)
			}
		}
		r.resolveLocal(expr, expr.Name, true)
	default:
		panic(fmt.Sprintf("resolving is not implemented for %T", expr))
	}
}

func (r *Resolver) resolveStmt(stmtNode ast.Stmt) {
	switch stmt := stmtNode.(type) {
	case *ast.Block:
		r.beginScope()
		_ = r.ResolveStmts(stmt.Statements)
		r.endScope()
	case *ast.Class:
		r.declare(stmt.Name)
		r.define(stmt.Name)
		if stmt.Superclass != nil {
			r.resolveExpr(stmt.Superclass)
			r.beginScope()
			r.scopes[len(r.scopes)-1]["super"] = &varInfo{stmt.Superclass.Name, vs_IMPLICIT}
		}
		r.beginScope()
		r.scopes[len(r.scopes)-1]["this"] = &varInfo{stmt.Name, vs_IMPLICIT}
		for _, method := range stmt.Methods {
			r.resolveFunction(method)
		}
		r.endScope()
		if stmt.Superclass != nil {
			r.endScope()
		}
	case *ast.Expression:
		r.resolveExpr(stmt.Expression)
	case *ast.Function:
		r.declare(stmt.Name)
		r.define(stmt.Name)
		r.resolveFunction(stmt)
	case *ast.If:
		r.resolveExpr(stmt.Condition)
		r.resolveStmt(stmt.ThenBranch)
		if stmt.ElseBranch != nil {
			r.resolveStmt(stmt.ElseBranch)
		}
	case *ast.Print:
		r.resolveExpr(stmt.Expression)
	case *ast.Control:
		if stmt.Keyword.Kind == token.RETURN && stmt.Value != nil {
			r.resolveExpr(stmt.Value)
		}
		return
	case *ast.Var:
		r.declare(stmt.Name)
		if stmt.Initializer != nil {
			r.resolveExpr(stmt.Initializer)
		}
		r.define(stmt.Name)
	case *ast.While:
		r.resolveExpr(stmt.Condition)
		r.resolveStmt(stmt.Body)
	default:
		panic(fmt.Sprintf("resolving is not implemented for '%T'", stmt))
	}
}
