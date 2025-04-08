package interpreter

type ReturnErr struct {
	Value any
}

func (re *ReturnErr) Error() string {
	return "Return Error"
}
