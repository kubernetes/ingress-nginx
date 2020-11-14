/*
Copyright 2019 The Kubernetes Authors.

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

package controller

import (
	"context"
	"os"
	"time"

	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/klog/v2"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

type leaderElectionConfig struct {
	Client clientset.Interface

	ElectionID string

	OnStartedLeading func(chan struct{})
	OnStoppedLeading func()
}

func setupLeaderElection(config *leaderElectionConfig) {
	var elector *leaderelection.LeaderElector

	// start a new context
	ctx := context.Background()

	var cancelContext context.CancelFunc

	var newLeaderCtx = func(ctx context.Context) context.CancelFunc {
		// allow to cancel the context in case we stop being the leader
		leaderCtx, cancel := context.WithCancel(ctx)
		go elector.Run(leaderCtx)
		return cancel
	}

	var stopCh chan struct{}
	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			klog.V(2).InfoS("I am the new leader")
			stopCh = make(chan struct{})

			if config.OnStartedLeading != nil {
				config.OnStartedLeading(stopCh)
			}
		},
		OnStoppedLeading: func() {
			klog.V(2).InfoS("I am not leader anymore")
			close(stopCh)

			// cancel the context
			cancelContext()

			cancelContext = newLeaderCtx(ctx)

			if config.OnStoppedLeading != nil {
				config.OnStoppedLeading()
			}
		},
		OnNewLeader: func(identity string) {
			klog.InfoS("New leader elected", "identity", identity)
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, _ := os.Hostname()

	recorder := broadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{
		Component: "ingress-leader-elector",
		Host:      hostname,
	})

	lock := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{Namespace: k8s.IngressPodDetails.Namespace, Name: config.ElectionID},
		Client:        config.Client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      k8s.IngressPodDetails.Name,
			EventRecorder: recorder,
		},
	}

	ttl := 30 * time.Second

	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,

		Callbacks: callbacks,
	})
	if err != nil {
		klog.Fatalf("unexpected error starting leader election: %v", err)
	}

	cancelContext = newLeaderCtx(ctx)
}
