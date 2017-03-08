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

package utils

import (
	"fmt"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"testing"
)

func TestGetSetClusterName(t *testing.T) {
	fooTests := []struct {
		cn  string
		ecn string
	}{
		{"", ""},
		{"fooCN", "fooCN"},
		{"fooCN01--fooCN02--fooCN03", "fooCN03"},
	}

	namer := &Namer{}
	for _, fooTest := range fooTests {
		namer.SetClusterName(fooTest.cn)
		rcn := namer.GetClusterName()
		if rcn != fooTest.ecn {
			t.Errorf("returned %v but expected %v", rcn, fooTest.ecn)
		}
	}
}

func TestGetSetFirewallName(t *testing.T) {
	fooTests := []struct {
		fn  string
		efn string
	}{
		{"", ""},
		{"fooFN", "fooFN"},
		{"fooFN01--fooFN02--fooFN03", "fooFN01--fooFN02--fooFN03"},
	}

	namer := &Namer{}
	for _, fooTest := range fooTests {
		namer.SetFirewallName(fooTest.fn)
		rfn := namer.GetFirewallName()
		if rfn != fooTest.efn {
			t.Errorf("returned %v but expected %v", rfn, fooTest.efn)
		}
	}
}

func TestNewNamer(t *testing.T) {
	cn, fn := "fooCN", "fooFN"

	namer := NewNamer(cn, fn)
	rcn := namer.GetClusterName()
	if rcn != cn {
		t.Errorf("returned %v but expected %v", rcn, cn)
	}

	rfn := namer.GetFirewallName()
	if rfn != fn {
		t.Errorf("returned %v but expected %v", rfn, fn)
	}
}

func TestTruncate(t *testing.T) {
	nameLenLimitValue := ""
	for i := 0; i < nameLenLimit; i++ {
		nameLenLimitValue = fmt.Sprintf("%s%s", nameLenLimitValue, "t")
	}

	fooTests := []struct {
		k  string
		ek string
	}{
		{"", ""},
		{"fooKey", "fooKey"},
		{nameLenLimitValue + "more", nameLenLimitValue + "0"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer("fooCN", "fooFN")
		rk := namer.Truncate(fooTest.k)
		if rk != fooTest.ek {
			t.Errorf("returned %v but expected %v", rk, fooTest.ek)
		}
	}
}

func TestDecorateName(t *testing.T) {
	nameLenLimitValue := ""
	for i := 0; i < nameLenLimit-6; i++ {
		nameLenLimitValue = fmt.Sprintf("%s%s", nameLenLimitValue, "t")
	}

	fooTests := []struct {
		cn  string
		n   string
		edn string
	}{
		{"", "", ""},
		{"", "fooName", "fooName"},
		{"fooCN", "fooName", "fooName--fooCN"},
		{"fooCN", nameLenLimitValue, nameLenLimitValue + "--fooC0"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer(fooTest.cn, "fooFN")
		rdn := namer.decorateName(fooTest.n)
		if rdn != fooTest.edn {
			t.Errorf("returned %v but expected %v", rdn, fooTest.edn)
		}
	}
}

func TestParseName(t *testing.T) {
	fooTests := []struct {
		n   string
		ecn string
		ers string
	}{
		{"", "", ""},
		{"foo", "", ""},
		{"foo01--", "", ""},
		{"foo01--foo02", "foo02", ""},
		{"foo01-foo02--foo03", "foo03", "foo02"},
	}

	namer := NewNamer("fooCN", "fooFN")
	for _, fooTest := range fooTests {
		nc := namer.ParseName(fooTest.n)
		if nc == nil {
			t.Errorf("returned nil")
			continue
		}

		if nc.ClusterName != fooTest.ecn {
			t.Errorf("returned %v but expected %v", nc.ClusterName, fooTest.ecn)
			continue
		}

		if nc.Resource != fooTest.ers {
			t.Errorf("returned %v but expected %v", nc.Resource, fooTest.ers)
			continue
		}
	}
}

func TestNameBelongsToCluster(t *testing.T) {
	nameLenLimitValue := "k8s-foo01--"
	for i := 0; i < nameLenLimit; i++ {
		nameLenLimitValue = fmt.Sprintf("%s%s", nameLenLimitValue, "t")
	}
	nn := &Namer{}
	trn := nn.Truncate(nameLenLimitValue)

	fooTests := []struct {
		n  string
		cn string
		ev bool
	}{
		{"", "", false},
		{"foo01", "", false},
		{"k8s-foo01", "", true},
		{"k8s-foo01", "foo01", false},
		{"k8s-foo01--foo02", "", false},
		{"k8s-foo01--foo02", "foo", false},
		{"k8s-foo01--foo02", "foo02", true},
		{"k8s-foo01--foo02", "k8s-foo01--foo02", true},
		{"k8s-foo01--foo02--foo03", "foo02", false},
		{"k8s-foo01--foo02--foo03", "foo02--foo03", false},
		{"k8s-foo01--foo02--foo03", "k8s-foo01--foo02--foo03", false},
		{trn, trn, true},
	}

	namer := NewNamer("fooCN", "fooFN")
	for _, fooTest := range fooTests {
		namer.SetClusterName(fooTest.cn)
		rv := namer.NameBelongsToCluster(fooTest.n)
		if rv != fooTest.ev {
			t.Errorf("returned %v but expected %v", rv, fooTest.ev)
		}
	}
}

func TestBeName(t *testing.T) {
	fooTests := []struct {
		cn string
		p  int64
		en string
	}{
		{"", 80, "k8s-be-80"},
		{"fooCN", 80, "k8s-be-80--fooCN"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer(fooTest.cn, "fooFN")
		rn := namer.BeName(fooTest.p)
		if rn != fooTest.en {
			t.Errorf("returned %v but expected %v", rn, fooTest.en)
		}
	}
}

func TestBePort(t *testing.T) {
	namer := NewNamer("fooCN", "fooFN")

	fooTests := []struct {
		bn string
		ie bool
		bp string
	}{
		{"", true, ""},
		{"fooBP", true, ""},
		{"k8s-be-1022 k8s-be-033", false, "1022"},
		{"k8s-be-022 k8s-be-033", false, "022"},
	}

	for _, fooTest := range fooTests {
		rbp, err := namer.BePort(fooTest.bn)
		if fooTest.ie {
			if err == nil {
				t.Errorf("expected error")
			}
			continue
		}

		if err != nil {
			t.Errorf("unexpected error")
			continue
		}

		if rbp != fooTest.bp {
			t.Errorf("returned %v but expected %v", rbp, fooTest.bp)
		}
	}
}

func TestIGName(t *testing.T) {
	fooTests := []struct {
		cn string
		en string
	}{
		{"", "k8s-ig"},
		{"fooCN", "k8s-ig--fooCN"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer(fooTest.cn, "fooFN")
		rn := namer.IGName()
		if rn != fooTest.en {
			t.Errorf("returned %v but expected %v", rn, fooTest.en)
		}
	}
}

func TestFrSuffix(t *testing.T) {
	fooTests := []struct {
		cn  string
		fn  string
		efn string
	}{
		{"", "", "l7"},
		{"fooCN", "", "l7--fooCN"},
		{"fooCN", "fooFN", "l7--fooFN"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer(fooTest.cn, fooTest.fn)
		rfn := namer.FrSuffix()
		if rfn != fooTest.efn {
			t.Errorf("returned %v but expected %v", rfn, fooTest.efn)
		}
	}
}

func TestFrName(t *testing.T) {
	fooTests := []struct {
		s   string
		efn string
	}{
		{"", "k8s-fw-"},
		{"foo", "k8s-fw-foo"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer("fooCN", "fooFN")
		rfn := namer.FrName(fooTest.s)
		if rfn != fooTest.efn {
			t.Errorf("returned %v but expected %v", rfn, fooTest.efn)
		}
	}
}

func TestLBName(t *testing.T) {
	fooTests := []struct {
		cn   string
		k    string
		elbn string
	}{
		{"", "", ""},
		{"", "default/k8s-01--foo01", "default-k8s-01--foo01"},
		{"fooCN", "default/k8s-01--fooCN", "default-k8s-01--fooCN"},
		{"fooCN", "default/k8s-01--foo01", "default-k8s-01--foo01--fooCN"},
	}

	for _, fooTest := range fooTests {
		namer := NewNamer(fooTest.cn, "fooFN")
		rlbn := namer.LBName(fooTest.k)
		if rlbn != fooTest.elbn {
			t.Errorf("returned %v but expected %v", rlbn, fooTest.elbn)
		}
	}
}

func TestGetDefaultBackend(t *testing.T) {
	var nilBS *compute.BackendService
	bs := &compute.BackendService{CreationTimestamp: "2017-03-08"}

	fooTests := []struct {
		gm  GCEURLMap
		ebs *compute.BackendService
	}{
		{map[string]map[string]*compute.BackendService{}, nilBS},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {},
		}, nilBS},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {
				DefaultBackendKey + "_non": bs,
			},
		}, nilBS},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {
				DefaultBackendKey: bs,
			},
		}, bs},
	}

	for _, fooTest := range fooTests {
		rbs := fooTest.gm.GetDefaultBackend()
		if rbs != fooTest.ebs {
			t.Errorf("returned %v but expected %v", rbs, fooTest.ebs)
		}
	}
}

func TestString(t *testing.T) {
	fooTests := []struct {
		gm GCEURLMap
		es string
	}{
		{
			map[string]map[string]*compute.BackendService{},
			"",
		},
		{
			map[string]map[string]*compute.BackendService{
				"localhost": {},
			},
			"localhost\n",
		},
		{
			map[string]map[string]*compute.BackendService{
				"localhost": {
					"10.0.0.0": nil,
				},
			},
			"localhost\n\t10.0.0.0: No backend\n",
		},
		{
			map[string]map[string]*compute.BackendService{
				"localhost": {
					"10.0.0.0": {
						Name: "foo",
					},
				},
			},
			"localhost\n\t10.0.0.0: foo\n",
		},
	}

	for _, fooTest := range fooTests {
		rs := fooTest.gm.String()
		if rs != fooTest.es {
			t.Errorf("returned %v but expected %v", rs, fooTest.es)
		}
	}
}

func TestPutDefaultBackend(t *testing.T) {
	obs := &compute.BackendService{CreationTimestamp: "2017-03-08"}
	nbs := &compute.BackendService{Name: "foo"}

	fooTests := []struct {
		gm GCEURLMap
	}{
		{map[string]map[string]*compute.BackendService{}},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {},
		}},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {
				DefaultBackendKey + "_non": obs,
			},
		}},
		{map[string]map[string]*compute.BackendService{
			DefaultBackendKey: {
				DefaultBackendKey: obs,
			},
		}},
	}

	for _, fooTest := range fooTests {
		fooTest.gm.PutDefaultBackend(nbs)
		rbs := fooTest.gm.GetDefaultBackend()
		if rbs != nbs {
			t.Errorf("returned %v but expected %v", rbs, nbs)
		}
	}
}

type fakeError struct{}

func (f *fakeError) Error() string {
	// do nothing
	return ""
}

func TestIsHTTPErrorCode(t *testing.T) {
	fooTests := []struct {
		err  error
		code int
		eb   bool
	}{
		{&fakeError{}, 202, false},
		{&googleapi.Error{Code: 202}, 208, false},
		{&googleapi.Error{Code: 202}, 202, true},
	}

	for _, fooTest := range fooTests {
		rb := IsHTTPErrorCode(fooTest.err, fooTest.code)
		if rb != fooTest.eb {
			t.Errorf("returned %v but expected %v", rb, fooTest.eb)
		}
	}
}

func TestCompareLinks(t *testing.T) {
	fooTests := []struct {
		l1 string
		l2 string
		eb bool
	}{
		{"", "", false},
		{"", "l2", false},
		{"l1", "", false},
		{"l", "l", true},
	}

	for _, fooTest := range fooTests {
		rb := CompareLinks(fooTest.l1, fooTest.l2)
		if rb != fooTest.eb {
			t.Errorf("returned %v but expected %v", rb, fooTest.eb)
		}
	}
}
