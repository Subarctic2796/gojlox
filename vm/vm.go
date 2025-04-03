package vm

type VM interface {
	ReportErr(line int, msg string)
	Report(line int, where, msg string)
}
