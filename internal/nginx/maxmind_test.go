/*
Copyright 2020 The Kubernetes Authors.

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

package nginx

import (
	"reflect"
	"testing"
)

func resetForTesting() {
	fileExists = _fileExists
	MaxmindLicenseKey = ""
	MaxmindEditionIDs = ""
	MaxmindEditionFiles = []string{}
	MaxmindMirror = ""
}

func TestGeoLite2DBExists(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		want      bool
		wantFiles []string
	}{
		{
			name:      "empty",
			wantFiles: []string{},
		},
		{
			name: "existing files",
			setup: func() {
				MaxmindEditionIDs = "GeoLite2-City,GeoLite2-ASN"
				fileExists = func(string) bool {
					return true
				}
			},
			want:      true,
			wantFiles: []string{"GeoLite2-City.mmdb", "GeoLite2-ASN.mmdb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetForTesting()
			// mimics assignment in flags.go
			config := &MaxmindEditionFiles

			if tt.setup != nil {
				tt.setup()
			}
			if got := GeoLite2DBExists(); got != tt.want {
				t.Errorf("GeoLite2DBExists() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(MaxmindEditionFiles, tt.wantFiles) {
				t.Errorf("nginx.MaxmindEditionFiles = %v, want %v", MaxmindEditionFiles, tt.wantFiles)
			}
			if !reflect.DeepEqual(*config, tt.wantFiles) {
				t.Errorf("config.MaxmindEditionFiles = %v, want %v", *config, tt.wantFiles)
			}
		})
	}
}
