package ast

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/token"
)

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
	String() string
}

type StmtVisitor interface {
	VisitBlockStmt(stmt *Block) (any, error)
	VisitClassStmt(stmt *Class) (any, error)
	VisitExpressionStmt(stmt *Expression) (any, error)
	VisitFunctionStmt(stmt *Function) (any, error)
	VisitIfStmt(stmt *If) (any, error)
	VisitPrintStmt(stmt *Print) (any, error)
	VisitControlStmt(stmt *Control) (any, error)
	VisitVarStmt(stmt *Var) (any, error)
	VisitWhileStmt(stmt *While) (any, error)
}

type Block struct {
	Statements []Stmt
}

func (stmt *Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(stmt)
}

func (stmt *Block) String() string {
	var sb strings.Builder
	sb.WriteString("(block ")
	for _, s := range stmt.Statements {
		sb.WriteString(s.String())
	}
	sb.WriteByte(')')
	return sb.String()
}

type Class struct {
	Name       *token.Token
	Superclass *Variable
	Methods    []*Function
}

func (stmt *Class) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitClassStmt(stmt)
}

func (stmt *Class) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(class %s", stmt.Name.Lexeme))
	if stmt.Superclass != nil {
		sb.WriteString(fmt.Sprintf(" < %s", stmt.Superclass))
	}
	for _, fn := range stmt.Methods {
		sb.WriteString(fmt.Sprintf(" %s", fn))
	}
	sb.WriteByte(')')
	return sb.String()
}

type Expression struct {
	Expression Expr
}

func (stmt *Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(stmt)
}

func (stmt *Expression) String() string {
	return fmt.Sprintf("(; %s)", stmt.Expression)
}

type Function struct {
	Name *token.Token
	Func *Lambda
	/* Params []*token.Token
	Body   []Stmt
	Kind   FnType */
}

func (stmt *Function) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitFunctionStmt(stmt)
}

func (stmt *Function) String() string {
	return fmt.Sprintf("(fun %s(%s", stmt.Name.Lexeme, stmt.Func)
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (stmt *If) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitIfStmt(stmt)
}

func (stmt *If) String() string {
	if stmt.ElseBranch != nil {
		return fmt.Sprintf("(if %s)", stmt.Condition)
	}
	return fmt.Sprintf("(if-else %s %s)", stmt.Condition, stmt.ElseBranch)
}

type Print struct {
	Expression Expr
}

func (stmt *Print) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(stmt)
}

func (stmt *Print) String() string {
	return fmt.Sprintf("(print %s", stmt.Expression)
}

type Control struct {
	Kind    ControlType
	Keyword *token.Token
	Value   Expr
}

func (stmt *Control) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitControlStmt(stmt)
}

func (stmt *Control) String() string {
	switch stmt.Kind {
	case CNTRL_RETURN:
		if stmt.Value == nil {
			return "(return)"
		}
		return fmt.Sprintf("(return %s)", stmt.Value)
	case CNTRL_BREAK:
		return "(break)"
	default:
		panic("unreachable")
	}
}

type Var struct {
	Name        *token.Token
	Initializer Expr
}

func (stmt *Var) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitVarStmt(stmt)
}

func (stmt *Var) String() string {
	sname := stmt.Name.Lexeme
	if stmt.Initializer == nil {
		return fmt.Sprintf("(var %s)", sname)
	}
	return fmt.Sprintf("(var %s = %s)", sname, stmt.Initializer)
}

type While struct {
	Condition Expr
	Body      Stmt
}

func (stmt *While) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitWhileStmt(stmt)
}

func (stmt *While) String() string {
	return fmt.Sprintf("(while %s %s)", stmt.Condition, stmt.Body)
}
