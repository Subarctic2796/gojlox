package interpreter

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

type Interpreter struct {
	ER errs.ErrorReporter
}

func NewInterpreter(ER errs.ErrorReporter) *Interpreter {
	return &Interpreter{ER}
}

func (i *Interpreter) Interpret(expr ast.Expr) {
	val := i.evaluate(expr)
	if err, ok := val.(error); ok {
		i.ER.ReportRTErr(err)
	} else {
		fmt.Println(i.stringify(val))
	}
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
