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
	ErrUnexpectedChar  = errors.New("Unexpected character")
	ErrUnterminatedStr = errors.New("Unterminated string")
	ErrParse           = errors.New("Parser Error")
)

type RunTimeErr struct {
	Tok *token.Token
	Msg string
}

func (e *RunTimeErr) Error() string {
	return fmt.Sprintf("RunTimeError: %s: %s", e.Tok, e.Msg)
}
