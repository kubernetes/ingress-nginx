/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	apierrs "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/util/workqueue"
)

// StoreToIngressLister makes a Store that lists Ingress.
type StoreToIngressLister struct {
	cache.Store
}

// taskQueue manages a work queue through an independent worker that
// invokes the given sync function for every work item inserted.
type taskQueue struct {
	// queue is the work queue the worker polls
	queue *workqueue.Type
	// sync is called for each item in the queue
	sync func(string)
	// workerDone is closed when the worker exits
	workerDone chan struct{}
}

func (t *taskQueue) run(period time.Duration, stopCh <-chan struct{}) {
	wait.Until(t.worker, period, stopCh)
}

// enqueue enqueues ns/name of the given api object in the task queue.
func (t *taskQueue) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		glog.Infof("could not get key for object %+v: %v", obj, err)
		return
	}
	t.queue.Add(key)
}

func (t *taskQueue) requeue(key string, err error) {
	glog.V(3).Infof("requeuing %v, err %v", key, err)
	t.queue.Add(key)
}

// worker processes work in the queue through sync.
func (t *taskQueue) worker() {
	for {
		key, quit := t.queue.Get()
		if quit {
			close(t.workerDone)
			return
		}
		glog.V(3).Infof("syncing %v", key)
		t.sync(key.(string))
		t.queue.Done(key)
	}
}

// shutdown shuts down the work queue and waits for the worker to ACK
func (t *taskQueue) shutdown() {
	t.queue.ShutDown()
	<-t.workerDone
}

// NewTaskQueue creates a new task queue with the given sync function.
// The sync function is called for every element inserted into the queue.
func NewTaskQueue(syncFn func(string)) *taskQueue {
	return &taskQueue{
		queue:      workqueue.New(),
		sync:       syncFn,
		workerDone: make(chan struct{}),
	}
}

// getPodDetails  returns runtime information about the pod: name, namespace and IP of the node
func getPodDetails(kubeClient *unversioned.Client) (*podInfo, error) {
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	err := waitForPodRunning(kubeClient, podNs, podName, time.Millisecond*200, time.Second*30)
	if err != nil {
		return nil, err
	}

	pod, _ := kubeClient.Pods(podNs).Get(podName)
	if pod == nil {
		return nil, fmt.Errorf("Unable to get POD information")
	}

	node, err := kubeClient.Nodes().Get(pod.Spec.NodeName)
	if err != nil {
		return nil, err
	}

	var externalIP string
	for _, address := range node.Status.Addresses {
		if address.Type == api.NodeExternalIP {
			if address.Address != "" {
				externalIP = address.Address
				break
			}
		}

		if externalIP == "" && address.Type == api.NodeLegacyHostIP {
			externalIP = address.Address
		}
	}

	return &podInfo{
		PodName:      podName,
		PodNamespace: podNs,
		NodeIP:       externalIP,
	}, nil
}

func isValidService(kubeClient *unversioned.Client, name string) error {
	if name == "" {
		return fmt.Errorf("empty string is not a valid service name")
	}

	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid name format (namespace/name) in service '%v'", name)
	}

	_, err := kubeClient.Services(parts[0]).Get(parts[1])
	return err
}

func isHostValid(host string, cns []string) bool {
	for _, cn := range cns {
		if matchHostnames(cn, host) {
			return true
		}
	}

	return false
}

func matchHostnames(pattern, host string) bool {
	host = strings.TrimSuffix(host, ".")
	pattern = strings.TrimSuffix(pattern, ".")

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

func parseNsName(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
}

func waitForPodRunning(kubeClient *unversioned.Client, ns, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodRunning {
			return true, nil
		}
		return false, nil
	}

	return waitForPodCondition(kubeClient, ns, podName, condition, interval, timeout)
}

// waitForPodCondition waits for a pod in state defined by a condition (func)
func waitForPodCondition(kubeClient *unversioned.Client, ns, podName string, condition func(pod *api.Pod) (bool, error),
	interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pod, err := kubeClient.Pods(ns).Get(podName)
		if err != nil {
			if apierrs.IsNotFound(err) {
				return false, nil
			}
		}

		done, err := condition(pod)
		if err != nil {
			return false, err
		}
		if done {
			return true, nil
		}

		return false, nil
	})
}
