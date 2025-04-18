package interpreter

import "fmt"

type LoxClass struct {
	Name       string
	SuperClass *LoxClass
	Methods    map[string]*LoxFn
	Statics    map[string]*LoxFn
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]*LoxFn, statics map[string]*LoxFn) *LoxClass {
	return &LoxClass{name, superclass, methods, statics}
}

func (lc *LoxClass) FindMethod(name string) *LoxFn {
	if val, ok := lc.Statics[name]; ok {
		return val
	}
	if val, ok := lc.Methods[name]; ok {
		return val
	}
	if lc.SuperClass != nil {
		return lc.SuperClass.FindMethod(name)
	}
	return nil
}

func (lc *LoxClass) String() string {
	return fmt.Sprint(lc.Name)
}

func (lc *LoxClass) Call(intprt *Interpreter, args []any) (any, error) {
	inst := NewLoxInstance(lc)
	init := lc.FindMethod("init")
	if init != nil {
		init.Bind(inst).Call(intprt, args)
	}
	return inst, nil
}

func (lc *LoxClass) Arity() int {
	init := lc.FindMethod("init")
	if init == nil {
		return 0
	}
	return init.Arity()
}
