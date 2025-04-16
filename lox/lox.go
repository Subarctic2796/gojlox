package lox

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/interpreter"
	"github.com/Subarctic2796/gojlox/parser"
	"github.com/Subarctic2796/gojlox/resolver"
	"github.com/Subarctic2796/gojlox/scanner"
	"github.com/Subarctic2796/gojlox/token"
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
		return err
	}
	intprt := interpreter.NewInterpreter(l)
	l.Run(string(f), intprt)
	if l.HadErr {
		os.Exit(65)
	}
	if l.HadRunTimeErr {
		os.Exit(70)
	}
	return nil
}

func (l *Lox) RunPrompt() error {
	scnr := bufio.NewScanner(os.Stdin)
	intprt := interpreter.NewInterpreter(l)
	for {
		fmt.Print("> ")
		if !scnr.Scan() {
			fmt.Print("\n")
			return scnr.Err()
		}
		l.Run(scnr.Text(), intprt)
		l.HadErr = false
	}
}

func (l *Lox) Run(src string, intprt *interpreter.Interpreter) {
	scanner := scanner.NewScanner(src, l)
	if l.HadErr {
		return
	}
	parser := parser.NewParser(scanner.ScanTokens(), l)
	stmts, errs := parser.Parse()
	if l.HadErr {
		fmt.Fprintf(os.Stderr, "%s\n", errs)
		return
	}

	resolver := resolver.NewResolver(l, intprt)
	resolver.ResolveStmts(stmts)
	if l.HadErr {
		return
	}
	intprt.Interpret(stmts)
}

func (l *Lox) ReportErr(line int, msg error) {
	l.Report(line, "", msg)
}

func (l *Lox) Report(line int, where string, msg error) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, msg)
	l.HadErr = true
	l.CurErr = msg
}

func (l *Lox) ReportTok(tok *token.Token, msg error) {
	if tok.Kind == token.EOF {
		l.Report(tok.Line, " at end", msg)
	} else {
		l.Report(tok.Line, fmt.Sprintf(" at '%s'", tok.Lexeme), msg)
	}
}

func (l *Lox) ReportRTErr(msg error) {
	err := &errs.RunTimeErr{}
	if errors.As(msg, &err) {
		fmt.Fprintf(os.Stderr, "%s\n[line %d]\n", msg, err.Tok.Line)
	}
	l.HadRunTimeErr = true
}
