package resolver

import "fmt"

type ResolverErrMsg string

const (
	AlreadyInScope            ResolverErrMsg = "Already a variable with this name in this scope"
	ReadLocalInOwnInitializer ResolverErrMsg = "Can't read local variable in its own initializer"
	ReturnTopLevel            ResolverErrMsg = "Can't return from top-level code"
	ReturnFromInit            ResolverErrMsg = "Can't return a value from an initializer"
	ThisOutSideClass          ResolverErrMsg = "Can't use 'this' outside of a class"
	SuperOutSideClass         ResolverErrMsg = "Can't use 'super' outside of a class"
	SuperWithNoSuperClass     ResolverErrMsg = "Can't use 'super' in a class with no superclass"
	SuperInStatic             ResolverErrMsg = "Can't use 'super' in a static method"
	SelfInheritance           ResolverErrMsg = "A class can't inherit from itself"
	InitIsStatic              ResolverErrMsg = "Can't use 'init' as a static function"
	LocalNotRead              ResolverErrMsg = "Local variable is not used"
)

type ResolverErr struct {
	Type ResolverErrMsg
}

func (e *ResolverErr) Error() string {
	return fmt.Sprintf("[ResolverError] %s", e.Type)
}
