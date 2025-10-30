/*
Copyright 2025 The Kubernetes Authors.

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
package store

import (
	"testing"

	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
)

func TestSetCompressionPriority(t *testing.T) {
	tests := []struct {
		name             string
		initialGzip      bool
		initialBrotli    bool
		compressionPrior string
		wantGzip         bool
		wantBrotli       bool
	}{
		{
			name:             "brotli priority",
			initialGzip:      true,
			initialBrotli:    false,
			compressionPrior: "brotli,gzip",
			wantGzip:         false,
			wantBrotli:       true,
		},
		{
			name:             "gzip priority",
			initialGzip:      false,
			initialBrotli:    true,
			compressionPrior: "gzip,brotli",
			wantGzip:         true,
			wantBrotli:       false,
		},
		{
			name:             "unknown defaults to gzip",
			initialGzip:      false,
			initialBrotli:    false,
			compressionPrior: "unknown",
			wantGzip:         true,
			wantBrotli:       false,
		},
		{
			name:             "empty does nothing",
			initialGzip:      false,
			initialBrotli:    false,
			compressionPrior: "",
			wantGzip:         false,
			wantBrotli:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &k8sStore{}
			s.backendConfig = ngx_config.NewDefault()
			s.backendConfig.UseGzip = tc.initialGzip
			s.backendConfig.EnableBrotli = tc.initialBrotli
			s.backendConfig.CompressionPriority = tc.compressionPrior

			s.setCompressionPriority()

			if s.backendConfig.UseGzip != tc.wantGzip {
				t.Fatalf("UseGzip = %v, want %v", s.backendConfig.UseGzip, tc.wantGzip)
			}
			if s.backendConfig.EnableBrotli != tc.wantBrotli {
				t.Fatalf("EnableBrotli = %v, want %v", s.backendConfig.EnableBrotli, tc.wantBrotli)
			}
		})
	}
}
