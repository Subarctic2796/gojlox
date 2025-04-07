// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Expr interface {
	Accept(visitor ExprVisitor) any
}

type ExprVisitor interface {
	VisitAssignExpr(expr *Assign) any
	VisitBinaryExpr(expr *Binary) any
	VisitGroupingExpr(expr *Grouping) any
	VisitLiteralExpr(expr *Literal) any
	VisitLogicalExpr(expr *Logical) any
	VisitUnaryExpr(expr *Unary) any
	VisitVariableExpr(expr *Variable) any
}

type Assign struct {
	Name  *token.Token
	Value Expr
}

func (expr *Assign) Accept(visitor ExprVisitor) any {
	return visitor.VisitAssignExpr(expr)
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Binary) Accept(visitor ExprVisitor) any {
	return visitor.VisitBinaryExpr(expr)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) any {
	return visitor.VisitGroupingExpr(expr)
}

type Literal struct {
	Value any
}

func (expr *Literal) Accept(visitor ExprVisitor) any {
	return visitor.VisitLiteralExpr(expr)
}

type Logical struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (expr *Logical) Accept(visitor ExprVisitor) any {
	return visitor.VisitLogicalExpr(expr)
}

type Unary struct {
	Operator *token.Token
	Right    Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(expr)
}

type Variable struct {
	Name *token.Token
}

func (expr *Variable) Accept(visitor ExprVisitor) any {
	return visitor.VisitVariableExpr(expr)
}
