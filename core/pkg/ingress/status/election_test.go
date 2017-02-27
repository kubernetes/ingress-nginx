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

package status

import (
	"encoding/json"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	tc "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"k8s.io/kubernetes/pkg/client/leaderelection/resourcelock"
)

func TestGetCurrentLeaderLeaderExist(t *testing.T) {
	fkER := resourcelock.LeaderElectionRecord{
		HolderIdentity:       "currentLeader",
		LeaseDurationSeconds: 30,
		AcquireTime:          unversioned.Now(),
		RenewTime:            unversioned.Now(),
		LeaderTransitions:    3,
	}
	leaderInfo, _ := json.Marshal(fkER)
	fkEndpoints := api.Endpoints{
		ObjectMeta: api.ObjectMeta{
			Name:      "ingress-controller-test",
			Namespace: api.NamespaceSystem,
			Annotations: map[string]string{
				resourcelock.LeaderElectionRecordAnnotationKey: string(leaderInfo),
			},
		},
	}
	fk := tc.NewSimpleClientset(&api.EndpointsList{Items: []api.Endpoints{fkEndpoints}})
	identity, endpoints, err := getCurrentLeader("ingress-controller-test", api.NamespaceSystem, fk)
	if err != nil {
		t.Fatalf("expected identitiy and endpoints but returned error %s", err)
	}

	if endpoints == nil {
		t.Fatalf("returned nil but expected an endpoints")
	}

	if identity != "currentLeader" {
		t.Fatalf("returned %v but expected %v", identity, "currentLeader")
	}
}

func TestGetCurrentLeaderLeaderNotExist(t *testing.T) {
	fkEndpoints := api.Endpoints{
		ObjectMeta: api.ObjectMeta{
			Name:        "ingress-controller-test",
			Namespace:   api.NamespaceSystem,
			Annotations: map[string]string{},
		},
	}
	fk := tc.NewSimpleClientset(&api.EndpointsList{Items: []api.Endpoints{fkEndpoints}})
	identity, endpoints, err := getCurrentLeader("ingress-controller-test", api.NamespaceSystem, fk)
	if err != nil {
		t.Fatalf("unexpeted error: %v", err)
	}

	if endpoints == nil {
		t.Fatalf("returned nil but expected an endpoints")
	}

	if identity != "" {
		t.Fatalf("returned %s but expected %s", identity, "")
	}
}

func TestGetCurrentLeaderAnnotationError(t *testing.T) {
	fkEndpoints := api.Endpoints{
		ObjectMeta: api.ObjectMeta{
			Name:      "ingress-controller-test",
			Namespace: api.NamespaceSystem,
			Annotations: map[string]string{
				resourcelock.LeaderElectionRecordAnnotationKey: "just-test-error-leader-annotation",
			},
		},
	}
	fk := tc.NewSimpleClientset(&api.EndpointsList{Items: []api.Endpoints{fkEndpoints}})
	_, _, err := getCurrentLeader("ingress-controller-test", api.NamespaceSystem, fk)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestNewElection(t *testing.T) {
	fk := tc.NewSimpleClientset(&api.EndpointsList{Items: []api.Endpoints{
		{
			ObjectMeta: api.ObjectMeta{
				Name:      "ingress-controller-test",
				Namespace: api.NamespaceSystem,
			},
		},
		{
			ObjectMeta: api.ObjectMeta{
				Name:      "ingress-controller-test-020",
				Namespace: api.NamespaceSystem,
			},
		},
	}})

	ne, err := NewElection("ingress-controller-test", "startLeader", api.NamespaceSystem, 4*time.Second, func(leader string) {
		// do nothing
		go t.Logf("execute callback fun, leader is: %s", leader)
	}, fk)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if ne == nil {
		t.Fatalf("unexpected nil")
	}
}
