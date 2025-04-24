package ast

//go:generate go tool stringer -type=FnType,ControlType -output=common_strings.go

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

type ControlType int

const (
	CNTRL_NONE ControlType = iota
	CNTRL_BREAK
	CNTRL_RETURN
)
