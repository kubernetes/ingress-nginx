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

// Storage backends used by the Ingress controller.
// Ingress controllers require their own storage for the following reasons:
// 1. There is only so much information we can pack into 64 chars allowed
//    by GCE for resource names.
// 2. An Ingress controller cannot assume total control over a project, in
//    fact in a majority of cases (ubernetes, tests, multiple gke clusters in
//    same project) there *will* be multiple controllers in a project.
// 3. If the Ingress controller pod is killed, an Ingress is deleted while
//    the pod is down, and then the controller is re-scheduled on another node,
//    it will leak resources. Note that this will happen today because
//    the only implemented storage backend is InMemoryPool.
// 4. Listing from cloudproviders is really slow.

package storage
