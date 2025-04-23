package lox

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/interpreter"
	"github.com/Subarctic2796/gojlox/lexer"
	"github.com/Subarctic2796/gojlox/parser"
	"github.com/Subarctic2796/gojlox/resolver"
)

type Lox struct {
	HadErr        bool
	HadRunTimeErr bool
	useV2         bool
}

func NewLox() *Lox {
	return &Lox{false, false, true}
	// return &Lox{false, false, false}
}

func (l *Lox) RunFile(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	intprt := interpreter.NewInterpreter(l.useV2)
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
	intprt := interpreter.NewInterpreter(l.useV2)
	for {
		fmt.Print("> ")
		if !scnr.Scan() {
			fmt.Print("\n")
			return scnr.Err()
		}
		_ = l.Run(scnr.Text(), intprt)
		l.HadErr, l.HadRunTimeErr = false, false
	}
}

func (l *Lox) Run(src string, intprt *interpreter.Interpreter) error {
	lex := scanner.NewLexer(src)
	toks, err := lex.ScanTokens()
	if err != nil {
		l.HadErr = true
		return err
	}

	parser := parser.NewParser(toks)
	stmts, err := parser.Parse()
	if err != nil {
		l.HadErr = true
		return err
	}

	rslvr := resolver.NewResolver(intprt, l.useV2)
	err = rslvr.ResolveStmts(stmts)
	if err != nil {
		l.HadErr = true
		return err
	}
	err = intprt.Interpret(stmts)
	if err != nil {
		l.HadRunTimeErr = true
		return err
	}
	return nil
}
