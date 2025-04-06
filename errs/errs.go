package errs

import (
	"errors"

	"github.com/Subarctic2796/gojlox/token"
)

type ErrorReporter interface {
	ReportErr(line int, msg error)
	ReportTok(tok *token.Token, msg error)
	Report(line int, where string, msg error)
}

var (
	ErrUnexpectedChar  = errors.New("Unexpected character")
	ErrUnterminatedStr = errors.New("Unterminated string")
	ErrParse           = errors.New("Parser Error")
)
