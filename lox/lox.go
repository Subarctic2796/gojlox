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
	interpreter   *interpreter.Interpreter
	resolver      *resolver.Resolver
	parser        *parser.Parser
	lexer         *scanner.Lexer
}

func NewLox() *Lox {
	l := Lox{
		false,
		false,
		interpreter.NewInterpreter(),
		nil,
		parser.NewParser(nil),
		scanner.NewLexer(""),
	}
	l.resolver = resolver.NewResolver(l.interpreter)
	return &l
}

func (l *Lox) RunFile(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	err = l.Run(string(f))
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
	for {
		fmt.Print("> ")
		if !scnr.Scan() {
			fmt.Print("\n")
			return scnr.Err()
		}
		_ = l.Run(scnr.Text())
		l.HadErr, l.HadRunTimeErr = false, false
	}
}

func (l *Lox) Run(src string) error {
	l.lexer.Reset(src)
	toks, err := l.lexer.ScanTokens()
	if err != nil {
		l.HadErr = true
		return err
	}

	l.parser.Reset(toks)
	stmts, err := l.parser.Parse()
	if err != nil {
		l.HadErr = true
		return err
	}

	l.resolver.Reset()
	err = l.resolver.ResolveStmts(stmts)
	if err != nil {
		l.HadErr = true
		return err
	}

	err = l.interpreter.Interpret(stmts)
	if err != nil {
		l.HadRunTimeErr = true
		return err
	}

	return nil
}
