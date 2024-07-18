/*
Copyright 2024 The Kubernetes Authors.

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

package crossplane

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

// THIS FILE SHOULD BE USED JUST FOR INTERNAL TESTS - Private functions

func Test_Internal_buildDirectives(t *testing.T) {
	t.Run("should be able to run a directive with a single argument", func(t *testing.T) {
		directive := buildDirective("somedirective", "bla")
		require.Equal(t, directive.Directive, "somedirective", []string{"bla"})
	})
	t.Run("should be able to run a directive with multiple different arguments", func(t *testing.T) {
		directive := buildDirective("somedirective", "bla", 5, true, seconds(10), []string{"xpto", "bla"})
		require.Equal(t, directive.Directive, "somedirective", []string{"bla", "5", "on", "10s", "xpto", "bla"})
	})
}

func Test_Internal_boolToStr(t *testing.T) {
	require.Equal(t, boolToStr(true), "on")
	require.Equal(t, boolToStr(false), "off")
}

func Test_Internal_buildLuaDictionaries(t *testing.T) {
	t.Skip("Maps are not sorted, need to fix this")
	cfg := &config.Configuration{
		LuaSharedDicts: map[string]int{
			"somedict":  1024,
			"otherdict": 1025,
		},
	}
	directives := buildLuaSharedDictionaries(cfg)
	require.Len(t, directives, 2)
	require.Equal(t, "lua_shared_dict", directives[0].Directive)
	require.Equal(t, []string{"somedict", "1M"}, directives[0].Args)
	require.Equal(t, "lua_shared_dict", directives[1].Directive)
	require.Equal(t, []string{"otherdict", "1025K"}, directives[1].Args)
}
