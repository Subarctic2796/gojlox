// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
}

type ExprVisitor interface {
	VisitAssignExpr(expr *Assign) (any, error)
	VisitBinaryExpr(expr *Binary) (any, error)
	VisitCallExpr(expr *Call) (any, error)
	VisitGroupingExpr(expr *Grouping) (any, error)
	VisitLiteralExpr(expr *Literal) (any, error)
	VisitLogicalExpr(expr *Logical) (any, error)
	VisitUnaryExpr(expr *Unary) (any, error)
	VisitVariableExpr(expr *Variable) (any, error)
}

type Assign struct {
	Name  *token.Token
	Value Expr
}

func (expr *Assign) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignExpr(expr)
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(expr)
}

type Call struct {
	Callee    Expr
	Paren     *token.Token
	Arguments []Expr
}

func (expr *Call) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(expr)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(expr)
}

type Literal struct {
	Value any
}

func (expr *Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(expr)
}

type Logical struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Logical) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(expr)
}

type Unary struct {
	Operator *token.Token
	Right    Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(expr)
}

type Variable struct {
	Name *token.Token
}

func (expr *Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(expr)
}
