package dh_test

import (
	"testing"
	"strings"

	"github.com/frioux/dh"
)

func TestPlanParser(t *testing.T) {
	r := strings.NewReader(`foo
bar

# comment
baz
biff # another comment
 oops
`)

	p := dh.Plan{}
	out, err := p.Parse(r)
	if err != nil {
		panic(err)
	}

	if strings.Join(out, " ") != "foo bar baz biff oops" {
		t.Error("plan parsed wrong")
	}
}
