// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type StmtVisitor interface {
	VisitBlockStmt(stmt *Block) (any, error)
	VisitBreakStmt(stmt *Break) (any, error)
	VisitClassStmt(stmt *Class) (any, error)
	VisitExpressionStmt(stmt *Expression) (any, error)
	VisitFunctionStmt(stmt *Function) (any, error)
	VisitIfStmt(stmt *If) (any, error)
	VisitPrintStmt(stmt *Print) (any, error)
	VisitReturnStmt(stmt *Return) (any, error)
	VisitVarStmt(stmt *Var) (any, error)
	VisitWhileStmt(stmt *While) (any, error)
}

type Block struct {
	Statements []Stmt
}

func (stmt *Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(stmt)
}

type Break struct{}

func (stmt *Break) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBreakStmt(stmt)
}

type Class struct {
	Name       *token.Token
	Superclass *Variable
	Methods    []*Function
}

func (stmt *Class) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitClassStmt(stmt)
}

type Expression struct {
	Expression Expr
}

func (stmt *Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(stmt)
}

type Function struct {
	Name   *token.Token
	Params []*token.Token
	Body   []Stmt
}

func (stmt *Function) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitFunctionStmt(stmt)
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (stmt *If) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitIfStmt(stmt)
}

type Print struct {
	Expression Expr
}

func (stmt *Print) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(stmt)
}

type Return struct {
	Keyword *token.Token
	Value   Expr
}

func (stmt *Return) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitReturnStmt(stmt)
}

type Var struct {
	Name        *token.Token
	Initializer Expr
}

func (stmt *Var) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitVarStmt(stmt)
}

type While struct {
	Condition Expr
	Body      Stmt
}

func (stmt *While) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitWhileStmt(stmt)
}
