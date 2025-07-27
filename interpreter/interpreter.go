package interpreter

import (
	"errors"
	"fmt"
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
	tok := token.NewToken(token.NONE, "", nil, -1)
	return &Interpreter{
		globals,
		globals,
		make(map[ast.Expr]int),
		nil,
		&ast.Binary{
			Left:     nil,
			Operator: &tok,
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

func (i *Interpreter) evaluate(expr ast.Expr) (any, error) {
	switch e := expr.(type) {
	case *ast.Assign:
		return i.evalAssign(e)
	case *ast.Binary:
		return i.evalBinary(e)
	case *ast.Grouping:
		return i.evaluate(e.Expression)
	case *ast.IndexedGet:
		return i.evalIndexGet(e)
	case *ast.This:
		return i.lookUpVariable(e.Keyword, e)
	case *ast.Variable:
		return i.lookUpVariable(e.Name, e)
	case *ast.Lambda:
		return NewUserFn("", e.Func, i.env), nil
	case *ast.Literal:
		return e.Value, nil
	case *ast.Call:
		callee, err := i.evaluate(e.Callee)
		if err != nil {
			return nil, err
		}
		args := make([]any, 0, len(e.Arguments))
		args = append(args, i)
		for _, arg := range e.Arguments {
			a, err := i.evaluate(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, a)
		}

		fn, ok := callee.(LoxCallable)
		if !ok {
			return nil, &RunTimeErr{
				Tok: e.Paren,
				Msg: "Can only call functions and classes",
			}
		}
		if fn.Arity() == -1 {
			return fn.Call(args...)
		}
		if len(args)-1 != fn.Arity() {
			msg := fmt.Sprintf("Expected %d arguments but got %d", fn.Arity(), len(args)-1)
			return nil, &RunTimeErr{Tok: e.Paren, Msg: msg}
		}
		return fn.Call(args...)
	case *ast.Get:
		obj, err := i.evaluate(e.Object)
		if err != nil {
			return nil, err
		}
		if klass, ok := obj.(*UserClass); ok {
			static := klass.FindMethod(e.Name.Lexeme)
			if static != nil {
				if static.Func.Kind != ast.FN_STATIC {
					return nil, &RunTimeErr{
						Tok: e.Name,
						Msg: fmt.Sprintf("Undefined static function '%s'", e.Name.Lexeme),
					}
				}
				return static, nil
			}
		}
		if inst, ok := obj.(*LoxInstance); ok {
			return inst.Get(e.Name)
		}
		return nil, &RunTimeErr{
			Tok: e.Name,
			Msg: "Only instances have properties",
		}
	case *ast.HashLiteral:
		pairs := make(map[any]any)
		for key, val := range e.Pairs {
			k, err := i.evaluate(key)
			if err != nil {
				return nil, err
			}
			v, err := i.evaluate(val)
			if err != nil {
				return nil, err
			}
			err = i.hashable(k, e.Brace)
			if err != nil {
				return nil, err
			}
			pairs[k] = v
		}
		return &LoxHashMap{pairs}, nil
	case *ast.IndexedSet:
		tmpErr := &RunTimeErr{Tok: e.Sqr, Msg: ""}
		obj, err := i.evaluate(e.Object)
		if err != nil {
			return nil, err
		}
		iter, ok := obj.(LoxIterable)
		if !ok {
			tmpErr.Msg = "Only iterables can be set using an index"
			return nil, tmpErr
		}
		val, err := i.evaluate(e.Value)
		if err != nil {
			return nil, err
		}
		idx, err := i.evaluate(e.Index)
		if err != nil {
			return nil, err
		}
		err = iter.IndexSet(idx, val)
		if err != nil {
			return nil, &RunTimeErr{Tok: e.Sqr, Msg: fmt.Sprint(err)}
		}
		return val, nil
	case *ast.Set:
		obj, err := i.evaluate(e.Object)
		if err != nil {
			return nil, err
		}
		inst, ok := obj.(*LoxInstance)
		if !ok {
			return nil, &RunTimeErr{
				Tok: e.Name,
				Msg: "Only instances have fields",
			}
		}
		val, err := i.evaluate(e.Value)
		if err != nil {
			return nil, err
		}
		inst.Set(e.Name, val)
		return val, nil
	case *ast.Super:
		dist := i.locals[e]
		superclass := i.env.GetAt(dist, "super").(*UserClass)
		obj := i.env.GetAt(dist-1, "this").(*LoxInstance) // 'this' on super
		method := superclass.FindMethod(e.Method.Lexeme)
		if method == nil {
			return nil, &RunTimeErr{
				Tok: e.Method,
				Msg: fmt.Sprintf("Undefined property '%s'", e.Method.Lexeme),
			}
		}
		return method.Bind(obj), nil
	case *ast.Unary:
		rhs, err := i.evaluate(e.Right)
		if err != nil {
			return nil, err
		}
		switch e.Operator.Kind {
		case token.MINUS:
			r, err := i.checkNumberOperand(e.Operator, rhs)
			if err != nil {
				return nil, err
			}
			return -r, nil
		case token.BANG:
			return !i.isTruthy(rhs), nil
		}
		// unreachable
		return nil, nil
	case *ast.ArrayLiteral:
		items := make([]any, 0, len(e.Elements))
		for _, elm := range e.Elements {
			e, err := i.evaluate(elm)
			if err != nil {
				return nil, err
			}
			items = append(items, e)
		}
		return &LoxArray{items}, nil
	case *ast.Logical:
		lhs, err := i.evaluate(e.Left)
		if err != nil {
			return nil, err
		}
		if e.Operator.Kind == token.OR {
			if i.isTruthy(lhs) {
				return lhs, nil
			}
		} else {
			if !i.isTruthy(lhs) {
				return lhs, nil
			}
		}
		return i.evaluate(e.Right)
	default:
		panic(fmt.Sprintf("evaluate is unimplemented for '%T'", e))
	}
}

func (i *Interpreter) execute(stmt ast.Stmt) (any, error) {
	var val any
	var err error
	switch s := stmt.(type) {
	case *ast.Block:
		return i.executeBlock(s.Statements, NewEnv(i.env))
	case *ast.Class:
		var supercls any = nil
		var err error
		if s.Superclass != nil {
			supercls, err = i.lookUpVariable(s.Superclass.Name, s.Superclass)
			if err != nil {
				return nil, err
			}
			if _, ok := supercls.(*UserClass); !ok {
				return nil, &RunTimeErr{
					Tok: s.Superclass.Name,
					Msg: "Superclass must be a class",
				}
			}
		}
		i.env.Define(s.Name.Lexeme, nil)
		if s.Superclass != nil {
			i.env = NewEnv(i.env)
			i.env.Define("super", supercls)
		}
		methods := make(map[string]*UserFn)
		for _, method := range s.Methods {
			methods[method.Name.Lexeme] = NewUserFn(method.Name.Lexeme, method, i.env)
		}
		scls, _ := supercls.(*UserClass)
		klass := NewUserClass(s.Name.Lexeme, scls, methods)
		if supercls != nil {
			i.env = i.env.Enclosing
		}
		err = i.env.Assign(s.Name, klass)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case *ast.Expression:
		return i.evaluate(s.Expression)
	case *ast.If:
		cond, err := i.evaluate(s.Condition)
		if err != nil {
			return nil, err
		}
		if i.isTruthy(cond) {
			return i.execute(s.ThenBranch)
		} else if s.ElseBranch != nil {
			return i.execute(s.ElseBranch)
		}
		return nil, nil
	case *ast.While:
		cond, err := i.evaluate(s.Condition)
		if err != nil {
			return nil, err
		}
		for i.isTruthy(cond) {
			_, err = i.execute(s.Body)
			if err != nil {
				if errors.Is(err, BreakErr) {
					return nil, nil
				}
				return nil, err
			}
			cond, err = i.evaluate(s.Condition)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *ast.Function:
		name := s.Name.Lexeme
		i.env.Define(name, NewUserFn(name, s, i.env))
		return nil, nil
	case *ast.Print:
		val, err = i.evaluate(s.Expression)
		if err != nil {
			return nil, err
		}
		fmt.Println(i.stringify(val))
		return nil, nil
	case *ast.Control:
		if s.Keyword.Kind == token.BREAK {
			return nil, BreakErr
		}
		if s.Value != nil {
			val, err = i.evaluate(s.Value)
			if err != nil {
				return nil, err
			}
		}
		return nil, &ReturnErr{Value: val}
	case *ast.Var:
		if s.Initializer != nil {
			val, err = i.evaluate(s.Initializer)
			if err != nil {
				return nil, err
			}
		}
		i.env.Define(s.Name.Lexeme, val)
		return nil, nil
	default:
		panic(fmt.Sprintf("execute is unimplemented for '%T'", s))
	}
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

func (i *Interpreter) hashable(obj any, brace *token.Token) error {
	switch val := obj.(type) {
	case string:
		return nil
	case float64:
		return nil
	case bool:
		return nil
	case nil:
		return nil
	case *LoxInstance:
		return nil
	default:
		return &RunTimeErr{
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
