/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inspector

import (
	"fmt"
	"regexp"
)

var (
	invalidAliasDirective = regexp.MustCompile(`(?s)\s*alias\s*.*;`)
	invalidRootDirective  = regexp.MustCompile(`(?s)\s*root\s*.*;`)
	invalidEtcDir         = regexp.MustCompile(`/etc/(passwd|shadow|group|nginx|ingress-controller)`)
	invalidSecretsDir     = regexp.MustCompile(`/var/run/secrets`)
	invalidByLuaDirective = regexp.MustCompile(`.*_by_lua.*`)

	// validPathType enforces alphanumeric, -, _ , . and / characters.
	// The field (?i) turns this regex case-insensitive
	// The remaining regex says that the string must start with a "/" (^/)
	// the group [[:alnum:]\_\-\/\.]* says that any amount of characters (A-Za-z0-9), _, - , . and /
	// are accepted until the end of the line
	// Nothing else is accepted.
	validPathType = regexp.MustCompile(`(?i)^/[[:alnum:]._\-/]*$`)

	invalidRegex = []regexp.Regexp{}
)

func init() {
	invalidRegex = []regexp.Regexp{
		*invalidAliasDirective,
		*invalidRootDirective,
		*invalidEtcDir,
		*invalidSecretsDir,
		*invalidByLuaDirective,
	}
}

// CheckRegex receives a value/configuration and validates if it matches with one of the
// forbidden regexes.
func CheckRegex(value string) error {
	for i := range invalidRegex {
		if invalidRegex[i].MatchString(value) {
			return fmt.Errorf("invalid value found: %s", value)
		}
	}
	return nil
}
