// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Expr interface {
	Accept(visitor ExprVisitor) any
}

type ExprVisitor interface {
	VisitBinaryExpr(expr *Binary) any
	VisitGroupingExpr(expr *Grouping) any
	VisitLiteralExpr(expr *Literal) any
	VisitUnaryExpr(expr *Unary) any
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

type Unary struct {
	Operator *token.Token
	Right    Expr
}

func (expr *Unary) Accept(visitor ExprVisitor) any {
	return visitor.VisitUnaryExpr(expr)
}
