package ast

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/token"
)

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
	String() string
}

type ExprVisitor interface {
	VisitAssignExpr(expr *Assign) (any, error)
	VisitBinaryExpr(expr *Binary) (any, error)
	VisitCallExpr(expr *Call) (any, error)
	VisitGetExpr(expr *Get) (any, error)
	VisitGroupingExpr(expr *Grouping) (any, error)
	VisitLambdaExpr(expr *Lambda) (any, error)
	VisitLiteralExpr(expr *Literal) (any, error)
	VisitLogicalExpr(expr *Logical) (any, error)
	VisitSetExpr(expr *Set) (any, error)
	VisitSuperExpr(expr *Super) (any, error)
	VisitThisExpr(expr *This) (any, error)
	VisitUnaryExpr(expr *Unary) (any, error)
	VisitVariableExpr(expr *Variable) (any, error)
}

type Assign struct {
	Name     *token.Token
	Operator *token.Token
	Value    Expr
}

func (expr *Assign) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignExpr(expr)
}

func (expr *Assign) String() string {
	sopr := expr.Operator.Lexeme
	sname := expr.Name.Lexeme
	return fmt.Sprintf("(%s %s %s", sopr, sname, expr.Value)
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(expr)
}

func (expr *Binary) String() string {
	sopr := expr.Operator.Lexeme
	return fmt.Sprintf("(%s %s %s)", sopr, expr.Left, expr.Right)
}

type Call struct {
	Callee    Expr
	Paren     *token.Token
	Arguments []Expr
}

func (expr *Call) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(expr)
}

func (expr *Call) String() string {
	var sb strings.Builder
	for _, arg := range expr.Arguments {
		sb.WriteString(" ")
		sb.WriteString(arg.String())
	}
	return fmt.Sprintf("(call %s%s", expr.Callee, sb.String())
}

type Get struct {
	Object Expr
	Name   *token.Token
}

func (expr *Get) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGetExpr(expr)
}

func (expr *Get) String() string {
	return fmt.Sprintf("(. %s %s)", expr.Object, expr.Name.Lexeme)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(expr)
}

func (expr *Grouping) String() string {
	return fmt.Sprintf("(group %s)", expr.Expression)
}

type Lambda struct {
	Params []*token.Token
	Body   []Stmt
	Kind   FnType
	// Body *Function
}

func (expr *Lambda) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLambdaExpr(expr)
}

func (expr *Lambda) String() string {
	// return fmt.Sprintf("(fun(%s))", expr.Name.Lexeme, expr.Func)
	var sb strings.Builder
	if expr.Kind == FN_LAMBDA {
		sb.WriteString("(fun(")
	}
	for _, param := range expr.Params {
		if param != expr.Params[0] {
			sb.WriteByte(' ')
		}
		sb.WriteString(param.Lexeme)
	}
	sb.WriteString(") ")
	for _, s := range expr.Body {
		sb.WriteString(s.String())
	}
	sb.WriteString(")")
	return sb.String()
}

type Literal struct {
	Value any
}

func (expr *Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(expr)
}

func (expr *Literal) String() string {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprint(expr.Value)
}

type Logical struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Logical) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(expr)
}

func (expr *Logical) String() string {
	sopr := expr.Operator.Lexeme
	return fmt.Sprintf("(%s %s %s)", sopr, expr.Left, expr.Right)
}

type Set struct {
	Object Expr
	Name   *token.Token
	Kind   *token.Token
	Value  Expr
}

func (expr *Set) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSetExpr(expr)
}

func (expr *Set) String() string {
	sname := expr.Name.Lexeme
	return fmt.Sprintf("(= %s %s %s)", expr.Object, sname, expr.Value)
}

type Super struct {
	Keyword *token.Token
	Method  *token.Token
}

func (expr *Super) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSuperExpr(expr)
}

func (expr *Super) String() string {
	return fmt.Sprintf("(super %s", expr.Method.Lexeme)
}

type This struct {
	Keyword *token.Token
}

func (expr *This) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitThisExpr(expr)
}

func (expr *This) String() string {
	return "this"
}

type Unary struct {
	Operator *token.Token
	Right    Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(expr)
}

func (expr *Unary) String() string {
	return fmt.Sprintf("(%s %s)", expr.Operator.Lexeme, expr.Right)
}

type Variable struct {
	Name *token.Token
}

func (expr *Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(expr)
}

func (expr *Variable) String() string {
	return expr.Name.Lexeme
}
