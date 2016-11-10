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

// This is the structure of the gce l7 controller:
// apiserver <-> controller ---> pools --> cloud
//                  |               |
//                  |-> Ingress     |-> backends
//                  |-> Services    |   |-> health checks
//                  |-> Nodes       |
//                                  |-> instance groups
//                                  |   |-> port per backend
//                                  |
//                                  |-> loadbalancers
//                                      |-> http proxy
//                                      |-> forwarding rule
//                                      |-> urlmap
// * apiserver: kubernetes api serer.
// * controller: gce l7 controller, watches apiserver and interacts
//	with sync pools. The controller doesn't know anything about the cloud.
//  Communication between the controller and pools is 1 way.
// * pool: the controller tells each pool about desired state by inserting
//	into shared memory store. The pools sync this with the cloud. Pools are
//  also responsible for periodically checking the edge links between various
//	cloud resources.
//
// A note on sync pools: this package has 3 sync pools: for node, instances and
// loadbalancer resources. A sync pool is meant to record all creates/deletes
// performed by a controller and periodically verify that links are not broken.
// For example, the controller might create a backend via backendPool.Add(),
// the backend pool remembers this and continuously verifies that the backend
// is connected to the right instance group, and that the instance group has
// the right ports open.
//
// A note on naming convention: per golang style guide for Initialisms, Http
// should be HTTP and Url should be URL, however because these interfaces
// must match their siblings in the Kubernetes cloud provider, which are in turn
// consistent with GCE compute API, there might be inconsistencies.

package controller
