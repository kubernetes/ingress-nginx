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

package resolver

import "testing"

func TestAuthSSLCertEqual(t *testing.T) {
	tests := []struct {
		exp         bool
		authcert    *AuthSSLCert
		compareWith *AuthSSLCert
	}{
		{true, &AuthSSLCert{
			Secret: "same_secret_value",
			CRLSHA: "same_crl_sha_value",
		}, &AuthSSLCert{
			Secret: "same_secret_value",
			CRLSHA: "same_crl_sha_value",
		}},

		{false, &AuthSSLCert{
			Secret: "test_ssl_secret",
		}, &AuthSSLCert{
			Secret: "test_ssl_secret",
			CRLSHA: "test_crl_sha",
		}},

		{false, &AuthSSLCert{}, &AuthSSLCert{
			Secret: "test_ssl_secret",
		}},

		{false, nil, &AuthSSLCert{
			Secret: "test_ssl_secret",
		}},
	}

	for _, test := range tests {
		out := test.authcert.Equal(test.compareWith)

		if out != test.exp {
			t.Errorf(
				"AuthSSLCert: %v compared with %v, expected %v but got %v",
				test.authcert, test.compareWith, test.exp, out)
		}
	}
}
