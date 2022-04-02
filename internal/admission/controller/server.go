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
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog/v2"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	admissionv1.AddToScheme(scheme)
}

// AdmissionController checks if an object
// is allowed in the cluster
type AdmissionController interface {
	HandleAdmission(runtime.Object) (runtime.Object, error)
}

// AdmissionControllerServer implements an HTTP server
// for kubernetes validating webhook
// https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook
type AdmissionControllerServer struct {
	AdmissionController AdmissionController
}

// NewAdmissionControllerServer instanciates an admission controller server with
// a default codec
func NewAdmissionControllerServer(ac AdmissionController) *AdmissionControllerServer {
	return &AdmissionControllerServer{
		AdmissionController: ac,
	}
}

// ServeHTTP implements http.Server method
func (acs *AdmissionControllerServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		klog.ErrorS(err, "Failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	codec := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Pretty: true,
	})

	obj, _, err := codec.Decode(data, nil, nil)
	if err != nil {
		klog.ErrorS(err, "Failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := acs.AdmissionController.HandleAdmission(obj)
	if err != nil {
		klog.ErrorS(err, "failed to process webhook request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := codec.Encode(result, w); err != nil {
		klog.ErrorS(err, "failed to encode response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
