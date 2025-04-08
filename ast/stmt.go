// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type StmtVisitor interface {
	VisitBlockStmt(stmt *Block) (any, error)
	VisitExpressionStmt(stmt *Expression) (any, error)
	VisitIfStmt(stmt *If) (any, error)
	VisitPrintStmt(stmt *Print) (any, error)
	VisitVarStmt(stmt *Var) (any, error)
	VisitWhileStmt(stmt *While) (any, error)
}

type Block struct {
	Statements []Stmt
}

func (stmt *Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(stmt)
}

type Expression struct {
	Expression Expr
}

func (stmt *Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(stmt)
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
