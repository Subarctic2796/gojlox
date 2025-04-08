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
		"Assign   : Token name, Expr value",
		"Binary   : Expr left, Token operator, Expr right",
		"Grouping : Expr expression",
		"Literal  : Object value",
		"Logical  : Expr left, Token operator, Expr right",
		"Unary    : Token operator, Expr right",
		"Variable : Token name",
	})

	defineAst(PATH, "Stmt", []string{
		"Block      : List<Stmt> statements",
		"Expression : Expr expression",
		"If         : Expr condition, Stmt thenBranch, Stmt elseBranch",
		"Print      : Expr expression",
		"Var        : Token name, Expr initializer",
		"While      : Expr condition, Stmt body",
	})
}

func defineAst(PATH, baseName string, types []string) {
	tmpl, err := template.New("ast").Parse(TMPL)
	if err != nil {
		panic(err)
	}
	realTypes := map[string]string{
		"Token":  "*token.Token",
		"Object": "any",
		"List<":  "[]",
		">":      "",
	}
	classes := make(map[string][]string)
	for _, t := range types {
		ls := strings.Split(t, ":")
		name, fstr := strings.TrimSpace(ls[0]), strings.TrimSpace(ls[1])
		for olds, news := range realTypes {
			fstr = strings.ReplaceAll(fstr, olds, news)
		}
		fields := strings.Split(fstr, ", ")
		for i, f := range fields {
			tmp := strings.Split(f, " ")
			fields[i] = fmt.Sprintf("%s %s", strings.Title(tmp[1]), tmp[0])
		}
		classes[name] = fields
	}

	bnLower := strings.ToLower(baseName)
	var sb strings.Builder
	err = tmpl.Execute(&sb, map[string]any{
		"BN":      baseName,
		"bn":      bnLower,
		"Classes": classes,
	})
	f, err := os.Create(fmt.Sprintf("%s/%s.go", PATH, bnLower))
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
    Accept(visitor {{.BN}}Visitor) any
}

type {{.BN}}Visitor interface {
    {{range $name, $fields := .Classes -}}
    Visit{{$name}}{{$.BN}}({{$.bn}} *{{$name}}) any
    {{end}}
}
{{range $name, $fields := .Classes}}
type {{$name}} struct {
    {{- range $field := $fields}}
    {{$field -}}
    {{end}}
}

func ({{$.bn}} *{{$name}}) Accept(visitor {{$.BN}}Visitor) any {
    return visitor.Visit{{$name}}{{$.BN}}({{$.bn}})
}
{{end}}`
