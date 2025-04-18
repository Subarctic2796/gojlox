package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

const PATH = "/home/benji/Coding/compilers/gojlox/ast"

func main() {
	defineAst(PATH, "Expr", []string{
		"Assign   : Name *token.Token, Operator *token.Token, Value Expr",
		"Binary   : Left Expr, Operator *token.Token, Right Expr",
		"Call     : Callee Expr, Paren *token.Token, Arguments []Expr",
		"Get      : Object Expr, Name *token.Token",
		"Lambda   : Params []*token.Token, Body []Stmt",
		"Grouping : Expression Expr",
		"Literal  : Value any",
		"Logical  : Left Expr, Operator *token.Token, Right Expr",
		"Set      : Object Expr, Name *token.Token, Kind *token.Token, Value Expr",
		"Super    : Keyword *token.Token, Method *token.Token",
		"This     : Keyword *token.Token",
		"Unary    : Operator *token.Token, Right Expr",
		"Variable : Name *token.Token",
	})

	defineAst(PATH, "Stmt", []string{
		"Block      : Statements []Stmt",
		"Break      : ",
		"Class      : Name *token.Token, Superclass *Variable, Methods []*Function, Statics []*Function",
		"Expression : Expression Expr",
		"Function   : Name *token.Token, Func *Lambda",
		"If         : Condition Expr, ThenBranch Stmt, ElseBranch Stmt",
		"Print      : Expression Expr",
		"Return     : Keyword *token.Token, Value Expr",
		"Var        : Name *token.Token, Initializer Expr",
		"While      : Condition Expr, Body Stmt",
	})
}

func defineAst(path, baseName string, types []string) {
	tmpl, err := template.New("ast").Parse(TMPL)
	if err != nil {
		panic(err)
	}
	classes := make(map[string][]string)
	for _, t := range types {
		ls := strings.Split(t, ":")
		name, fstr := strings.TrimSpace(ls[0]), strings.TrimSpace(ls[1])
		if len(fstr) == 0 {
			classes[name] = []string{}
		} else {
			classes[name] = strings.Split(fstr, ", ")
		}
	}

	bnLower := strings.ToLower(baseName)
	var sb strings.Builder
	err = tmpl.Execute(&sb, map[string]any{
		"BN":      baseName,
		"bn":      bnLower,
		"Classes": classes,
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Create(fmt.Sprintf("%s/%s.go", path, bnLower))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(sb.String())
}

const TMPL = `// GENERATED CODE DO NOT EDIT
package ast

import "github.com/Subarctic2796/gojlox/token"

type {{.BN}} interface {
    Accept(visitor {{.BN}}Visitor) (any, error)
}

type {{.BN}}Visitor interface {
    {{range $name, $fields := .Classes -}}
    Visit{{$name}}{{$.BN}}({{$.bn}} *{{$name}}) (any, error)
    {{end}}
}
{{range $name, $fields := .Classes}}
type {{$name}} struct {
    {{- range $field := $fields}}
    {{$field -}}
    {{end}}
}

func ({{$.bn}} *{{$name}}) Accept(visitor {{$.BN}}Visitor) (any, error) {
    return visitor.Visit{{$name}}{{$.BN}}({{$.bn}})
}
{{end}}`
