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
