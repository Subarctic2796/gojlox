package errs

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/token"
)

type ErrorReporter interface {
	ReportErr(line int, msg error)
	ReportTok(tok *token.Token, msg error)
	Report(line int, where string, msg error)
	ReportRTErr(msg error)
}

var (
	ErrUnexpectedChar      = errors.New("Unexpected character")
	ErrUnterminatedStr     = errors.New("Unterminated string")
	ErrUnterminatedComment = errors.New("Unterminated comment")
	ErrParse               = errors.New("Parser Error")
)

type RunTimeErr struct {
	Tok *token.Token
	Msg string
}

func (e *RunTimeErr) Error() string {
	return fmt.Sprintf("[RunTimeError] %s: %s", e.Tok, e.Msg)
}

type ResolverErrMsg string

const (
	AlreadyInScope            ResolverErrMsg = "Already a variable with this name in this scope"
	ReadLocalInOwnInitializer ResolverErrMsg = "Can't read local variable in its own initializer"
	ReturnTopLevel            ResolverErrMsg = "Can't return from top-level code"
	ReturnFromInit            ResolverErrMsg = "Can't return a value from an initializer"
	ThisOutSideClass          ResolverErrMsg = "Can't use 'this' outside of a class"
	SuperOutSideClass         ResolverErrMsg = "Can't use 'super' outside of a class"
	SuperWithNoSuperClass     ResolverErrMsg = "Can't use 'super' in a class with no superclass"
	SelfInheritance           ResolverErrMsg = "A class can't inherit from itself"
)

type ResolverErr struct {
	Type ResolverErrMsg
}

func (e *ResolverErr) Error() string {
	return fmt.Sprintf("[ResolverError] %s", e.Type)
}
