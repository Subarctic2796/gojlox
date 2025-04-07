package interpreter

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

type Interpreter struct {
	ER  errs.ErrorReporter
	env *Env
}

func NewInterpreter(ER errs.ErrorReporter) *Interpreter {
	return &Interpreter{ER, NewEnv()}
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) {
	for _, s := range stmts {
		val := i.execute(s)
		if err, ok := val.(error); ok {
			i.ER.ReportRTErr(err)
			return
		}
	}
}

func (i *Interpreter) execute(stmt ast.Stmt) error {
	err := stmt.Accept(i)
	if err, ok := err.(error); ok {
		return err
	}
	return nil
}

func (i *Interpreter) stringify(obj any) string {
	if obj == nil {
		return "nil"
	}
	if nobj, ok := obj.(float64); ok {
		txt := fmt.Sprint(nobj)
		if strings.HasSuffix(txt, ".0") {
			txt = txt[:len(txt)-2]
		}
		return txt
	}
	return fmt.Sprint(obj)
}

func (i *Interpreter) evaluate(expr ast.Expr) any {
	return expr.Accept(i)
}

func (i *Interpreter) evaluateBlock(statements []ast.Stmt, env *Env) {
	prv := i.env
	i.env = env
	for _, stmt := range statements {
		err := i.execute(stmt)
		if err != nil {
			break
		}
	}
	i.env = prv
}

func (i *Interpreter) VisitWhileStmt(stmt *ast.While) any {
	for i.isTruthy(i.evaluate(stmt.Condition)) {
		err := i.execute(stmt.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) VisitLogicalExpr(expr *ast.Logical) any {
	lhs := i.evaluate(expr.Left)
	if expr.Operator.Kind == token.OR {
		if i.isTruthy(lhs) {
			return lhs
		}
	} else {
		if !i.isTruthy(lhs) {
			return lhs
		}
	}
	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitIfStmt(stmt *ast.If) any {
	if i.isTruthy(i.evaluate(stmt.Condition)) {
		err := i.execute(stmt.ThenBranch)
		if err != nil {
			return err
		}
	} else if stmt.ElseBranch != nil {
		err := i.execute(stmt.ElseBranch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *ast.Block) any {
	i.evaluateBlock(stmt.Statements, NewEnvWithEnclosing(i.env))
	return nil
}

func (i *Interpreter) VisitAssignExpr(expr *ast.Assign) any {
	val := i.evaluate(expr.Value)
	err := i.env.Assign(expr.Name, val)
	if err != nil {
		return err
	}
	return val
}

func (i *Interpreter) VisitVariableExpr(expr *ast.Variable) any {
	val, err := i.env.Get(expr.Name)
	if err != nil {
		return err
	}
	return val
}

func (i *Interpreter) VisitVarStmt(stmt *ast.Var) any {
	var val any
	if stmt.Initializer != nil {
		val = i.evaluate(stmt.Initializer)
	}
	i.env.Define(stmt.Name.Lexeme, val)
	return nil
}

func (i *Interpreter) VisitBinaryExpr(expr *ast.Binary) any {
	lhs := i.evaluate(expr.Left)
	rhs := i.evaluate(expr.Right)

	switch expr.Operator.Kind {
	case token.GREATER:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) > (rhs.(float64))
	case token.GREATER_EQUAL:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) >= (rhs.(float64))
	case token.LESS:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) < (rhs.(float64))
	case token.LESS_EQUAL:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) <= (rhs.(float64))
	case token.MINUS:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) - (rhs.(float64))
	case token.BANG_EQUAL:
		return !i.isEqual(lhs, rhs)
	case token.EQUAL_EQUAL:
		return i.isEqual(lhs, rhs)
	case token.PLUS:
		if l, ok := lhs.(float64); ok {
			if r, ok := rhs.(float64); ok {
				return l + r
			}
		}
		if l, ok := lhs.(string); ok {
			if r, ok := rhs.(string); ok {
				return l + r
			}
		}
		return &errs.RunTimeErr{Tok: expr.Operator, Msg: "Operands must be two numbers or two strings"}
	case token.SLASH:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) / (rhs.(float64))
	case token.STAR:
		err := i.checkNumberOperands(expr.Operator, lhs, rhs)
		if err != nil {
			return err
		}
		return (lhs.(float64)) * (rhs.(float64))
	}
	// unreachable
	return nil
}

func (i *Interpreter) VisitGroupingExpr(expr *ast.Grouping) any {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr *ast.Literal) any {
	return expr.Value
}

func (i *Interpreter) VisitUnaryExpr(expr *ast.Unary) any {
	rhs := i.evaluate(expr.Right)
	switch expr.Operator.Kind {
	case token.MINUS:
		err := i.checkNumberOperand(expr.Operator, rhs)
		if err != nil {
			return err
		}
		return -(rhs.(float64))
	case token.BANG:
		return !i.isTruthy(rhs)
	}

	// unreachable
	return nil
}

func (i *Interpreter) VisitExpressionStmt(stmt *ast.Expression) any {
	val := i.evaluate(stmt.Expression)
	if err, ok := val.(error); ok {
		return err
	}
	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt *ast.Print) any {
	val := i.evaluate(stmt.Expression)
	fmt.Println(i.stringify(val))
	return nil
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

func (i *Interpreter) checkNumberOperand(oprtr *token.Token, opr any) error {
	if _, ok := opr.(float64); ok {
		return nil
	}
	return &errs.RunTimeErr{Tok: oprtr, Msg: "Operand must be a number"}
}

func (i *Interpreter) checkNumberOperands(oprtr *token.Token, lhs any, rhs any) error {
	_, lok := lhs.(float64)
	_, rok := rhs.(float64)
	if lok && rok {
		return nil
	}
	return &errs.RunTimeErr{Tok: oprtr, Msg: "Operands must be a number"}
}
