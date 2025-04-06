package lox

import (
	"bufio"
	"fmt"
	"github.com/Subarctic2796/gojlox/scanner"
	"os"
)

type Lox struct {
	HadErr bool
	CurErr error
}

func NewLox() *Lox {
	return &Lox{false, nil}
}

func (l *Lox) RunFile(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	l.Run(string(f))
	if l.HadErr {
		os.Exit(65)
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
		l.Run(scnr.Text())
		l.HadErr = false
	}
}

func (l *Lox) Run(src string) {
	scanner := scanner.NewScanner(src, l)
	for _, tok := range scanner.ScanTokens() {
		fmt.Println(tok)
	}
}

func (l *Lox) ReportErr(line int, msg error) {
	l.Report(line, "", msg)
}

func (l *Lox) Report(line int, where string, msg error) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, msg)
	l.HadErr = true
	l.CurErr = msg
}
