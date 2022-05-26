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

import "testing"

func TestCheckRegex(t *testing.T) {

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "must refuse invalid root",
			wantErr: true,
			value:   "   root    blabla/lala ;",
		},
		{
			name:    "must refuse invalid alias",
			wantErr: true,
			value:   "   alias    blabla/lala ;",
		},
		{
			name:    "must refuse invalid alias with line break",
			wantErr: true,
			value:   "alias #\n/lalala/1/;",
		},
		{
			name:    "must refuse invalid attempt to call /etc",
			wantErr: true,
			value:   "location /etc/nginx/lalala",
		},
		{
			name:    "must refuse invalid attempt to call k8s secret",
			wantErr: true,
			value:   "ssl_cert /var/run/secrets/kubernetes.io/lalala; xpto",
		},
		{
			name:    "must refuse invalid attempt to call lua directives",
			wantErr: true,
			value:   "set_by_lua lala",
		},
		{
			name:    "must pass with valid configuration",
			wantErr: false,
			value:   "/test/mypage1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckRegex(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("CheckRegex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
