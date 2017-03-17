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

package status

import (
	"encoding/json"
	"os"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	def_api "k8s.io/client-go/pkg/api"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/ingress/core/pkg/ingress/status/leaderelection"
	"k8s.io/ingress/core/pkg/ingress/status/leaderelection/resourcelock"
)

func getCurrentLeader(electionID, namespace string, c client.Interface) (string, *api.Endpoints, error) {
	endpoints, err := c.Core().Endpoints(namespace).Get(electionID, meta_v1.GetOptions{})
	if err != nil {
		return "", nil, err
	}
	val, found := endpoints.Annotations[resourcelock.LeaderElectionRecordAnnotationKey]
	if !found {
		return "", endpoints, nil
	}
	electionRecord := resourcelock.LeaderElectionRecord{}
	if err = json.Unmarshal([]byte(val), &electionRecord); err != nil {
		return "", nil, err
	}
	return electionRecord.HolderIdentity, endpoints, err
}

// NewElection creates an election.  'namespace'/'election' should be an existing Kubernetes Service
// 'id' is the id if this leader, should be unique.
func NewElection(electionID,
	id,
	namespace string,
	ttl time.Duration,
	callback func(leader string),
	c client.Interface) (*leaderelection.LeaderElector, error) {

	_, err := c.Core().Endpoints(namespace).Get(electionID, meta_v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.Core().Endpoints(namespace).Create(&api.Endpoints{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: electionID,
				},
			})
			if err != nil && !errors.IsConflict(err) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(stop <-chan struct{}) {
			callback(id)
		},
		OnStoppedLeading: func() {
			leader, _, err := getCurrentLeader(electionID, namespace, c)
			if err != nil {
				glog.Errorf("failed to get leader: %v", err)
				// empty string means leader is unknown
				callback("")
				return
			}
			callback(leader)
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	recorder := broadcaster.NewRecorder(def_api.Scheme, api.EventSource{
		Component: "ingress-leader-elector",
		Host:      hostname,
	})

	lock := resourcelock.EndpointsLock{
		EndpointsMeta: meta_v1.ObjectMeta{Namespace: namespace, Name: electionID},
		Client:        c,
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	}

	config := leaderelection.LeaderElectionConfig{
		Lock:          &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callbacks,
	}

	return leaderelection.NewLeaderElector(config)
}
