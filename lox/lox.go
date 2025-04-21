package lox

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/interpreter"
	"github.com/Subarctic2796/gojlox/parser"
	"github.com/Subarctic2796/gojlox/resolver"
	"github.com/Subarctic2796/gojlox/scanner"
)

type Lox struct {
	HadErr        bool
	HadRunTimeErr bool
	CurErr        error
}

func NewLox() *Lox {
	return &Lox{false, false, nil}
}

func (l *Lox) RunFile(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	intprt := interpreter.NewInterpreter()
	err = l.Run(string(f), intprt)
	if err != nil {
		if l.HadErr {
			os.Exit(65)
		}
		if l.HadRunTimeErr {
			os.Exit(70)
		}
		panic(err)
	}
	return nil
}

func (l *Lox) RunPrompt() error {
	scnr := bufio.NewScanner(os.Stdin)
	intprt := interpreter.NewInterpreter()
	for {
		fmt.Print("> ")
		if !scnr.Scan() {
			fmt.Print("\n")
			if err := scnr.Err(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
		}
		l.Run(scnr.Text(), intprt)
		l.HadErr, l.HadRunTimeErr, l.CurErr = false, false, nil
	}
}

func (l *Lox) Run(src string, intprt *interpreter.Interpreter) error {
	scanner := scanner.NewScanner(src)
	toks, err := scanner.ScanTokens()
	if err != nil {
		l.HadErr, l.CurErr = true, err
		return l.CurErr
	}

	parser := parser.NewParser(toks)
	stmts, err := parser.Parse()
	if err != nil {
		l.HadErr, l.CurErr = true, err
		return l.CurErr
	}

	rslvr := resolver.NewResolver(intprt)
	err = rslvr.ResolveStmts(stmts)
	if err != nil {
		l.HadErr, l.CurErr = true, err
		return l.CurErr
	}
	err = intprt.Interpret(stmts)
	if err != nil {
		l.HadRunTimeErr, l.CurErr = true, err
		return l.CurErr
	}
	return nil
}
