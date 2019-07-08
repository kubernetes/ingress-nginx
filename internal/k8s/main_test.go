/*
Copyright 2015 The Kubernetes Authors.

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

package k8s

import (
	"os"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestParseNameNS(t *testing.T) {
	tests := []struct {
		title  string
		input  string
		ns     string
		name   string
		expErr bool
	}{
		{"empty string", "", "", "", true},
		{"demo", "demo", "", "", true},
		{"kube-system", "kube-system", "", "", true},
		{"default/kube-system", "default/kube-system", "default", "kube-system", false},
	}

	for _, test := range tests {
		ns, name, err := ParseNameNS(test.input)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but returned nil", test.title)
			}
			continue
		}
		if test.ns != ns {
			t.Errorf("%v: expected %v but returned %v", test.title, test.ns, ns)
		}
		if test.name != name {
			t.Errorf("%v: expected %v but returned %v", test.title, test.name, name)
		}
	}
}

func TestGetNodeIP(t *testing.T) {
	fKNodes := []struct {
		cs *testclient.Clientset
		n  string
		ea string
		i  bool
	}{
		// empty node list
		{testclient.NewSimpleClientset(), "demo", "", true},

		// node not exist
		{testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Status: apiv1.NodeStatus{
				Addresses: []apiv1.NodeAddress{
					{
						Type:    apiv1.NodeInternalIP,
						Address: "10.0.0.1",
					},
				},
			},
		}}}), "notexistnode", "", true},

		// node exist
		{testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Status: apiv1.NodeStatus{
				Addresses: []apiv1.NodeAddress{
					{
						Type:    apiv1.NodeInternalIP,
						Address: "10.0.0.1",
					},
				},
			},
		}}}), "demo", "10.0.0.1", true},

		// search the correct node
		{testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "demo1",
				},
				Status: apiv1.NodeStatus{
					Addresses: []apiv1.NodeAddress{
						{
							Type:    apiv1.NodeInternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "demo2",
				},
				Status: apiv1.NodeStatus{
					Addresses: []apiv1.NodeAddress{
						{
							Type:    apiv1.NodeInternalIP,
							Address: "10.0.0.2",
						},
					},
				},
			},
		}}), "demo2", "10.0.0.2", true},

		// get NodeExternalIP
		{testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Status: apiv1.NodeStatus{
				Addresses: []apiv1.NodeAddress{
					{
						Type:    apiv1.NodeInternalIP,
						Address: "10.0.0.1",
					}, {
						Type:    apiv1.NodeExternalIP,
						Address: "10.0.0.2",
					},
				},
			},
		}}}), "demo", "10.0.0.2", false},

		// get NodeInternalIP
		{testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Status: apiv1.NodeStatus{
				Addresses: []apiv1.NodeAddress{
					{
						Type:    apiv1.NodeExternalIP,
						Address: "",
					}, {
						Type:    apiv1.NodeInternalIP,
						Address: "10.0.0.2",
					},
				},
			},
		}}}), "demo", "10.0.0.2", true},
	}

	for _, fk := range fKNodes {
		address := GetNodeIPOrName(fk.cs, fk.n, fk.i)
		if address != fk.ea {
			t.Errorf("expected %s, but returned %s", fk.ea, address)
		}
	}
}

func TestGetPodDetails(t *testing.T) {
	// POD_NAME & POD_NAMESPACE not exist
	os.Setenv("POD_NAME", "")
	os.Setenv("POD_NAMESPACE", "")
	_, err1 := GetPodDetails(testclient.NewSimpleClientset())
	if err1 == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD_NAME not exist
	os.Setenv("POD_NAME", "")
	os.Setenv("POD_NAMESPACE", apiv1.NamespaceDefault)
	_, err2 := GetPodDetails(testclient.NewSimpleClientset())
	if err2 == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD_NAMESPACE not exist
	os.Setenv("POD_NAME", "testpod")
	os.Setenv("POD_NAMESPACE", "")
	_, err3 := GetPodDetails(testclient.NewSimpleClientset())
	if err3 == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD not exist
	os.Setenv("POD_NAME", "testpod")
	os.Setenv("POD_NAMESPACE", apiv1.NamespaceDefault)
	_, err4 := GetPodDetails(testclient.NewSimpleClientset())
	if err4 == nil {
		t.Errorf("expected an error but returned nil")
	}

	// success to get PodInfo
	fkClient := testclient.NewSimpleClientset(
		&apiv1.PodList{Items: []apiv1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testpod",
				Namespace: apiv1.NamespaceDefault,
				Labels: map[string]string{
					"first":  "first_label",
					"second": "second_label",
				},
			},
		}}},
		&apiv1.NodeList{Items: []apiv1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Status: apiv1.NodeStatus{
				Addresses: []apiv1.NodeAddress{
					{
						Type:    apiv1.NodeInternalIP,
						Address: "10.0.0.1",
					},
				},
			},
		}}})

	epi, err5 := GetPodDetails(fkClient)
	if err5 != nil {
		t.Errorf("expected a PodInfo but returned error")
		return
	}

	if epi == nil {
		t.Errorf("expected a PodInfo but returned nil")
	}
}
