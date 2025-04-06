package errs

import "errors"

type ErrorReporter interface {
	ReportErr(line int, msg error)
	Report(line int, where string, msg error)
}

var (
	ErrUnexpectedChar  = errors.New("Unexpected character")
	ErrUnterminatedStr = errors.New("Unterminated string")
)
