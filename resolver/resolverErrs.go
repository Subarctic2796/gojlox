package resolver

import "errors"

var (
	ErrAlreadyInScope            = errors.New("Already a variable with this name in this scope")
	ErrReadLocalInOwnInitializer = errors.New("Can't read local variable in its own initializer")
	ErrReturnTopLevel            = errors.New("Can't return from top-level code")
	ErrReturnFromInit            = errors.New("Can't return a value from an initializer")
	ErrThisNotInClass            = errors.New("Can't use 'this' outside of a class")
	ErrSuperNotInClass           = errors.New("Can't use 'super' outside of a class")
	ErrSuperWithNoSuperClass     = errors.New("Can't use 'super' in a class with no superclass")
	ErrSuperInStatic             = errors.New("Can't use 'super' in a static method")
	ErrInheritsSelf              = errors.New("A class can't inherit from itself")
	ErrInitIsStatic              = errors.New("Can't use 'init' as a static function")
	ErrLocalNotRead              = errors.New("Local variable is not used")
)
