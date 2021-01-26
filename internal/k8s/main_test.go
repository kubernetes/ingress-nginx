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
		name          string
		cs            *testclient.Clientset
		nodeName      string
		ea            string
		useInternalIP bool
	}{
		{
			"empty node list",
			testclient.NewSimpleClientset(),
			"demo", "", true,
		},
		{
			"node does not exist",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
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
			}}}), "notexistnode", "", true,
		},
		{
			"node exist and only has an internal IP address (useInternalIP=false)",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
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
			}}}), "demo", "10.0.0.1", false,
		},
		{
			"node exist and only has an internal IP address",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
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
			}}}), "demo", "10.0.0.1", true,
		},
		{
			"node exist and only has an external IP address",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "demo",
				},
				Status: apiv1.NodeStatus{
					Addresses: []apiv1.NodeAddress{
						{
							Type:    apiv1.NodeExternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			}}}), "demo", "10.0.0.1", false,
		},
		{
			"multiple nodes - choose the right one",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{
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
			}}),
			"demo2", "10.0.0.2", true,
		},
		{
			"node with both IP internal and external IP address - returns external IP",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
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
			}}}),
			"demo", "10.0.0.2", false,
		},
		{
			"node with both IP internal and external IP address - returns internal IP",
			testclient.NewSimpleClientset(&apiv1.NodeList{Items: []apiv1.Node{{
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
			}}}),
			"demo", "10.0.0.2", true},
	}

	for _, fk := range fKNodes {
		address := GetNodeIPOrName(fk.cs, fk.nodeName, fk.useInternalIP)
		if address != fk.ea {
			t.Errorf("%v - expected %s, but returned %s", fk.name, fk.ea, address)
		}
	}
}

func TestGetIngressPod(t *testing.T) {
	// POD_NAME & POD_NAMESPACE not exist
	os.Setenv("POD_NAME", "")
	os.Setenv("POD_NAMESPACE", "")
	err := GetIngressPod(testclient.NewSimpleClientset())
	if err == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD_NAME not exist
	os.Setenv("POD_NAME", "")
	os.Setenv("POD_NAMESPACE", apiv1.NamespaceDefault)
	err = GetIngressPod(testclient.NewSimpleClientset())
	if err == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD_NAMESPACE not exist
	os.Setenv("POD_NAME", "testpod")
	os.Setenv("POD_NAMESPACE", "")
	err = GetIngressPod(testclient.NewSimpleClientset())
	if err == nil {
		t.Errorf("expected an error but returned nil")
	}

	// POD not exist
	os.Setenv("POD_NAME", "testpod")
	os.Setenv("POD_NAMESPACE", apiv1.NamespaceDefault)
	err = GetIngressPod(testclient.NewSimpleClientset())
	if err == nil {
		t.Errorf("expected an error but returned nil")
	}

	// success to get PodInfo
	fkClient := testclient.NewSimpleClientset(
		&apiv1.PodList{Items: []apiv1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testpod",
				Namespace: apiv1.NamespaceDefault,
				Labels: map[string]string{
					"first":                       "first_label",
					"second":                      "second_label",
					"app.kubernetes.io/component": "controller",
					"app.kubernetes.io/instance":  "ingress-nginx",
					"app.kubernetes.io/name":      "ingress-nginx",
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

	err = GetIngressPod(fkClient)
	if err != nil {
		t.Errorf("expected a PodInfo but returned error: %v", err)
		return
	}
}
