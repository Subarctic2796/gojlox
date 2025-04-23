package resolver

import (
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/ast"
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

type clsType int

const (
	cls_NONE clsType = iota
	cls_CLASS
	cls_SUBCLASS
)

type Resolver struct {
	intprt *interpreter.Interpreter
	useV2  bool
	scopes []map[string]*varInfo
	curFN  ast.FnType
	curCLS clsType
	curErr error
}

func NewResolver(intptr *interpreter.Interpreter, useV2 bool) *Resolver {
	return &Resolver{
		intptr,
		useV2,
		make([]map[string]*varInfo, 0),
		ast.FN_NONE,
		cls_NONE,
		nil,
	}
}

func (r *Resolver) ResolveStmts(stmts []ast.Stmt) error {
	var fn func(stmt ast.Stmt)
	if !r.useV2 {
		fn = r.resolveStmt
	} else {
		fn = r.resolveStmt2
	}
	for _, s := range stmts {
		fn(s)
	}
	return r.curErr
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
	_, _ = stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	_, _ = expr.Accept(r)
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

func (r *Resolver) resolveLambda(fn *ast.Lambda, kind ast.FnType) {
	enclosingFun := r.curFN
	r.curFN = kind
	r.beginScope()
	defer func() { r.endScope(); r.curFN = enclosingFun }()
	for _, param := range fn.Params {
		r.declare(param)
		r.define(param)
	}
	_ = r.ResolveStmts(fn.Body)
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
	if _, ok := r.scopes[len(r.scopes)-1][name.Lexeme]; ok {
		r.reportTok(name, ErrAlreadyInScope)
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
	r.resolveLambda(expr, ast.FN_FUNC)
	return nil, nil
}

func (r *Resolver) VisitThisExpr(expr *ast.This) (any, error) {
	// moved to parser
	/* if r.curCLS == cls_NONE {
		r.reportTok(expr.Keyword, ErrThisNotInClass)
		return nil, nil
	} */
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
	// moved to parser
	/* if r.curCLS == cls_NONE {
		r.reportTok(expr.Keyword, ErrSuperNotInClass)
	} */
	// } else if r.curCLS != cls_SUBCLASS {
	if r.curCLS != cls_SUBCLASS {
		r.reportTok(expr.Keyword, ErrSuperWithNoSuperClass)
	}
	if r.curFN == ast.FN_STATIC {
		r.reportTok(expr.Keyword, ErrSuperInStatic)
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
			r.reportTok(expr.Name, ErrReadLocalInOwnInitializer)
		}
	}
	r.resolveLocal(expr, expr.Name, true)
	return nil, nil
}

func (r *Resolver) VisitBlockStmt(stmt *ast.Block) (any, error) {
	r.beginScope()
	defer func() { r.endScope() }()
	_ = r.ResolveStmts(stmt.Statements)
	return nil, nil
}

func (r *Resolver) VisitExpressionStmt(stmt *ast.Expression) (any, error) {
	r.resolveExpr(stmt.Expression)
	return nil, nil
}

func (r *Resolver) VisitFunctionStmt(stmt *ast.Function) (any, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	r.resolveLambda(stmt.Func, ast.FN_FUNC)
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
	// moved to parser
	/* if r.curFN == ast.FN_NONE {
		r.reportTok(stmt.Keyword, ErrReturnTopLevel)
	} */
	if stmt.Value != nil {
		if r.curFN == ast.FN_INIT {
			r.reportTok(stmt.Keyword, ErrReturnFromInit)
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
		// moved to parser
		/* if stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
			r.reportTok(stmt.Superclass.Name, ErrInheritsSelf)
		} */
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
		decl := ast.FN_METHOD
		if method.Func.Kind == ast.FN_STATIC {
			decl = ast.FN_STATIC
		}
		if method.Name.Lexeme == "init" {
			if method.Func.Kind == ast.FN_STATIC {
				r.reportTok(method.Name, ErrInitIsStatic)
			}
			method.Func.Kind = ast.FN_INIT
			decl = ast.FN_INIT
		}
		r.resolveLambda(method.Func, decl)
	}
	return nil, nil
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

func (r *Resolver) resolveExpr2(expr ast.Expr) {
	switch ex := expr.(type) {
	case *ast.Assign:
		r.exprAssign(ex)
	case *ast.Binary:
		r.exprBinary(ex)
	case *ast.Call:
		r.exprCall(ex)
	case *ast.Get:
		r.exprGet(ex)
	case *ast.Grouping:
		r.exprGrouping(ex)
	case *ast.Lambda:
		r.exprLambda(ex)
	case *ast.Literal:
		r.exprLiteral(ex)
	case *ast.Logical:
		r.exprLogical(ex)
	case *ast.Set:
		r.exprSet(ex)
	case *ast.Super:
		r.exprSuper(ex)
	case *ast.This:
		r.exprThis(ex)
	case *ast.Unary:
		r.exprUnary(ex)
	case *ast.Variable:
		r.exprVariable(ex)
	}
}

func (r *Resolver) resolveStmt2(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.Block:
		r.stmtBlock(s)
	case *ast.Break:
		r.stmtBreak(s)
	case *ast.Class:
		r.classStmt(s)
	case *ast.Expression:
		r.stmtExpression(s)
	case *ast.Function:
		r.stmtFunction(s)
	case *ast.If:
		r.stmtIf(s)
	case *ast.Print:
		r.stmtPrint(s)
	case *ast.Return:
		r.stmtReturn(s)
	case *ast.Var:
		r.stmtVar(s)
	case *ast.While:
		r.stmtWhile(s)
	}
}

func (r *Resolver) exprVariable(expr *ast.Variable) {
	if len(r.scopes) != 0 {
		state, ok := r.scopes[len(r.scopes)-1][expr.Name.Lexeme]
		if ok && state.status == vs_DECLARED {
			r.reportTok(expr.Name, ErrReadLocalInOwnInitializer)
		}
	}
	r.resolveLocal(expr, expr.Name, true)
}

func (r *Resolver) exprUnary(expr *ast.Unary) {
	r.resolveExpr2(expr.Right)
}

func (r *Resolver) exprThis(expr *ast.This) {
	r.resolveLocal(expr, expr.Keyword, true)
}

func (r *Resolver) exprSuper(expr *ast.Super) {
	if r.curCLS != cls_SUBCLASS {
		r.reportTok(expr.Keyword, ErrSuperWithNoSuperClass)
	}
	if r.curFN == ast.FN_STATIC {
		r.reportTok(expr.Keyword, ErrSuperInStatic)
	}
	r.resolveLocal(expr, expr.Keyword, true)
}

func (r *Resolver) exprSet(expr *ast.Set) {
	r.resolveExpr2(expr.Value)
	r.resolveExpr2(expr.Object)
}

func (r *Resolver) exprLogical(expr *ast.Logical) {
	r.resolveExpr2(expr.Left)
	r.resolveExpr2(expr.Right)
}

func (r *Resolver) exprLiteral(_ *ast.Literal) {}

func (r *Resolver) exprLambda(expr *ast.Lambda) {
	r.resolveLambda(expr, ast.FN_FUNC)
}

func (r *Resolver) exprGrouping(expr *ast.Grouping) {
	r.resolveExpr2(expr.Expression)
}

func (r *Resolver) exprGet(expr *ast.Get) {
	r.resolveExpr2(expr.Object)
}

func (r *Resolver) exprCall(expr *ast.Call) {
	r.resolveExpr2(expr.Callee)
	for _, arg := range expr.Arguments {
		r.resolveExpr2(arg)
	}
}

func (r *Resolver) exprBinary(expr *ast.Binary) {
	r.resolveExpr2(expr.Left)
	r.resolveExpr2(expr.Right)
}

func (r *Resolver) exprAssign(expr *ast.Assign) {
	r.resolveExpr2(expr.Value)
	r.resolveLocal(expr, expr.Name, false)
}

func (r *Resolver) stmtBlock(stmt *ast.Block) {
	r.beginScope()
	defer func() { r.endScope() }()
	_ = r.ResolveStmts(stmt.Statements)
}

func (r *Resolver) stmtBreak(_ *ast.Break) {}

func (r *Resolver) classStmt(stmt *ast.Class) {
	enclosingCLS := r.curCLS
	r.curCLS = cls_CLASS
	r.declare(stmt.Name)
	r.define(stmt.Name)
	if stmt.Superclass != nil {
		// moved to parser
		/* if stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
			r.reportTok(stmt.Superclass.Name, ErrInheritsSelf)
		} */
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
		decl := ast.FN_METHOD
		if method.Func.Kind == ast.FN_STATIC {
			decl = ast.FN_STATIC
		}
		if method.Name.Lexeme == "init" {
			if method.Func.Kind == ast.FN_STATIC {
				r.reportTok(method.Name, ErrInitIsStatic)
			}
			method.Func.Kind = ast.FN_INIT
			decl = ast.FN_INIT
		}
		r.resolveLambda(method.Func, decl)
	}
}

func (r *Resolver) stmtExpression(stmt *ast.Expression) {
	r.resolveExpr2(stmt.Expression)
}

func (r *Resolver) stmtFunction(stmt *ast.Function) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	r.resolveLambda(stmt.Func, ast.FN_FUNC)
}

func (r *Resolver) stmtIf(stmt *ast.If) {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt2(stmt.ElseBranch)
	}
}

func (r *Resolver) stmtPrint(stmt *ast.Print) {
	r.resolveExpr2(stmt.Expression)
}

func (r *Resolver) stmtReturn(stmt *ast.Return) {
	if stmt.Value != nil {
		if r.curFN == ast.FN_INIT {
			r.reportTok(stmt.Keyword, ErrReturnFromInit)
		}
		r.resolveExpr2(stmt.Value)
	}
}

func (r *Resolver) stmtVar(stmt *ast.Var) {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr2(stmt.Initializer)
	}
	r.define(stmt.Name)
}

func (r *Resolver) stmtWhile(stmt *ast.While) {
	r.resolveExpr2(stmt.Condition)
	r.resolveStmt2(stmt.Body)
}
