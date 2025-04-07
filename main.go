package main

import (
	"fmt"
	"os"

	"github.com/Subarctic2796/gojlox/lox"
)

func main() {
	os.Args = os.Args[1:]
	lox := lox.LOX
	switch len(os.Args) {
	case 0:
		lox.RunPrompt()
	case 1:
		lox.RunFile(os.Args[0])
	default:
		fmt.Fprintln(os.Stderr, "Usage: gojlox [script]")
		os.Exit(64)
	}
}
