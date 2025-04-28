package interpreter

import "fmt"

type UserClass struct {
	Name       string
	SuperClass *UserClass
	Methods    map[string]*UserFn
}

func NewUserClass(name string, superclass *UserClass, methods map[string]*UserFn) *UserClass {
	return &UserClass{name, superclass, methods}
}

func (lc *UserClass) FindMethod(name string) *UserFn {
	if val, ok := lc.Methods[name]; ok {
		return val
	}
	if lc.SuperClass != nil {
		return lc.SuperClass.FindMethod(name)
	}
	return nil
}

func (lc *UserClass) String() string {
	return fmt.Sprint(lc.Name)
}

// func (lc *UserClass) Call(intprt *Interpreter, args []any) (any, error) {
func (lc *UserClass) Call(args []any) (any, error) {
	inst := NewLoxInstance(lc)
	init := lc.FindMethod("init")
	if init != nil {
		_, _ = init.Bind(inst).Call(args)
		// _, _ = init.Bind(inst).Call(intprt, args)
	}
	return inst, nil
}

func (lc *UserClass) Arity() int {
	init := lc.FindMethod("init")
	if init == nil {
		return 0
	}
	return init.Arity()
}
