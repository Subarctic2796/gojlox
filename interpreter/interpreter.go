package interpreter

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/token"
)

type Interpreter struct {
	Globals, env *Env
	locals       map[ast.Expr]int
	CurErr       error
	tmpBin       *ast.Binary
}

func NewInterpreter() *Interpreter {
	globals := NewEnv(nil)
	for name, fn := range NativeFns {
		globals.Define(name, fn)
	}
	return &Interpreter{
		globals,
		globals,
		make(map[ast.Expr]int),
		nil,
		&ast.Binary{
			Left:     nil,
			Operator: token.NewToken(token.NONE, "", nil, -1),
			Right:    nil,
		},
	}
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) error {
	for _, s := range stmts {
		_, err := i.execute(s)
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

func (i *Interpreter) evaluate(exprNode ast.Expr) (any, error) {
	switch expr := exprNode.(type) {
	case *ast.Assign:
		return i.evalAssign(expr)
	case *ast.Binary:
		return i.evalBinary(expr)
	case *ast.Call:
		return i.evalCall(expr)
	case *ast.Get:
		return i.evalGet(expr)
	case *ast.Grouping:
		return i.evaluate(expr.Expression)
	case *ast.HashLiteral:
		return i.evalHashLiteral(expr)
	case *ast.IndexedGet:
		return i.evalIndexGet(expr)
	case *ast.IndexedSet:
		return i.evalIndexSet(expr)
	case *ast.Set:
		return i.evalSet(expr)
	case *ast.Super:
		return i.evalSuper(expr)
	case *ast.Unary:
		return i.evalUnary(expr)
	case *ast.This:
		return i.lookUpVariable(expr.Keyword, expr)
	case *ast.Variable:
		return i.lookUpVariable(expr.Name, expr)
	case *ast.Lambda:
		return NewUserFn("", expr.Func, i.env), nil
	case *ast.Literal:
		return expr.Value, nil
	case *ast.ArrayLiteral:
		items := make([]any, 0, len(expr.Elements))
		for _, elm := range expr.Elements {
			e, err := i.evaluate(elm)
			if err != nil {
				return nil, err
			}
			items = append(items, e)
		}
		return &LoxArray{items}, nil
	case *ast.Logical:
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
	default:
		panic(fmt.Sprintf("evaluate is unimplemented for '%T'", expr))
	}
}

func (i *Interpreter) execute(stmtNode ast.Stmt) (any, error) {
	var val any
	var err error
	switch stmt := stmtNode.(type) {
	case *ast.Block:
		return i.executeBlock(stmt.Statements, NewEnv(i.env))
	case *ast.Class:
		return i.stmtClass(stmt)
	case *ast.Expression:
		return i.evaluate(stmt.Expression)
	case *ast.If:
		return i.stmtIf(stmt)
	case *ast.While:
		return i.stmtWhile(stmt)
	case *ast.Function:
		name := stmt.Name.Lexeme
		fn := NewUserFn(name, stmt, i.env)
		i.env.Define(name, fn)
		return nil, nil
	case *ast.Print:
		val, err = i.evaluate(stmt.Expression)
		if err != nil {
			return nil, err
		}
		fmt.Println(i.stringify(val))
		return nil, nil
	case *ast.Control:
		if stmt.Keyword.Kind == token.BREAK {
			return nil, BreakErr
		}
		if stmt.Value != nil {
			val, err = i.evaluate(stmt.Value)
			if err != nil {
				return nil, err
			}
		}
		return nil, &ReturnErr{Value: val}
	case *ast.Var:
		if stmt.Initializer != nil {
			val, err = i.evaluate(stmt.Initializer)
			if err != nil {
				return nil, err
			}
		}
		i.env.Define(stmt.Name.Lexeme, val)
		return nil, nil
	default:
		panic(fmt.Sprintf("execute is unimplemented for '%T'", stmt))
	}
}

func (i *Interpreter) stmtWhile(stmt *ast.While) (any, error) {
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

func (i *Interpreter) stmtIf(stmt *ast.If) (any, error) {
	cond, err := i.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if i.isTruthy(cond) {
		return i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		return i.execute(stmt.ElseBranch)
	}
	return nil, nil
}

func (i *Interpreter) stmtClass(stmt *ast.Class) (any, error) {
	var supercls any = nil
	var err error
	if stmt.Superclass != nil {
		supercls, err = i.lookUpVariable(stmt.Superclass.Name, stmt.Superclass)
		if err != nil {
			return nil, err
		}
		if _, ok := supercls.(*UserClass); !ok {
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
	methods := make(map[string]*UserFn)
	for _, method := range stmt.Methods {
		methods[method.Name.Lexeme] = NewUserFn(method.Name.Lexeme, method, i.env)
	}
	scls, _ := supercls.(*UserClass)
	klass := NewUserClass(stmt.Name.Lexeme, scls, methods)
	if supercls != nil {
		i.env = i.env.Enclosing
	}
	err = i.env.Assign(stmt.Name, klass)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) executeBlock(stmts []ast.Stmt, env *Env) (any, error) {
	prv := i.env
	defer func() { i.env = prv }()
	i.env = env
	for _, stmt := range stmts {
		_, err := i.execute(stmt)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) evalAssign(expr *ast.Assign) (any, error) {
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	oprType := token.NONE
	switch expr.Operator.Kind {
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
	if oprType != token.NONE {
		tmp, err := i.lookUpVariable(expr.Name, expr)
		if err != nil {
			return nil, err
		}
		i.tmpBin.Left = &ast.Literal{Value: tmp}
		i.tmpBin.Right = &ast.Literal{Value: val}
		i.tmpBin.Operator.Kind = oprType
		i.tmpBin.Operator.Line = expr.Operator.Line
		val, err = i.evalBinary(i.tmpBin)
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

func (i *Interpreter) evalBinary(expr *ast.Binary) (any, error) {
	lhs, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}
	rhs, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Kind {
	case token.NEQ:
		return !i.isEqual(lhs, rhs), nil
	case token.EQ_EQ:
		return i.isEqual(lhs, rhs), nil
	case token.GT:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l > r, nil
	case token.GT_EQ:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l >= r, nil
	case token.LT:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l < r, nil
	case token.LT_EQ:
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
	case token.STAR:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return l * r, nil
	case token.PERCENT:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		return float64(int(l) % int(r)), nil
	case token.SLASH:
		l, r, err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return nil, err
		}
		if r == 0.0 {
			return nil, &RunTimeErr{Tok: expr.Operator, Msg: "Division by 0"}
		}
		return l / r, nil
	case token.PLUS:
		// looks ugly but is faster as if lhs is not a float/string then
		// we don't have to do check if rhs is a floa/string
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
	}
	// unreachable
	return nil, nil
}

func (i *Interpreter) evalCall(expr *ast.Call) (any, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(expr.Arguments))
	args = append(args, i)
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
		return fn.Call(args...)
	}
	if len(args)-1 != fn.Arity() {
		msg := fmt.Sprintf("Expected %d arguments but got %d", fn.Arity(), len(args)-1)
		return nil, &RunTimeErr{Tok: expr.Paren, Msg: msg}
	}
	return fn.Call(args...)
}

func (i *Interpreter) evalGet(expr *ast.Get) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	if klass, ok := obj.(*UserClass); ok {
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

func (i *Interpreter) evalSet(expr *ast.Set) (any, error) {
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

func (i *Interpreter) evalSuper(expr *ast.Super) (any, error) {
	dist := i.locals[expr]
	superclass := i.env.GetAt(dist, "super").(*UserClass)
	obj := i.env.GetAt(dist-1, "this").(*LoxInstance) // 'this' on super
	method := superclass.FindMethod(expr.Method.Lexeme)
	if method == nil {
		return nil, &RunTimeErr{
			Tok: expr.Method,
			Msg: fmt.Sprintf("Undefined property '%s'", expr.Method.Lexeme),
		}
	}
	return method.Bind(obj), nil
}

func (i *Interpreter) evalUnary(expr *ast.Unary) (any, error) {
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

func (i *Interpreter) evalIndexGet(expr *ast.IndexedGet) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	isRange := expr.Colon != nil
	switch iter := obj.(type) {
	case LoxIterable:
		var start any = 0.0
		var stop any = nil
		if expr.Start != nil {
			start, err = i.evaluate(expr.Start)
			if err != nil {
				return nil, err
			}
		}
		if expr.Stop != nil {
			stop, err = i.evaluate(expr.Stop)
			if err != nil {
				return nil, err
			}
		}
		var val any
		if isRange {
			val, err = iter.IndexRange(start, stop)
		} else {
			val, err = iter.IndexGet(start)
		}
		if err != nil {
			return nil, &RunTimeErr{Tok: expr.Sqr, Msg: err.Error()}
		}
		return val, nil
	case string:
		start := 0
		if expr.Start != nil {
			start, err = i.checkIndex(expr.Sqr, expr.Start, len(iter), "string")
			if err != nil {
				return nil, err
			}
		}
		stop := len(iter)
		if expr.Stop != nil {
			stop, err = i.checkIndex(expr.Sqr, expr.Stop, len(iter), "string")
			if err != nil {
				return nil, err
			}
		}
		if isRange {
			return iter[start:stop], nil
		}
		return string(iter[start]), nil
	default:
		return nil, &RunTimeErr{
			Tok: expr.Sqr,
			Msg: "Can only index an iterable type",
		}
	}
}

func (i *Interpreter) evalIndexSet(expr *ast.IndexedSet) (any, error) {
	tmpErr := &RunTimeErr{Tok: expr.Sqr, Msg: ""}
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}
	iter, ok := obj.(LoxIterable)
	if !ok {
		tmpErr.Msg = "Only iterables can be set using an index"
		return nil, tmpErr
	}
	val, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	idx, err := i.evaluate(expr.Index)
	if err != nil {
		return nil, err
	}
	err = iter.IndexSet(idx, val)
	if err != nil {
		return nil, &RunTimeErr{Tok: expr.Sqr, Msg: fmt.Sprint(err)}
	}
	return val, nil
}

func (i *Interpreter) evalHashLiteral(expr *ast.HashLiteral) (any, error) {
	pairs := make(map[uint]*LoxPair)
	for key, val := range expr.Pairs {
		k, err := i.evaluate(key)
		if err != nil {
			return nil, err
		}
		v, err := i.evaluate(val)
		if err != nil {
			return nil, err
		}
		hash, err := i.hashObj(k, expr.Brace)
		if err != nil {
			return nil, err
		}
		pairs[hash] = &LoxPair{k, v}
	}
	return &LoxHashMap{pairs}, nil
}

func (i *Interpreter) hashObj(obj any, brace *token.Token) (uint, error) {
	hasher := fnv.New64a()
	switch val := obj.(type) {
	case string:
		hasher.Write([]byte(val))
		return uint(hasher.Sum64()), nil
	case float64:
		return uint(val + 1), nil
	case bool:
		if val {
			return 3, nil
		}
		return 5, nil
	case *LoxInstance:
		return val.Hash(), nil
	default:
		return 0, &RunTimeErr{
			Tok: brace,
			Msg: fmt.Sprintf("Unhashable type '%T'", val),
		}
	}
}

func (i *Interpreter) checkIndex(sqr *token.Token, index ast.Expr, cnt int, kind string) (int, error) {
	fdx, err := i.evaluate(index)
	if err != nil {
		return 0, err
	}
	idx, err := i.checkInt(fdx)
	if err != nil {
		return 0, &RunTimeErr{
			Tok: sqr,
			Msg: fmt.Sprintf("Can only use integers to index an %s", kind),
		}
	}
	ogIdx := idx
	// support negative indexes
	if idx < 0 {
		idx = cnt + idx
	}
	if idx >= 0 && idx < cnt {
		return idx, nil
	}
	return 0, &RunTimeErr{
		Tok: sqr,
		Msg: fmt.Sprintf("Index out of bounds. index: %d, length: %d", ogIdx, cnt),
	}
}

func (i *Interpreter) checkInt(val any) (int, error) {
	if fval, ok := val.(float64); ok {
		if fval == float64(int(fval)) {
			return int(fval), nil
		}
	}
	return 0, fmt.Errorf("'%s' is not a integer", val)
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
	return 0, &RunTimeErr{Tok: oprtr, Msg: "Operand must be a number"}
}

func (i *Interpreter) checkNumberOperands(oprtr *token.Token, lhs any, rhs any) (float64, float64, error) {
	if l, lok := lhs.(float64); lok {
		if r, rok := rhs.(float64); rok {
			return l, r, nil
		}
	}
	return 0, 0, &RunTimeErr{Tok: oprtr, Msg: "Operands must be a number"}
}

func (i *Interpreter) reportRunTimeErr(msg error) {
	fmt.Fprintln(os.Stderr, msg)
	i.CurErr = msg
}
