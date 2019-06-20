package common

type (
	NameAndCmdline struct {
		Name    string
		Cmdline []string
	}

	MatchNamer interface {
		// MatchAndName returns false if the match failed, otherwise
		// true and the resulting name.
		MatchAndName(NameAndCmdline) (bool, string)
	}
)
