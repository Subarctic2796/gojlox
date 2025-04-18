package resolver

import (
	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/interpreter"
	"github.com/Subarctic2796/gojlox/token"
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

type Resolver struct {
	ER     errs.ErrorReporter
	intprt *interpreter.Interpreter
	scopes []map[string]*varInfo
	curFN  fnType
	curCLS clsType
}

func NewResolver(er errs.ErrorReporter, intptr *interpreter.Interpreter) *Resolver {
	return &Resolver{
		er,
		intptr,
		make([]map[string]*varInfo, 0),
		fn_NONE,
		cls_NONE,
	}
}

func (r *Resolver) ResolveStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		r.resolveStmt(s)
	}
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	expr.Accept(r)
}

func (r *Resolver) resolveLocal(expr ast.Expr, name *token.Token, isRead bool) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			r.intprt.Resolve(expr, len(r.scopes)-1-i)
			if isRead {
				r.scopes[i][name.Lexeme].status = vs_READ
			}
			return
		}
	}
}

func (r *Resolver) resolveLambda(fn *ast.Lambda, kind fnType) {
	enclosingFun := r.curFN
	r.curFN = kind
	r.beginScope()
	defer func() { r.endScope(); r.curFN = enclosingFun }()
	for _, param := range fn.Params {
		r.declare(param)
		r.define(param)
	}
	r.ResolveStmts(fn.Body)
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]*varInfo))
}

func (r *Resolver) endScope() {
	for _, vi := range r.scopes[len(r.scopes)-1] {
		if vi.status == vs_DEFINED {
			r.ER.ReportTok(vi.name, &errs.ResolverErr{
				Type: errs.LocalNotRead,
			})
		}
	}
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name *token.Token) {
	if len(r.scopes) == 0 {
		return
	}
	if _, ok := r.scopes[len(r.scopes)-1][name.Lexeme]; ok {
		r.ER.ReportTok(name, &errs.ResolverErr{
			Type: errs.AlreadyInScope,
		})
	}
	r.scopes[len(r.scopes)-1][name.Lexeme] = &varInfo{name, vs_DECLARED}
}

func (r *Resolver) define(name *token.Token) {
	if len(r.scopes) == 0 {
		return
	}
	r.scopes[len(r.scopes)-1][name.Lexeme].status = vs_DEFINED
}

func (r *Resolver) VisitAssignExpr(expr *ast.Assign) (any, error) {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name, false)
	return nil, nil
}

func (r *Resolver) VisitLambdaExpr(expr *ast.Lambda) (any, error) {
	r.resolveLambda(expr, fn_FUNC)
	return nil, nil
}

func (r *Resolver) VisitThisExpr(expr *ast.This) (any, error) {
	if r.curCLS == cls_NONE {
		r.ER.ReportTok(expr.Keyword, &errs.ResolverErr{
			Type: errs.ThisOutSideClass,
		})
		return nil, nil
	}
	r.resolveLocal(expr, expr.Keyword, true)
	return nil, nil
}

func (r *Resolver) VisitSetExpr(expr *ast.Set) (any, error) {
	r.resolveExpr(expr.Value)
	r.resolveExpr(expr.Object)
	return nil, nil
}

func (r *Resolver) VisitGetExpr(expr *ast.Get) (any, error) {
	r.resolveExpr(expr.Object)
	return nil, nil
}

func (r *Resolver) VisitSuperExpr(expr *ast.Super) (any, error) {
	if r.curCLS == cls_NONE {
		r.ER.ReportTok(expr.Keyword, &errs.ResolverErr{
			Type: errs.SuperOutSideClass,
		})
	} else if r.curCLS != cls_SUBCLASS {
		r.ER.ReportTok(expr.Keyword, &errs.ResolverErr{
			Type: errs.SuperWithNoSuperClass,
		})
	}
	if r.curFN == fn_STATIC {
		r.ER.ReportTok(expr.Keyword, &errs.ResolverErr{
			Type: errs.SuperInStatic,
		})
	}
	r.resolveLocal(expr, expr.Keyword, true)
	return nil, nil
}

func (r *Resolver) VisitBinaryExpr(expr *ast.Binary) (any, error) {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil, nil
}

func (r *Resolver) VisitCallExpr(expr *ast.Call) (any, error) {
	r.resolveExpr(expr.Callee)
	for _, arg := range expr.Arguments {
		r.resolveExpr(arg)
	}
	return nil, nil
}

func (r *Resolver) VisitGroupingExpr(expr *ast.Grouping) (any, error) {
	r.resolveExpr(expr.Expression)
	return nil, nil
}

func (r *Resolver) VisitLiteralExpr(expr *ast.Literal) (any, error) {
	return nil, nil
}

func (r *Resolver) VisitLogicalExpr(expr *ast.Logical) (any, error) {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil, nil
}

func (r *Resolver) VisitUnaryExpr(expr *ast.Unary) (any, error) {
	r.resolveExpr(expr.Right)
	return nil, nil
}

func (r *Resolver) VisitVariableExpr(expr *ast.Variable) (any, error) {
	if len(r.scopes) != 0 {
		state, ok := r.scopes[len(r.scopes)-1][expr.Name.Lexeme]
		if ok && state.status == vs_DECLARED {
			r.ER.ReportTok(expr.Name, &errs.ResolverErr{
				Type: errs.ReadLocalInOwnInitializer,
			})
		}
	}
	r.resolveLocal(expr, expr.Name, true)
	return nil, nil
}

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) (any, error) {
	r.beginScope()
	defer r.endScope()
	r.ResolveStmts(stmt.Statements)
	return nil, nil
}

func (r *Resolver) VisitExpressionStmt(stmt *ast.Expression) (any, error) {
	r.resolveExpr(stmt.Expression)
	return nil, nil
}

func (r *Resolver) VisitFunctionStmt(stmt *ast.Function) (any, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	r.resolveLambda(stmt.Func, fn_FUNC)
	return nil, nil
}

func (r *Resolver) VisitIfStmt(stmt *ast.If) (any, error) {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt(stmt.ElseBranch)
	}
	return nil, nil
}

func (r *Resolver) VisitPrintStmt(stmt *ast.Print) (any, error) {
	r.resolveExpr(stmt.Expression)
	return nil, nil
}

func (r *Resolver) VisitBreakStmt(stmt *ast.Break) (any, error) {
	return nil, nil
}

func (r *Resolver) VisitReturnStmt(stmt *ast.Return) (any, error) {
	if r.curFN == fn_NONE {
		r.ER.ReportTok(stmt.Keyword, &errs.ResolverErr{
			Type: errs.ReturnTopLevel,
		})
	}
	if stmt.Value != nil {
		if r.curFN == fn_INIT {
			r.ER.ReportTok(stmt.Keyword, &errs.ResolverErr{
				Type: errs.ReturnFromInit,
			})
		}
		r.resolveExpr(stmt.Value)
	}
	return nil, nil
}

func (r *Resolver) VisitVarStmt(stmt *ast.Var) (any, error) {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil, nil
}

func (r *Resolver) VisitWhileStmt(stmt *ast.While) (any, error) {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	return nil, nil
}

func (r *Resolver) VisitClassStmt(stmt *ast.Class) (any, error) {
	enclosingCLS := r.curCLS
	r.curCLS = cls_CLASS
	r.declare(stmt.Name)
	r.define(stmt.Name)
	if stmt.Superclass != nil {
		if stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
			r.ER.ReportTok(stmt.Superclass.Name, &errs.ResolverErr{
				Type: errs.SelfInheritance,
			})
		}
		r.curCLS = cls_SUBCLASS
		r.resolveExpr(stmt.Superclass)
		r.beginScope()
		r.scopes[len(r.scopes)-1]["super"] = &varInfo{stmt.Superclass.Name, vs_IMPLICIT}
	}
	r.beginScope()
	defer func() {
		r.endScope()
		r.curCLS = enclosingCLS
		if stmt.Superclass != nil {
			r.endScope()
		}
	}()
	r.scopes[len(r.scopes)-1]["this"] = &varInfo{stmt.Name, vs_IMPLICIT}
	for _, method := range stmt.Methods {
		decl := fn_METHOD
		if method.Name.Lexeme == "init" {
			decl = fn_INIT
		}
		r.resolveLambda(method.Func, decl)
	}
	for _, method := range stmt.Statics {
		if method.Name.Lexeme == "init" {
			r.ER.ReportTok(method.Name, &errs.ResolverErr{
				Type: errs.InitIsStatic,
			})
		}
		r.resolveLambda(method.Func, fn_STATIC)
	}
	// if stmt.Superclass != nil {
	// 	r.endScope()
	// }
	return nil, nil
}
