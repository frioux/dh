package dh_test

import (
	"testing"
	"strings"

	"github.com/frioux/dh"
)

func TestMonotonicVersioning(t *testing.T) {
	mvi := dh.MonotonicVersionIterator{0, 5}
	out := []string{}
	for n := mvi.Next(); n != nil; n = mvi.Next() {
		out = append(out, n.String())
	}

	if strings.Join(out, " ") != "001 002 003 004 005" {
		t.Errorf("incorrect results for Monotonic Iterator")
	}
}
