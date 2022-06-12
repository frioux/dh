package dh

import (
	"fmt"
)

type Version interface {
	String() string
}

type VersionIterator[V Version] interface {
	Next() V
}

type MonotonicVersion struct {
	v int
}

func (v MonotonicVersion) String() string { return fmt.Sprintf("%03d", v.v) }

type MonotonicVersionIterator struct {
	Start, Last int
}

func (i *MonotonicVersionIterator) Next() *MonotonicVersion {
	if i.Start < i.Last {
		i.Start++
		return &MonotonicVersion{i.Start}
	}

	return nil
}
