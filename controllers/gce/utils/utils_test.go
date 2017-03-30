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
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"reflect"
	"testing"
)

type fakeError struct{}

func (f *fakeError) Error() string {
	return ""
}

func TestGetSetClusterName(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		ecn   string
	}{
		{"empty_clusterName", "", ""},
		{"normal_clusterName", "560b5b8154db29b3", "560b5b8154db29b3"},
		{"clusterName_with_clusterNameDelimiter", "k8s-ig-deployment--560b5b8154db29b3", "560b5b8154db29b3"},
	}
	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := &Namer{firewallName: "fooFN"}
			namer.SetClusterName(fooTest.cn)
			rcn := namer.GetClusterName()
			if rcn != fooTest.ecn {
				t.Errorf("Returned %v but expected %v", rcn, fooTest.ecn)
			}
		})
	}
}

func TestGetSetFirewallName(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		fn    string
		efn   string
	}{
		{"empty_clusterName_and_fire_name", "", "", ""},
		{"empty_fireName", "560b5b8154db29b3", "", "560b5b8154db29b3"},
		{"normal_fireName", "560b5b8154db29b3", "fooFN", "fooFN"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := &Namer{clusterName: fooTest.cn}
			namer.SetFirewallName(fooTest.fn)
			rfn := namer.GetFirewallName()
			if rfn != fooTest.efn {
				t.Errorf("Returned %v but expected %v", rfn, fooTest.efn)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	fooTests := []struct {
		tName string
		k     string
		ek    string
	}{
		{"empty_name", "", ""},
		{"not_truncated", "k8s-tps-e2e-tests-ingress-deployment--560b5b8154db29b3", "k8s-tps-e2e-tests-ingress-deployment--560b5b8154db29b3"},
		{"will_be_turncated", "k8s-tps-e2e-tests-ingress-upgrade-xxmpf-static-ip--560b5b8154db29b3", "k8s-tps-e2e-tests-ingress-upgrade-xxmpf-static-ip--560b5b8154d0"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer("fooCN", "fooFN")
			rk := namer.Truncate(fooTest.k)
			if rk != fooTest.ek {
				t.Errorf("Returned %v but expected %v", rk, fooTest.ek)
			}
		})
	}
}

func TestDecorateName(t *testing.T) {
	fooTests := []struct {
		tName string
		n     string
		cn    string
		edn   string
	}{
		{"empty_clusterName", "", "", ""},
		{"empty_decode", "", "560b5b8154db29b3", "--560b5b8154db29b3"},
		{"normal_test", "k8s-tps-e2e-ingress-upgrace-xyxloe-auth", "560b5b8154db29b3", "k8s-tps-e2e-ingress-upgrace-xyxloe-auth--560b5b8154db29b3"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.cn, "fooFN")
			rdn := namer.decorateName(fooTest.n)
			if rdn != fooTest.edn {
				t.Errorf("Returned %v but expected %v", rdn, fooTest.edn)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	fooTests := []struct {
		tName string
		n     string
		ecn   string
		ers   string
	}{
		{"empty_name", "", "", ""},
		{"unexpected_regular_match", "foo", "", ""},
		{"name_with_clusterNameDelimiter", "ip--560b5b8154db29b3", "560b5b8154db29b3", ""},
		{"name_with_clusterNameDelimiter_and_resource", "k8s-tps--560b5b8154db29b3", "560b5b8154db29b3", "tps"},
	}

	namer := NewNamer("fooCN", "fooFN")
	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			nc := namer.ParseName(fooTest.n)
			if nc == nil {
				t.Fatal("Unexpected nil")
			}

			if nc.ClusterName != fooTest.ecn {
				t.Fatalf("Returned %v but expected %v", nc.ClusterName, fooTest.ecn)
			}

			if nc.Resource != fooTest.ers {
				t.Errorf("Returned %v but expected %v", nc.Resource, fooTest.ers)
			}
		})
	}
}

func TestNameBelongsToCluster(t *testing.T) {
	nn := &Namer{}

	fooTests := []struct {
		tName       string
		name        string
		clusterName string
		ev          bool
	}{
		{"empty_clusterName_and_name", "", "", false},
		{"not_part_of_cluster", "560b5b8154db29b3", "560b5b8154db29b3", false},
		{"only_part1_and_empty_clusterName", "k8s-ig-560b5b8154db29b3", "", true},
		{"only_part1_and_normal_clusterName", "k8s-ig-560b5b8154db29b3", "560b5b8154db29b3", false},
		{"too_manay_parts", "k8s-ig--static-ip--560b5b8154db29b3", "560b5b8154db29b3", false},
		{"normal_test", "k8s-ig-test-xrp--560b5b8154db29b3", "560b5b8154db29b3", true},
		{"truncated_normal_test", nn.Truncate("k8s-tps-e2e-tests-ingress-upgrade-xxmpf-static-ip--560b5b8154db29b3"), "560b5b8154db29b3", true},
		{"truncated_unexpected_test", nn.Truncate("k8s-tps-e2e-tests-ingress-ip--560b5b8154db29b0"), "560b5b8154db29b3", false},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.clusterName, "fooFN")
			rv := namer.NameBelongsToCluster(fooTest.name)
			if rv != fooTest.ev {
				t.Errorf("Returned %v but expected %v", rv, fooTest.ev)
			}

		})
	}
}

func TestBeName(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		p     int64
		en    string
	}{
		{"empty_clusterName", "", 80, "k8s-be-80"},
		{"normal_test", "560b5b8154db29b3", 80, "k8s-be-80--560b5b8154db29b3"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.cn, "fooFN")
			rn := namer.BeName(fooTest.p)
			if rn != fooTest.en {
				t.Errorf("Returned %v but expected %v", rn, fooTest.en)
			}
		})
	}
}

func TestBePort(t *testing.T) {
	namer := NewNamer("560b5b8154db29b3", "fooFN")

	fooTests := []struct {
		tName string
		bn    string
		ie    bool
		bp    string
	}{
		{"empty_check", "", true, ""},
		{"not_matched_check", "fooBP", true, ""},
		{"unable_to_lookup_port", "k8s-be-", true, ""},
		{"unexpected_regular_match", "k8s-be-12345678901234567890123456789", true, ""},
		{"normal_test", "k8s-be-81", false, "81"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			rbp, err := namer.BePort(fooTest.bn)
			if fooTest.ie {
				if err == nil {
					t.Error("Expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if rbp != fooTest.bp {
					t.Errorf("Returned %v but expected %v", rbp, fooTest.bp)
				}
			}
		})
	}
}

func TestIGName(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		en    string
	}{
		{"empty_clusterName", "", "k8s-ig"},
		{"normal_clusterName", "560b5b8154db29b3", "k8s-ig--560b5b8154db29b3"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.cn, "fw-fooFN")
			rn := namer.IGName()
			if rn != fooTest.en {
				t.Errorf("Returned %v but expected %v", rn, fooTest.en)
			}
		})
	}
}

func TestFrSuffix(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		fn    string
		efn   string
	}{
		{"empty_clusterNameAndFireName", "", "", "l7"},
		{"empty_fireName", "560b5b8154db29b3", "", "l7--560b5b8154db29b3"},
		{"normal_fireName", "560b5b8154db29b3", "fooFN", "l7--fooFN"},
		{"will_be_truncated", "560b5b8154db29b3", "foo-e2e-ingress-firewall-xyttxtps-zsy13-auth--560b5b8154db29b3", "l7--foo-e2e-ingress-firewall-xyttxtps-zsy13-auth--560b5b8154db0"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.cn, fooTest.fn)
			rfn := namer.FrSuffix()
			if rfn != fooTest.efn {
				t.Errorf("Returned %v but expected %v", rfn, fooTest.efn)
			}
		})
	}
}

func TestFrName(t *testing.T) {
	fooTests := []struct {
		tName string
		s     string
		efn   string
	}{
		{"empty_fireName", "", "k8s-fw-"},
		{"normal_fireName", "fooFN", "k8s-fw-fooFN"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer("fooCN", "fooFN")
			rfn := namer.FrName(fooTest.s)
			if rfn != fooTest.efn {
				t.Errorf("Returned %v but expected %v", rfn, fooTest.efn)
			}
		})
	}
}

func TestLBName(t *testing.T) {
	fooTests := []struct {
		tName string
		cn    string
		k     string
		elbn  string
	}{
		{"empty_clusterNameAndKey", "", "", ""},
		{"empty_clusterName", "", "default/k8s-tps-e2e-ingress-deployment", "default-k8s-tps-e2e-ingress-deployment"},
		{"end_with_cluster_name", "560b5b8154db29b3", "default/k8s-tps-e2e-ingress-deployment--560b5b8154db29b3", "default-k8s-tps-e2e-ingress-deployment--560b5b8154db29b3"},
		{"not_end_with_cluster_name", "560b5b8154db29b3", "default/k8s-tps-e2e-ingress-deployment", "default-k8s-tps-e2e-ingress-deployment--560b5b8154db29b3"},
		{"will_be_turncted", "560b5b8154db29b3", "default/k8s-tps-e2e-ingress-deployment-xytps-auth", "default-k8s-tps-e2e-ingress-deployment-xytps-auth--560b5b8154d0"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			namer := NewNamer(fooTest.cn, "fooFN")
			rlbn := namer.LBName(fooTest.k)
			if rlbn != fooTest.elbn {
				t.Errorf("Returned %v but expected %v", rlbn, fooTest.elbn)
			}
		})
	}
}

func TestGetDefaultBackend(t *testing.T) {
	var nilBS *compute.BackendService
	bs := &compute.BackendService{CreationTimestamp: "2017-03-08"}

	fooTests := []struct {
		tName string
		gm    GCEURLMap
		ebs   *compute.BackendService
	}{
		{"empty_GCEURLMap", map[string]map[string]*compute.BackendService{}, nilBS},
		{"include_default_but_empty", map[string]map[string]*compute.BackendService{DefaultBackendKey: {}}, nilBS},
		{"inluce_other_keys_for_default_backend", map[string]map[string]*compute.BackendService{DefaultBackendKey: {DefaultBackendKey + "_non": bs}}, nilBS},
		{"defalut_backend_has_been_exist", map[string]map[string]*compute.BackendService{DefaultBackendKey: {DefaultBackendKey: bs}}, bs},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			rbs := fooTest.gm.GetDefaultBackend()
			if rbs != fooTest.ebs {
				t.Errorf("Returned %v but expected %v", rbs, fooTest.ebs)
			}
		})
	}
}

func TestString(t *testing.T) {
	fooTests := []struct {
		tName string
		gm    GCEURLMap
		es    string
	}{
		{"empty_GCEURLMap", map[string]map[string]*compute.BackendService{}, ""},
		{"host_exist_url_not_exist", map[string]map[string]*compute.BackendService{"k8s-50xyoeu02x": {}}, "k8s-50xyoeu02x\n"},
		{"backend_not_exist", map[string]map[string]*compute.BackendService{"k8s-50xyoeu02x": {"http://10.0.0.1/lb": nil}}, "k8s-50xyoeu02x\n\thttp://10.0.0.1/lb: No backend\n"},
		{"single_host", map[string]map[string]*compute.BackendService{"k8s-50xyoeu02x": {"http://10.0.0.1/lb": {Name: "foo"}}}, "k8s-50xyoeu02x\n\thttp://10.0.0.1/lb: foo\n"},
		{"multi_host", map[string]map[string]*compute.BackendService{"k8s-50xyoeu02x": {"http://10.0.0.1/lb": {Name: "foo1"}}, "k8s-8xiy02zis": {"http://10.0.0.2/lb": {Name: "foo2"}}}, "k8s-50xyoeu02x\n\thttp://10.0.0.1/lb: foo1\nk8s-8xiy02zis\n\thttp://10.0.0.2/lb: foo2\n"},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			rs := fooTest.gm.String()
			if rs != fooTest.es {
				t.Errorf("Returned\n%v\nbut expected\n%v", rs, fooTest.es)
			}
		})
	}
}

func TestPutDefaultBackend(t *testing.T) {
	obs := &compute.BackendService{CreationTimestamp: "2017-03-28"}
	nbs := &compute.BackendService{Name: "foo"}

	fooTests := []struct {
		tName string
		gm    GCEURLMap
	}{
		{"empty_GCEURLMap", map[string]map[string]*compute.BackendService{}},
		{"include_default_but_empty", map[string]map[string]*compute.BackendService{DefaultBackendKey: {}}},
		{"inluce_other_keys_for_default_backend", map[string]map[string]*compute.BackendService{DefaultBackendKey: {"not_default_backend_key": obs}}},
		{"defalut_backend_has_been_exist", map[string]map[string]*compute.BackendService{DefaultBackendKey: {DefaultBackendKey: obs}}},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			fooTest.gm.PutDefaultBackend(nbs)
			rbs := fooTest.gm.GetDefaultBackend()
			if !reflect.DeepEqual(rbs, nbs) {
				t.Errorf("Returned %v but expected %v", rbs, nbs)
			}
		})
	}
}

func TestIsHTTPErrorCode(t *testing.T) {
	fooTests := []struct {
		tName string
		err   error
		code  int
		eb    bool
	}{
		{"not_googleapi_error", &fakeError{}, 202, false},
		{"not_the same_error_code", &googleapi.Error{Code: 202}, 208, false},
		{"the_same_error_code", &googleapi.Error{Code: 202}, 202, true},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			rb := IsHTTPErrorCode(fooTest.err, fooTest.code)
			if rb != fooTest.eb {
				t.Errorf("Returned %v but expected %v", rb, fooTest.eb)
			}
		})
	}
}

func TestCompareLinks(t *testing.T) {
	fooTests := []struct {
		tName string
		l1    string
		l2    string
		eb    bool
	}{
		{"empty_links_compare", "", "", false},
		{"empty_link1_compare", "", "http://ip:port/link2", false},
		{"empty_link2_compare", "http://ip:port/link1", "", false},
		{"equaled_links_compare", "http://ip:port/link", "http://ip:port/link", true},
		{"not_euqaled_links_compare", "http://ip:port/link1", "http://ip:port/link2", false},
	}

	for _, fooTest := range fooTests {
		t.Run(fooTest.tName, func(t *testing.T) {
			rb := CompareLinks(fooTest.l1, fooTest.l2)
			if rb != fooTest.eb {
				t.Errorf("Returned %v but expected %v", rb, fooTest.eb)
			}
		})
	}
}
