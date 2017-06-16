/*
Copyright 2017 The Kubernetes Authors.

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

package ingress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEqualConfiguration(t *testing.T) {
	ap, _ := filepath.Abs("../../../tests/manifests/configuration-a.json")
	a, err := readJSON(ap)
	if err != nil {
		t.Errorf("unexpected error reading JSON file: %v", err)
	}

	bp, _ := filepath.Abs("../../../tests/manifests/configuration-b.json")
	b, err := readJSON(bp)
	if err != nil {
		t.Errorf("unexpected error reading JSON file: %v", err)
	}

	cp, _ := filepath.Abs("../../../tests/manifests/configuration-c.json")
	c, err := readJSON(cp)
	if err != nil {
		t.Errorf("unexpected error reading JSON file: %v", err)
	}

	if !a.Equal(b) {
		t.Errorf("expected equal configurations (configuration-a.json and configuration-b.json)")
	}

	if !b.Equal(a) {
		t.Errorf("expected equal configurations (configuration-a.json and configuration-b.json)")
	}

	if a.Equal(c) {
		t.Errorf("expected equal configurations (configuration-a.json and configuration-c.json)")
	}

}

func readJSON(p string) (*Configuration, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	var c Configuration

	d := json.NewDecoder(f)
	err = d.Decode(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
