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

package auth

import (
	"reflect"
	"testing"

	api "k8s.io/api/core/v1"
)

func TestDumpSecretAuthFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		secret   *api.Secret
		want     []byte
		wantErr  bool
	}{
		{
			name:     "should reject invalid secret",
			filename: "/tmp/secret1",
			secret: &api.Secret{
				Data: map[string][]byte{
					"nonauth": []byte("bla"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "should reject invalid path",
			filename: "/somethinginvalid/path",
			secret: &api.Secret{
				Data: map[string][]byte{
					"auth": []byte("bla"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "should accept secret",
			filename: "/tmp/secret1",
			secret: &api.Secret{
				Data: map[string][]byte{
					"auth": []byte("bla"),
				},
			},
			want:    []byte("bla"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpSecretAuthFile(tt.filename, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpSecretAuthFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpSecretAuthFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDumpSecretAuthMap(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		secret   *api.Secret
		want     []byte
		wantErr  bool
	}{
		{
			name:     "should reject invalid path",
			filename: "/somethinginvalid/path",
			secret: &api.Secret{
				Data: map[string][]byte{
					"auth": []byte("bla"),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "should accept secret",
			filename: "/tmp/secret1",
			secret: &api.Secret{
				Data: map[string][]byte{
					"user1": []byte("bla"),
					"user2": []byte("blo"),
				},
			},
			want:    []byte("user1:bla\nuser2:blo\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DumpSecretAuthMap(tt.filename, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("DumpSecretAuthMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DumpSecretAuthMap() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
