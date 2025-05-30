package ast

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/token"
)

type Stmt interface {
	String() string
}

type Block struct {
	Statements []Stmt
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

func (stmt *Expression) String() string {
	return fmt.Sprintf("(; %s)", stmt.Expression)
}

type Function struct {
	Name *token.Token
	// Func *Lambda
	Params []*token.Token
	Body   []Stmt
	Kind   FnType
}

func (stmt *Function) String() string {
	// return fmt.Sprintf("(fun %s(%s", stmt.Name.Lexeme, stmt.Func)
	var sb strings.Builder
	if stmt.Kind == FN_LAMBDA {
		sb.WriteString("(fun(")
	} else {
		sb.WriteString(fmt.Sprintf("(fun %s(", stmt.Name.Lexeme))
	}
	for _, param := range stmt.Params {
		if param != stmt.Params[0] {
			sb.WriteByte(' ')
		}
		sb.WriteString(param.Lexeme)
	}
	sb.WriteString(") ")
	for _, s := range stmt.Body {
		sb.WriteString(s.String())
	}
	sb.WriteString(")")
	return sb.String()
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
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

func (stmt *Print) String() string {
	return fmt.Sprintf("(print %s", stmt.Expression)
}

type Control struct {
	Keyword *token.Token
	Value   Expr
}

func (stmt *Control) String() string {
	switch stmt.Keyword.Kind {
	case token.RETURN:
		if stmt.Value == nil {
			return "(return)"
		}
		return fmt.Sprintf("(return %s)", stmt.Value)
	case token.BREAK:
		return "(break)"
	default:
		panic("unreachable")
	}
}

type Var struct {
	Name        *token.Token
	Initializer Expr
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

func (stmt *While) String() string {
	return fmt.Sprintf("(while %s %s)", stmt.Condition, stmt.Body)
}
