package ast

type FnType int

const (
	FN_NONE FnType = iota
	FN_LAMBDA
	FN_FUNC
	FN_INIT
	FN_METHOD
	FN_STATIC
)
