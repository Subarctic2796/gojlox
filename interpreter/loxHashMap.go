package interpreter

import (
	"fmt"
	"strings"
)

type LoxPair struct {
	Key   any
	Value any
}

type LoxHashMap struct {
	Pairs map[uint]*LoxPair
}

func (lhm *LoxHashMap) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	for _, v := range lhm.Pairs {
		sb.WriteString(fmt.Sprintf("%s: %s, ", v.Key, v.Value))
	}
	sb.WriteString("}")
	return sb.String()
}

func (lhm *LoxHashMap) Iterable() {}
