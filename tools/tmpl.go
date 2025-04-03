package main

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
