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
	"fmt"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	"k8s.io/ingress-nginx/internal/ingress/annotations/cors"
)

func buildCorsDirectives(locationcors *cors.Config) ngx_crossplane.Directives {
	directives := make(ngx_crossplane.Directives, 0)
	if len(locationcors.CorsAllowOrigin) > 0 {
		directives = append(directives, buildCorsOriginRegex(locationcors.CorsAllowOrigin)...)
	}
	directives = append(directives,
		buildBlockDirective("if",
			[]string{"$request_method", "=", "OPTIONS"}, ngx_crossplane.Directives{
				buildDirective("set", "$cors", "${cors}options"),
			},
		),
		commonCorsDirective(locationcors, false),
		commonCorsDirective(locationcors, true),
	)

	return directives
}

// commonCorsDirective builds the common cors directives for a location
func commonCorsDirective(cfg *cors.Config, options bool) *ngx_crossplane.Directive {
	corsDir := "true"
	if options {
		corsDir = "trueoptions"
	}
	corsBlock := buildBlockDirective("if", []string{"$cors", "=", corsDir},
		ngx_crossplane.Directives{
			buildDirective("more_set_headers", "Access-Control-Allow-Origin: $http_origin"),
			buildDirective("more_set_headers", fmt.Sprintf("Access-Control-Allow-Methods: %s", cfg.CorsAllowMethods)),
			buildDirective("more_set_headers", fmt.Sprintf("Access-Control-Allow-Headers: %s", cfg.CorsAllowHeaders)),
			buildDirective("more_set_headers", fmt.Sprintf("Access-Control-Max-Age: %d", cfg.CorsMaxAge)),
		},
	)

	if cfg.CorsAllowCredentials {
		corsBlock.Block = append(corsBlock.Block,
			buildDirective("more_set_headers", fmt.Sprintf("Access-Control-Allow-Credentials: %t", cfg.CorsAllowCredentials)),
		)
	}
	if cfg.CorsExposeHeaders != "" {
		corsBlock.Block = append(corsBlock.Block,
			buildDirective("more_set_headers", fmt.Sprintf("Access-Control-Expose-Headers: %s", cfg.CorsExposeHeaders)),
		)
	}

	if options {
		corsBlock.Block = append(corsBlock.Block,
			buildDirective("more_set_headers", "Content-Type: text/plain charset=UTF-8"),
			buildDirective("more_set_headers", "Content-Length: 0"),
			buildDirective("return", "204"),
		)
	}
	return corsBlock
}
