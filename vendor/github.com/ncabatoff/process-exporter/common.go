package common

import "fmt"

type (
	ProcAttributes struct {
		Name     string
		Cmdline  []string
		Username string
	}

	MatchNamer interface {
		// MatchAndName returns false if the match failed, otherwise
		// true and the resulting name.
		MatchAndName(ProcAttributes) (bool, string)
		fmt.Stringer
	}
)
