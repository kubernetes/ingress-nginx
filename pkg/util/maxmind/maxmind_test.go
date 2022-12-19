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

package maxmind

import (
	"reflect"
	"testing"
)

func TestGeoLite2DBExists(t *testing.T) {
	tests := []struct {
		name              string
		maxmindEditionIDs string
		setup             func()
		want              bool
		wantFiles         []string
		wantLen           int
	}{
		{
			name:      "empty",
			wantFiles: nil,
			wantLen:   0,
		},
		{
			name:              "existing files",
			maxmindEditionIDs: "GeoLite2-City,GeoLite2-ASN",
			want:              true,
			wantFiles:         []string{"GeoLite2-City.mmdb", "GeoLite2-ASN.mmdb"},
			wantLen:           2,
			setup: func() {
				fileExists = func(string) bool {
					return true
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			files, got := GeoLite2DBExists(tt.maxmindEditionIDs)
			if got != tt.want {
				t.Errorf("GeoLite2DBExists() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(files, tt.wantFiles) {
				t.Errorf("nginx.MaxmindEditionFiles = %v, want %v", files, tt.wantFiles)
			}
			if !reflect.DeepEqual(len(files), tt.wantLen) {
				t.Errorf("nginx.MaxmindEditionFiles = %v, want %v", len(files), tt.wantLen)
			}

		})
	}
}
