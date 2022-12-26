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

package ssl

import (
	"bytes"
	"fmt"
)

// AssembleCACertWithCertAndKey creates a new buffer bytes with key, cert and CA to be used
// by ingress
func AssembleCACertWithCertAndKey(ca, certkey []byte) ([]byte, error) {
	var buffer bytes.Buffer

	_, err := buffer.Write(certkey)
	if err != nil {
		return nil, fmt.Errorf("could not append certkey to buffer: %w", err)
	}

	_, err = buffer.Write([]byte("\n"))
	if err != nil {
		return nil, fmt.Errorf("could not append newline to buffer: %w", err)
	}

	_, err = buffer.Write(ca)
	if err != nil {
		return nil, fmt.Errorf("could not write ca data to buffer: %w", err)
	}
	return buffer.Bytes(), nil
}
