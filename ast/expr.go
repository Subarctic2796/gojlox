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

type Get struct {
	Object Expr
	Name   *token.Token
}

func (expr *Get) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGetExpr(expr)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(expr)
}

type Lambda struct {
	Params []*token.Token
	Body   []Stmt
	Kind   FnType
}

func (expr *Lambda) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLambdaExpr(expr)
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

type Set struct {
	Object Expr
	Name   *token.Token
	Kind   *token.Token
	Value  Expr
}

func (expr *Set) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSetExpr(expr)
}

type Super struct {
	Keyword *token.Token
	Method  *token.Token
}

func (expr *Super) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSuperExpr(expr)
}

type This struct {
	Keyword *token.Token
}

func (expr *This) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitThisExpr(expr)
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
