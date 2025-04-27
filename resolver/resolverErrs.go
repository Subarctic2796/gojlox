package resolver

import "errors"

var (
	ErrAlreadyInScope       = errors.New("Already a variable with this name in this scope")
	ErrLocalInitializesSelf = errors.New("Can't read local variable in its own initializer")
	ErrLocalNotRead         = errors.New("Local variable is not used")
	ErrUnHashAble           = errors.New("Can only use: strings, numbers, bools, and instances as hashmap keys")
)
