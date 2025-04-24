package interpreter

import (
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/ast"
)

type Debugger struct {
	Stmts   []ast.Stmt
	Intprt  *Interpreter
	DBG_FNS map[string]NativeFn
}

func NewDebugger(stmts []ast.Stmt, intprt *Interpreter) *Debugger {
	dbgFns := make(map[string]NativeFn)
	dbgFns["dumpast"] = &DumpAST{stmts}
	return &Debugger{
		Stmts:   stmts,
		Intprt:  intprt,
		DBG_FNS: dbgFns,
	}
}

func (d *Debugger) RunInterpreter() {
	d.Intprt.Globals.Define("dumpast", d.DBG_FNS["dumpast"])
	err := d.Intprt.Interpret(d.Stmts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

type DumpAST struct {
	stmts []ast.Stmt
}

func (d *DumpAST) Call(intprt *Interpreter, args []any) (any, error) {
	for _, stmt := range d.stmts {
		fmt.Println(stmt)
	}
	return nil, nil
}

func (DumpAST) Arity() int {
	return 0
}

func (DumpAST) String() string {
	return "<native fn dumpast>"
}

func (DumpAST) Name() string {
	return "dumpast"
}
