package interpreter

import (
	"errors"
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/token"
)

type Interpreter struct {
	useV2        bool
	Globals, env *Env
	locals       map[ast.Expr]int
	CurErr       error
	tmpBin       *ast.Binary
}

func NewInterpreter(useV2 bool) *Interpreter {
	globals := NewEnv(nil)
	for _, fn := range NativeFns {
		globals.Define(fn.Name(), fn)
	}
	return &Interpreter{
		useV2,
		globals,
		globals,
		make(map[ast.Expr]int),
		nil,
		&ast.Binary{Left: nil, Operator: token.NewToken(token.NONE, "", nil, -1), Right: nil},
	}
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) error {
	var executor func(stmt ast.Stmt) (any, error)
	if !i.useV2 {
		executor = i.execute
	} else {
		executor = i.execute2
	}
	for _, s := range stmts {
		// _, err := i.execute(s)
		_, err := executor(s)
		if err != nil {
			i.reportRunTimeErr(err)
			return err
		}
	}
	return nil
}

func (i *Interpreter) Resolve(expr ast.Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) evaluate(expr ast.Expr) (any, error) {
	return expr.Accept(i)
}

func (i *Interpreter) execute(stmt ast.Stmt) (any, error) {
	return stmt.Accept(i)
}

func (i *Interpreter) stringify(obj any) string {
	if obj == nil {
		return "nil"
	}
	return fmt.Sprint(obj)
}

func (i *Interpreter) lookUpVariable(name *token.Token, expr ast.Expr) (any, error) {
	if dist, ok := i.locals[expr]; ok {
		return i.env.GetAt(dist, name.Lexeme), nil
	} else {
		return i.Globals.Get(name)
	}
}

func (i *Interpreter) VisitBreakStmt(stmt *ast.Break) (any, error) {
	return nil, BreakErr
}

func (i *Interpreter) VisitReturnStmt(stmt *ast.Return) (any, error) {
	var val any
	var err error
	if stmt.Value != nil {
		val, err = i.evaluate(stmt.Value)
		if err != nil {
			return nil, err
		}
	}
	return nil, &ReturnErr{Value: val}
}

func (i *Interpreter) VisitClassStmt(stmt *ast.Class) (any, error) {
	var supercls any = nil
	var err error
	if stmt.Superclass != nil {
		supercls, err = i.evaluate(stmt.Superclass)
		if err != nil {
			return nil, err
		}
		if _, ok := supercls.(*LoxClass); !ok {
			return nil, &RunTimeErr{
				Tok: stmt.Superclass.Name,
				Msg: "Superclass must be a class",
			}
		}
	}
	i.env.Define(stmt.Name.Lexeme, nil)
	if stmt.Superclass != nil {
		i.env = NewEnv(i.env)
		i.env.Define("super", supercls)
	}
	methods := make(map[string]*LoxFn)
	for _, method := range stmt.Methods {
		methods[method.Name.Lexeme] = NewLoxFn(method.Name.Lexeme, method.Func, i.env)
	}
	scls, _ := supercls.(*LoxClass)
	klass := NewLoxClass(stmt.Name.Lexeme, scls, methods)
	if supercls != nil {
		i.env = i.env.Enclosing
	}
	err = i.env.Assign(stmt.Name, klass)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) VisitFunctionStmt(stmt *ast.Function) (any, error) {
	name := stmt.Name.Lexeme
	fn := NewLoxFn(name, stmt.Func, i.env)
	i.env.Define(stmt.Name.Lexeme, fn)
	return nil, nil
}

func (i *Interpreter) VisitLambdaExpr(expr *ast.Lambda) (any, error) {
	return NewLoxFn("", expr, i.env), nil
}

func (i *Interpreter) VisitSuperExpr(expr *ast.Super) (any, error) {
	dist := i.locals[expr]
	superclass := i.env.GetAt(dist, "super").(*LoxClass)
	obj := i.env.GetAt(dist-1, "this").(*LoxInstance)
	method := superclass.FindMethod(expr.Method.Lexeme)
	if method == nil {
		return nil, &RunTimeErr{
			Tok: expr.Method,
			Msg: fmt.Sprintf("Undefined property '%s'", expr.Method.Lexeme),
		}
	}
	return method.Bind(obj), nil
}

func (i *Interpreter) VisitCallExpr(expr *ast.Call) (any, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(expr.Arguments))
	for _, arg := range expr.Arguments {
		a, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	fn, ok := callee.(LoxCallable)
	if !ok {
		return nil, &RunTimeErr{
			Tok: expr.Paren,
			Msg: "Can only call functions and classes",
		}
	}
	if fn.Arity() == -1 {
		return fn.Call(i, args)
	}
	if len(args) != fn.Arity() {
		msg := fmt.Sprintf("Expected %d arguments but got %d", fn.Arity(), len(args))
		return nil, &RunTimeErr{Tok: expr.Paren, Msg: msg}
	}
	return fn.Call(i, args)
}

func (i *Interpreter) VisitGetExpr(expr *ast.Get) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	if klass, ok := obj.(*LoxClass); ok {
		static := klass.FindMethod(expr.Name.Lexeme)
		if static != nil {
			if static.Func.Kind != ast.FN_STATIC {
				return nil, &RunTimeErr{
					Tok: expr.Name,
					Msg: fmt.Sprintf("Undefined static function '%s'", expr.Name.Lexeme),
				}
			}
			return static, nil
		}
	}
	if inst, ok := obj.(*LoxInstance); ok {
		return inst.Get(expr.Name)
	}
	return nil, &RunTimeErr{
		Tok: expr.Name,
		Msg: "Only instances have properties",
	}
}

func (i *Interpreter) VisitThisExpr(expr *ast.This) (any, error) {
	return i.lookUpVariable(expr.Keyword, expr)
}

func (i *Interpreter) VisitSetExpr(expr *ast.Set) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	inst, ok := obj.(*LoxInstance)
	if !ok {
		return nil, &RunTimeErr{
			Tok: expr.Name,
			Msg: "Only instances have fields",
		}
	}
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	inst.Set(expr.Name, val)
	return val, nil
}

func (i *Interpreter) VisitWhileStmt(stmt *ast.While) (any, error) {
	cond, err := i.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	for i.isTruthy(cond) {
		_, err = i.execute(stmt.Body)
		if err != nil {
			if errors.Is(err, BreakErr) {
				return nil, nil
			}
			return nil, err
		}
		cond, err = i.evaluate(stmt.Condition)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) VisitLogicalExpr(expr *ast.Logical) (any, error) {
	lhs, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	if expr.Operator.Kind == token.OR {
		if i.isTruthy(lhs) {
			return lhs, nil
		}
	} else {
		if !i.isTruthy(lhs) {
			return lhs, nil
		}
	}
	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitIfStmt(stmt *ast.If) (any, error) {
	cond, err := i.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if i.isTruthy(cond) {
		_, err = i.execute(stmt.ThenBranch)
		if err != nil {
			return nil, err
		}
	} else if stmt.ElseBranch != nil {
		_, err = i.execute(stmt.ElseBranch)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) VisitBlockStmt(stmt *ast.Block) (any, error) {
	return i.executeBlock(stmt.Statements, NewEnv(i.env))
}

func (i *Interpreter) executeBlock(stmts []ast.Stmt, env *Env) (any, error) {
	prv := i.env
	defer func() { i.env = prv }()
	i.env = env
	for _, stmt := range stmts {
		if _, err := i.execute(stmt); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) VisitAssignExpr(expr *ast.Assign) (any, error) {
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	oprType := token.NONE
	switch expr.Operator.Kind {
	case token.PLUS_EQUAL:
		oprType = token.PLUS
	case token.MINUS_EQUAL:
		oprType = token.MINUS
	case token.SLASH_EQUAL:
		oprType = token.SLASH
	case token.STAR_EQUAL:
		oprType = token.STAR
	}
	if oprType != token.NONE {
		tmp, err := i.lookUpVariable(expr.Name, expr)
		if err != nil {
			return nil, err
		}
		// lval, rval := &ast.Literal{Value: tmp}, &ast.Literal{Value: val}
		// opr := &token.Token{Kind: oprType, Lexeme: "", Literal: nil, Line: expr.Operator.Line}
		// val, err = i.VisitBinaryExpr(&ast.Binary{Left: lval, Operator: opr, Right: rval})
		i.tmpBin.Left = &ast.Literal{Value: tmp}
		i.tmpBin.Right = &ast.Literal{Value: val}
		i.tmpBin.Operator.Kind = oprType
		i.tmpBin.Operator.Line = expr.Operator.Line
		val, err = i.VisitBinaryExpr(i.tmpBin)
		if err != nil {
			return nil, err
		}
	}
	if dist, ok := i.locals[expr]; ok {
		i.env.AssignAt(dist, expr.Name, val)
	} else {
		err = i.Globals.Assign(expr.Name, val)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (i *Interpreter) VisitVariableExpr(expr *ast.Variable) (any, error) {
	return i.lookUpVariable(expr.Name, expr)
}

func (i *Interpreter) VisitVarStmt(stmt *ast.Var) (any, error) {
	var val any
	var err error
	if stmt.Initializer != nil {
		val, err = i.evaluate(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}
	i.env.Define(stmt.Name.Lexeme, val)
	return nil, nil
}

func (i *Interpreter) VisitBinaryExpr(expr *ast.Binary) (any, error) {
	lhs, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	rhs, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Kind {
	case token.GREATER:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l > r, nil
	case token.GREATER_EQUAL:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l >= r, nil
	case token.LESS:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l < r, nil
	case token.LESS_EQUAL:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l <= r, nil
	case token.MINUS:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l - r, nil
	case token.BANG_EQUAL:
		return !i.isEqual(lhs, rhs), nil
	case token.EQUAL_EQUAL:
		return i.isEqual(lhs, rhs), nil
	case token.PLUS:
		l, lok := lhs.(float64)
		r, rok := rhs.(float64)
		if lok && rok {
			return l + r, nil
		}
		ls, lok := lhs.(string)
		rs, rok := rhs.(string)
		if lok && rok {
			return ls + rs, nil
		}
		return nil, &RunTimeErr{
			Tok: expr.Operator,
			Msg: "Operands must be two numbers or two strings",
		}
	case token.SLASH:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		if r == 0.0 {
			return nil, &RunTimeErr{
				Tok: expr.Operator,
				Msg: "Division by 0",
			}
		}
		return l / r, nil
	case token.STAR:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	}
	// unreachable
	return nil, nil
}

func (i *Interpreter) VisitGroupingExpr(expr *ast.Grouping) (any, error) {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr *ast.Literal) (any, error) {
	return expr.Value, nil
}

func (i *Interpreter) VisitUnaryExpr(expr *ast.Unary) (any, error) {
	rhs, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Operator.Kind {
	case token.MINUS:
		r, err := i.checkNumberOperand(expr.Operator, rhs)
		if err != nil {
			return nil, err
		}
		return -r, nil
	case token.BANG:
		return !i.isTruthy(rhs), nil
	}

	// unreachable
	return nil, nil
}

func (i *Interpreter) VisitExpressionStmt(stmt *ast.Expression) (any, error) {
	_, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) VisitPrintStmt(stmt *ast.Print) (any, error) {
	val, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(i.stringify(val))
	return nil, nil
}

func (i *Interpreter) isTruthy(obj any) bool {
	if obj == nil {
		return false
	}
	if bobj, ok := obj.(bool); ok {
		return bobj
	}
	return true
}

func (i *Interpreter) isEqual(a any, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func (i *Interpreter) checkNumberOperand(oprtr *token.Token, opr any) (float64, error) {
	if r, ok := opr.(float64); ok {
		return r, nil
	}
	return 0, &RunTimeErr{
		Tok: oprtr,
		Msg: "Operand must be a number",
	}
}

func (i *Interpreter) checkNumberOperands(oprtr *token.Token, lhs any, rhs any) (float64, float64, error) {
	l, lok := lhs.(float64)
	r, rok := rhs.(float64)
	if lok && rok {
		return l, r, nil
	}
	return 0, 0, &RunTimeErr{
		Tok: oprtr,
		Msg: "Operands must be a number",
	}
}

func (i *Interpreter) reportRunTimeErr(msg error) {
	fmt.Fprintln(os.Stderr, msg)
	i.CurErr = msg
}

func (i *Interpreter) evaluate2(expr ast.Expr) (any, error) {
	switch ex := expr.(type) {
	case *ast.Assign:
		return i.exprAssign(ex)
	case *ast.Binary:
		return i.exprBinary(ex)
	case *ast.Call:
		return i.exprCall(ex)
	case *ast.Get:
		return i.exprGet(ex)
	case *ast.Grouping:
		return i.exprGrouping(ex)
	case *ast.Lambda:
		return i.exprLambda(ex)
	case *ast.Literal:
		return i.exprLiteral(ex)
	case *ast.Logical:
		return i.exprLogical(ex)
	case *ast.Set:
		return i.exprSet(ex)
	case *ast.Super:
		return i.exprSuper(ex)
	case *ast.This:
		return i.exprThis(ex)
	case *ast.Unary:
		return i.exprUnary(ex)
	case *ast.Variable:
		return i.exprVariable(ex)
	}
	return nil, nil
}

func (i *Interpreter) execute2(stmt ast.Stmt) (any, error) {
	switch s := stmt.(type) {
	case *ast.Block:
		return i.stmtBlock(s)
	case *ast.Break:
		return i.stmtBreak(s)
	case *ast.Class:
		return i.classStmt(s)
	case *ast.Expression:
		return i.stmtExpression(s)
	case *ast.Function:
		return i.stmtFunction(s)
	case *ast.If:
		return i.stmtIf(s)
	case *ast.Print:
		return i.stmtPrint(s)
	case *ast.Return:
		return i.stmtReturn(s)
	case *ast.Var:
		return i.stmtVar(s)
	case *ast.While:
		return i.stmtWhile(s)
	}
	return nil, nil
}

func (i *Interpreter) stmtWhile(stmt *ast.While) (any, error) {
	cond, err := i.evaluate2(stmt.Condition)
	if err != nil {
		return nil, err
	}
	for i.isTruthy(cond) {
		_, err = i.execute2(stmt.Body)
		if err != nil {
			if errors.Is(err, BreakErr) {
				return nil, nil
			}
			return nil, err
		}
		cond, err = i.evaluate2(stmt.Condition)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) stmtVar(stmt *ast.Var) (any, error) {
	var val any
	var err error
	if stmt.Initializer != nil {
		val, err = i.evaluate2(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}
	i.env.Define(stmt.Name.Lexeme, val)
	return nil, nil
}

func (i *Interpreter) stmtReturn(stmt *ast.Return) (any, error) {
	var val any
	var err error
	if stmt.Value != nil {
		val, err = i.evaluate2(stmt.Value)
		if err != nil {
			return nil, err
		}
	}
	return nil, &ReturnErr{Value: val}
}

func (i *Interpreter) stmtPrint(stmt *ast.Print) (any, error) {
	val, err := i.evaluate2(stmt.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(i.stringify(val))
	return nil, nil
}

func (i *Interpreter) stmtIf(stmt *ast.If) (any, error) {
	cond, err := i.evaluate2(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if i.isTruthy(cond) {
		_, err := i.execute2(stmt.ThenBranch)
		if err != nil {
			return nil, err
		}
	} else if stmt.ElseBranch != nil {
		_, err := i.execute2(stmt.ElseBranch)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) stmtFunction(stmt *ast.Function) (any, error) {
	name := stmt.Name.Lexeme
	fn := NewLoxFn(name, stmt.Func, i.env)
	i.env.Define(name, fn)
	return nil, nil
}

func (i *Interpreter) stmtExpression(stmt *ast.Expression) (any, error) {
	_, err := i.evaluate2(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) classStmt(stmt *ast.Class) (any, error) {
	var supercls any = nil
	var err error
	if stmt.Superclass != nil {
		supercls, err := i.exprVariable(stmt.Superclass)
		if err != nil {
			return nil, err
		}
		if _, ok := supercls.(*LoxClass); !ok {
			return nil, &RunTimeErr{
				Tok: stmt.Superclass.Name,
				Msg: "Superclass must be a class",
			}
		}
	}
	i.env.Define(stmt.Name.Lexeme, nil)
	if stmt.Superclass != nil {
		i.env = NewEnv(i.env)
		i.env.Define("super", supercls)
	}
	methods := make(map[string]*LoxFn)
	for _, method := range stmt.Methods {
		methods[method.Name.Lexeme] = NewLoxFn(method.Name.Lexeme, method.Func, i.env)
	}
	scls, _ := supercls.(*LoxClass)
	klass := NewLoxClass(stmt.Name.Lexeme, scls, methods)
	if supercls != nil {
		i.env = i.env.Enclosing
	}
	err = i.env.Assign(stmt.Name, klass)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) stmtBreak(_ *ast.Break) (any, error) {
	return nil, BreakErr
}

func (i *Interpreter) stmtBlock(stmt *ast.Block) (any, error) {
	return i.executeBlock2(stmt.Statements, NewEnv(i.env))
}

func (i *Interpreter) executeBlock2(stmts []ast.Stmt, env *Env) (any, error) {
	prv := i.env
	defer func() { i.env = prv }()
	i.env = env
	for _, stmt := range stmts {
		_, err := i.execute2(stmt)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) exprAssign(expr *ast.Assign) (any, error) {
	val, err := i.evaluate2(expr.Value)
	if err != nil {
		return nil, err
	}
	oprType := token.NONE
	switch expr.Operator.Kind {
	case token.PLUS_EQUAL:
		oprType = token.PLUS
	case token.MINUS_EQUAL:
		oprType = token.MINUS
	case token.SLASH_EQUAL:
		oprType = token.SLASH
	case token.STAR_EQUAL:
		oprType = token.STAR
	}
	if oprType != token.NONE {
		tmp, err := i.lookUpVariable(expr.Name, expr)
		if err != nil {
			return nil, err
		}
		/* lval, rval := &ast.Literal{Value: tmp}, &ast.Literal{Value: val}
		opr := &token.Token{Kind: oprType, Lexeme: "", Literal: nil, Line: expr.Operator.Line}
		val, err = i.VisitBinaryExpr(&ast.Binary{Left: lval, Operator: opr, Right: rval}) */
		i.tmpBin.Left = &ast.Literal{Value: tmp}
		i.tmpBin.Right = &ast.Literal{Value: val}
		i.tmpBin.Operator.Kind = oprType
		i.tmpBin.Operator.Line = expr.Operator.Line
		val, err = i.exprBinary(i.tmpBin)
		if err != nil {
			return nil, err
		}
	}
	if dist, ok := i.locals[expr]; ok {
		i.env.AssignAt(dist, expr.Name, val)
	} else {
		err = i.Globals.Assign(expr.Name, val)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func (i *Interpreter) exprBinary(expr *ast.Binary) (any, error) {
	lhs, err := i.evaluate2(expr.Left)
	if err != nil {
		return nil, err
	}
	rhs, err := i.evaluate2(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Kind {
	case token.GREATER:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l > r, nil
	case token.GREATER_EQUAL:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l >= r, nil
	case token.LESS:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l < r, nil
	case token.LESS_EQUAL:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l <= r, nil
	case token.MINUS:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l - r, nil
	case token.BANG_EQUAL:
		return !i.isEqual(lhs, rhs), nil
	case token.EQUAL_EQUAL:
		return i.isEqual(lhs, rhs), nil
	case token.PLUS:
		if l, ok := lhs.(float64); ok {
			if r, ok := rhs.(float64); ok {
				return l + r, nil
			}
		}
		if l, ok := lhs.(string); ok {
			if r, ok := rhs.(string); ok {
				return l + r, nil
			}
		}
		return nil, &RunTimeErr{
			Tok: expr.Operator,
			Msg: "Operands must be two numbers or two strings",
		}
	case token.SLASH:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		if r == 0.0 {
			return nil, &RunTimeErr{
				Tok: expr.Operator,
				Msg: "Division by 0",
			}
		}
		return l / r, nil
	case token.STAR:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	}
	// unreachable
	return nil, nil
}

func (i *Interpreter) exprCall(expr *ast.Call) (any, error) {
	callee, err := i.evaluate2(expr.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(expr.Arguments))
	for _, arg := range expr.Arguments {
		a, err := i.evaluate2(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	fn, ok := callee.(LoxCallable)
	if !ok {
		return nil, &RunTimeErr{
			Tok: expr.Paren,
			Msg: "Can only call functions and classes",
		}
	}
	if fn.Arity() == -1 {
		return fn.Call(i, args)
	}
	if len(args) != fn.Arity() {
		msg := fmt.Sprintf("Expected %d arguments but got %d", fn.Arity(), len(args))
		return nil, &RunTimeErr{Tok: expr.Paren, Msg: msg}
	}
	return fn.Call(i, args)
}

func (i *Interpreter) exprGet(expr *ast.Get) (any, error) {
	obj, err := i.evaluate2(expr.Object)
	if err != nil {
		return nil, err
	}
	if klass, ok := obj.(*LoxClass); ok {
		static := klass.FindMethod(expr.Name.Lexeme)
		if static != nil {
			if static.Func.Kind != ast.FN_STATIC {
				return nil, &RunTimeErr{
					Tok: expr.Name,
					Msg: fmt.Sprintf("Undefined static function '%s'", expr.Name.Lexeme),
				}
			}
			return static, nil
		}
	}
	if inst, ok := obj.(*LoxInstance); ok {
		return inst.Get(expr.Name)
	}
	return nil, &RunTimeErr{
		Tok: expr.Name,
		Msg: "Only instances have properties",
	}
}

func (i *Interpreter) exprGrouping(expr *ast.Grouping) (any, error) {
	return i.evaluate2(expr.Expression)
}

func (i *Interpreter) exprLambda(expr *ast.Lambda) (any, error) {
	return NewLoxFn("", expr, i.env), nil
}

func (i *Interpreter) exprLiteral(expr *ast.Literal) (any, error) {
	return expr.Value, nil
}

func (i *Interpreter) exprLogical(expr *ast.Logical) (any, error) {
	lhs, err := i.evaluate2(expr.Left)
	if err != nil {
		return nil, err
	}
	if expr.Operator.Kind == token.OR {
		if i.isTruthy(lhs) {
			return lhs, nil
		}
	} else {
		if !i.isTruthy(lhs) {
			return lhs, nil
		}
	}
	return i.evaluate2(expr.Right)
}

func (i *Interpreter) exprSet(expr *ast.Set) (any, error) {
	obj, err := i.evaluate2(expr.Object)
	if err != nil {
		return nil, err
	}
	inst, ok := obj.(*LoxInstance)
	if !ok {
		return nil, &RunTimeErr{
			Tok: expr.Name,
			Msg: "Only instances have fields",
		}
	}
	val, err := i.evaluate2(expr.Value)
	if err != nil {
		return nil, err
	}
	inst.Set(expr.Name, val)
	return val, nil
}

func (i *Interpreter) exprSuper(expr *ast.Super) (any, error) {
	dist := i.locals[expr]
	superclass := i.env.GetAt(dist, "super").(*LoxClass)
	obj := i.env.GetAt(dist-1, "this").(*LoxInstance)
	method := superclass.FindMethod(expr.Method.Lexeme)
	if method == nil {
		return nil, &RunTimeErr{
			Tok: expr.Method,
			Msg: fmt.Sprintf("Undefined property '%s'", expr.Method.Lexeme),
		}
	}
	return method.Bind(obj), nil
}

func (i *Interpreter) exprThis(expr *ast.This) (any, error) {
	return i.lookUpVariable(expr.Keyword, expr)
}

func (i *Interpreter) exprUnary(expr *ast.Unary) (any, error) {
	rhs, err := i.evaluate2(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Operator.Kind {
	case token.MINUS:
		r, err := i.checkNumberOperand(expr.Operator, rhs)
		if err != nil {
			return nil, err
		}
		return -r, nil
	case token.BANG:
		return !i.isTruthy(rhs), nil
	}
	// unreachable
	return nil, nil
}

func (i *Interpreter) exprVariable(expr *ast.Variable) (any, error) {
	return i.lookUpVariable(expr.Name, expr)
}
