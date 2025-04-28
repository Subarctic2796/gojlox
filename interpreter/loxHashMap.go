package interpreter

import (
	"fmt"
	"hash/fnv"
	"strings"
)

type LoxPair struct {
	Key   any
	Value any
}

type LoxHashMap struct {
	Pairs map[uint]*LoxPair
}

func (lhm *LoxHashMap) hashObj(obj any) (uint, error) {
	hasher := fnv.New64a()
	switch val := obj.(type) {
	case string:
		hasher.Write([]byte(val))
		return uint(hasher.Sum64()), nil
	case float64:
		return uint(val), nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unhashable type '%T'", obj)
	}
}

func (lhm *LoxHashMap) IndexGet(index any) (any, error) {
	hash, err := lhm.hashObj(index)
	if err != nil {
		return nil, err
	}
	if pair, ok := lhm.Pairs[hash]; ok {
		return pair.Value, nil
	}
	return nil, fmt.Errorf("key '%s' not present", index)
}

func (lhm *LoxHashMap) IndexSet(index any, value any) error {
	// index is LoxPair.Key
	key, err := lhm.hashObj(index)
	if err != nil {
		return err
	}
	if _, ok := lhm.Pairs[key]; ok {
		lhm.Pairs[key].Value = value
		return nil
	}
	lhm.Pairs[key] = &LoxPair{index, value}
	return nil
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
