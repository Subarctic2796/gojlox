package interpreter

import "fmt"

type LoxArray struct {
	Items []any
}

func (la *LoxArray) String() string {
	return fmt.Sprint(la.Items)
}

func (la *LoxArray) checkIndex(index any) (int, error) {
	fidx, ok := index.(float64)
	if !ok {
		return -1, fmt.Errorf("can only use numbers to index arrays got '%s'", index)
	}
	idx := int(fidx)
	ogidx := idx
	if idx < 0 {
		idx = len(la.Items) + idx
	}
	if idx >= 0 && idx < len(la.Items) {
		return idx, nil
	}
	return -1, fmt.Errorf("index out of bounds. index: %d, length: %d", ogidx, len(la.Items))
}

func (la *LoxArray) IndexGet(index any) (any, error) {
	start, err := la.checkIndex(index)
	if err != nil {
		return nil, err
	}
	return la.Items[start], nil
}

func (la *LoxArray) IndexRange(startIndex, stopIndex any) (any, error) {
	var err error
	start := 0
	if startIndex != nil {
		start, err = la.checkIndex(startIndex)
		if err != nil {
			return nil, err
		}
	}
	if stopIndex == nil {
		return &LoxArray{la.Items[start:]}, nil
	}
	stop, err := la.checkIndex(stopIndex)
	if err != nil {
		return nil, err
	}
	return &LoxArray{la.Items[start:stop]}, nil
}

func (la *LoxArray) IndexSet(index any, value any) error {
	idx, err := la.checkIndex(index)
	if err != nil {
		return err
	}
	la.Items[idx] = value
	return nil
}
