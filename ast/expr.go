package ast

import (
	"fmt"
	"strings"

	"github.com/Subarctic2796/gojlox/token"
)

type Expr interface {
	String() string
}

type ArrayLiteral struct {
	Sqr      *token.Token
	Elements []Expr
}

func (expr *ArrayLiteral) String() string {
	var sb strings.Builder
	sb.WriteString("([\n")
	for _, elm := range expr.Elements {
		sb.WriteString(fmt.Sprintf("    %s,\n", elm))
	}
	sb.WriteString("])")
	return sb.String()
}

type Assign struct {
	Name     *token.Token
	Operator *token.Token
	Value    Expr
}

func (expr *Assign) String() string {
	sopr := expr.Operator.Lexeme
	sname := expr.Name.Lexeme
	return fmt.Sprintf("(%s %s %s)", sopr, sname, expr.Value)
}

type Binary struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
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

func (expr *Call) String() string {
	var sb strings.Builder
	for _, arg := range expr.Arguments {
		sb.WriteString(" ")
		sb.WriteString(arg.String())
	}
	return fmt.Sprintf("(call %s%s", expr.Callee, sb.String())
}

type IndexedGet struct {
	Object Expr
	Sqr    *token.Token
	Index  Expr
}

func (expr *IndexedGet) String() string {
	return fmt.Sprintf("(%s[%s])", expr.Object, expr.Index)
}

type Get struct {
	Object Expr
	Name   *token.Token
}

func (expr *Get) String() string {
	return fmt.Sprintf("(. %s %s)", expr.Object, expr.Name.Lexeme)
}

type Grouping struct {
	Expression Expr
}

func (expr *Grouping) String() string {
	return fmt.Sprintf("(group %s)", expr.Expression)
}

type HashLiteral struct {
	Brace *token.Token
	Pairs map[Expr]Expr
}

func (expr *HashLiteral) String() string {
	var sb strings.Builder
	sb.WriteString("({\n")
	for k, v := range expr.Pairs {
		sb.WriteString(fmt.Sprintf("   %s: %s,\n", k, v))
	}
	sb.WriteString("})")
	return sb.String()
}

type Lambda struct {
	Func *Function
}

func (expr *Lambda) String() string {
	return fmt.Sprint(expr.Func)
	// var sb strings.Builder
	// if expr.Kind == FN_LAMBDA {
	// 	sb.WriteString("(fun(")
	// }
	// for _, param := range expr.Params {
	// 	if param != expr.Params[0] {
	// 		sb.WriteByte(' ')
	// 	}
	// 	sb.WriteString(param.Lexeme)
	// }
	// sb.WriteString(") ")
	// for _, s := range expr.Body {
	// 	sb.WriteString(s.String())
	// }
	// sb.WriteString(")")
	// return sb.String()
}

type Literal struct {
	Value any
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

func (expr *Logical) String() string {
	sopr := expr.Operator.Lexeme
	return fmt.Sprintf("(%s %s %s)", sopr, expr.Left, expr.Right)
}

type Set struct {
	Object Expr
	Name   *token.Token
	Value  Expr
}

func (expr *Set) String() string {
	sname := expr.Name.Lexeme
	return fmt.Sprintf("(= %s %s %s)", expr.Object, sname, expr.Value)
}

type IndexedSet struct {
	Object Expr
	Sqr    *token.Token
	Index  Expr
	Value  Expr
}

func (expr *IndexedSet) String() string {
	return fmt.Sprintf("(%s[%s] = %s)", expr.Object, expr.Index, expr.Value)
}

type Super struct {
	Keyword *token.Token
	Method  *token.Token
}

func (expr *Super) String() string {
	return fmt.Sprintf("(super %s)", expr.Method.Lexeme)
}

type This struct {
	Keyword *token.Token
}

func (expr *This) String() string {
	return "(this)"
}

type Unary struct {
	Operator *token.Token
	Right    Expr
}

func (expr *Unary) String() string {
	return fmt.Sprintf("(%s %s)", expr.Operator.Lexeme, expr.Right)
}

type Variable struct {
	Name *token.Token
}

func (expr *Variable) String() string {
	return expr.Name.Lexeme
}
