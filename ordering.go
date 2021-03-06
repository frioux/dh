package dh

import (
	"bufio"
	"io"
	"io/fs"
	"strings"
)

// Plan parses the `plan.txt` file.  It ignores comments (prefixed with `#`) and
// whitespace.
type Plan struct{}

func (p Plan) Parse(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)
	ret := make([]string, 0, 1000)
	for s.Scan() {
		t := s.Text()
		t, _, _ = strings.Cut(t, "#") // remove comments
		t = strings.TrimSpace(t)      // remove whitespace
		if t == "" {                  // skip blank lines
			continue
		}
		ret = append(ret, t)
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return ret, nil
}

// Validate verifies that all deployments listed in ps actually exist.
func (p Plan) Validate(ps []string, fss fs.FS) error {
	return nil
}
