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

package inspector

import (
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

// DeepInspect is the function called by admissionwebhook and store syncer to check
// if an object contains invalid configurations that may represent a security risk,
// and returning an error in this case
func DeepInspect(obj interface{}) error {
	switch obj := obj.(type) {
	case *networking.Ingress:
		return InspectIngress(obj)
	case *corev1.Service:
		return InspectService(obj)
	default:
		klog.Warningf("received invalid object to inspect: %T", obj)
		return nil
	}
}
