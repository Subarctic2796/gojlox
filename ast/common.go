package ast

//go:generate go tool stringer -type=FnType -output=common_strings.go -trimprefix=FN_

type FnType int

const (
	FN_NONE FnType = iota
	FN_NATIVE
	FN_LAMBDA
	FN_FUNC
	FN_INIT
	FN_METHOD
	FN_STATIC
)
