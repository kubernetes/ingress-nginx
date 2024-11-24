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
	"reflect"
	"testing"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	"github.com/stretchr/testify/require"
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

func Test_Internal_buildMapDirectives(t *testing.T) {
	t.Run("should be able to run build a map directive with empty block", func(t *testing.T) {
		directive := buildMapDirective("somedirective", "bla", ngx_crossplane.Directives{buildDirective("something", "otherstuff")})
		require.Equal(t, directive.Directive, "map")
		require.Equal(t, directive.Args, []string{"somedirective", "bla"})
		require.Equal(t, directive.Block[0].Directive, "something")
		require.Equal(t, directive.Block[0].Args, []string{"otherstuff"})
	})
}

func Test_Internal_boolToStr(t *testing.T) {
	require.Equal(t, boolToStr(true), "on")
	require.Equal(t, boolToStr(false), "off")
}

func Test_Internal_buildCorsOriginRegex(t *testing.T) {
	tests := []struct {
		name        string
		corsOrigins []string
		want        ngx_crossplane.Directives
	}{
		{
			name:        "wildcard returns a single directive",
			corsOrigins: []string{"*"},
			want: ngx_crossplane.Directives{
				buildDirective("set", "$http_origin", "*"),
				buildDirective("set", "$cors", "true"),
			},
		},
		{
			name:        "multiple hosts should be changed properly",
			corsOrigins: []string{"*.xpto.com", "  lalala.com"},
			want: ngx_crossplane.Directives{
				buildBlockDirective("if", []string{"$http_origin", "~*", "(([A-Za-z0-9\\-]+\\.xpto\\.com)|(lalala\\.com))$"},
					ngx_crossplane.Directives{buildDirective("set", "$cors", "true")},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCorsOriginRegex(tt.corsOrigins); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCorsOriginRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}
