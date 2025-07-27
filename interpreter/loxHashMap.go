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

func Hashable(obj any) error {
	switch obj.(type) {
	case nil:
		return nil
	case float64:
		return nil
	case string:
		return nil
	case bool:
		return nil
	case *LoxInstance:
		return nil
	default:
		return fmt.Errorf("unhashable type '%T'", obj)
	}
}

type LoxHashMap struct {
	// Pairs map[uint]*LoxPair
	Pairs map[any]any
}

func (lhm *LoxHashMap) hashObj(obj any) (uint, error) {
	hasher := fnv.New64a()
	switch val := obj.(type) {
	case string:
		hasher.Write([]byte(val))
		return uint(hasher.Sum64()), nil
	case float64:
		return uint(val + 1), nil
	case bool:
		if val {
			return 3, nil
		}
		return 5, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unhashable type '%T'", obj)
	}
}

func (lhm *LoxHashMap) IndexGet(index any) (any, error) {
	// index is key
	if err := Hashable(index); err != nil {
		return nil, err
	}
	if val, ok := lhm.Pairs[index]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key '%s' not present", index)
	/* hash, err := lhm.hashObj(index)
	if err != nil {
		return nil, err
	}
	if pair, ok := lhm.Pairs[hash]; ok {
		return pair.Value, nil
	}
	return nil, fmt.Errorf("key '%s' not present", index) */
}

func (lhm *LoxHashMap) IndexRange(start, stop any) (any, error) {
	return nil, RangeHashMapErr
}

func (lhm *LoxHashMap) IndexSet(index any, value any) error {
	// index is key
	if err := Hashable(index); err != nil {
		return err
	}
	if _, ok := lhm.Pairs[index]; ok {
		lhm.Pairs[index] = value
		return nil
	}
	lhm.Pairs[index] = value
	return nil
	/* key, err := lhm.hashObj(index)
	if err != nil {
		return err
	}
	if _, ok := lhm.Pairs[key]; ok {
		lhm.Pairs[key].Value = value
		return nil
	}
	lhm.Pairs[key] = &LoxPair{index, value}
	return nil */
}

func (lhm *LoxHashMap) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	for k, v := range lhm.Pairs {
		sb.WriteString(fmt.Sprintf("%s: %s, ", k, v))
	}
	sb.WriteString("}")
	return sb.String()
}
