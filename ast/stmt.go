// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type Stmt interface {
	Accept(visitor StmtVisitor) any
}

type StmtVisitor interface {
	VisitBlockStmt(stmt *Block) any
	VisitExpressionStmt(stmt *Expression) any
	VisitIfStmt(stmt *If) any
	VisitPrintStmt(stmt *Print) any
	VisitVarStmt(stmt *Var) any
	VisitWhileStmt(stmt *While) any
}

type Block struct {
	Statements []Stmt
}

func (stmt *Block) Accept(visitor StmtVisitor) any {
	return visitor.VisitBlockStmt(stmt)
}

type Expression struct {
	Expression Expr
}

func (stmt *Expression) Accept(visitor StmtVisitor) any {
	return visitor.VisitExpressionStmt(stmt)
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (stmt *If) Accept(visitor StmtVisitor) any {
	return visitor.VisitIfStmt(stmt)
}

type Print struct {
	Expression Expr
}

func (stmt *Print) Accept(visitor StmtVisitor) any {
	return visitor.VisitPrintStmt(stmt)
}

type Var struct {
	Name        *token.Token
	Initializer Expr
}

func (stmt *Var) Accept(visitor StmtVisitor) any {
	return visitor.VisitVarStmt(stmt)
}

type While struct {
	Condition Expr
	Body      Stmt
}

func (stmt *While) Accept(visitor StmtVisitor) any {
	return visitor.VisitWhileStmt(stmt)
}
