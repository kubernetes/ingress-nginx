package proc

import (
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

// read everything in the iterator
func consumeIter(pi ProcIter) ([]ProcIdInfo, error) {
	infos := []ProcIdInfo{}
	for pi.Next() {
		info, err := Info(pi)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}
