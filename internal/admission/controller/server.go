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
	"io/ioutil"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

// AdmissionController checks if an object
// is allowed in the cluster
type AdmissionController interface {
	HandleAdmission(*v1beta1.AdmissionReview)
}

// AdmissionControllerServer implements an HTTP server
// for kubernetes validating webhook
// https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook
type AdmissionControllerServer struct {
	AdmissionController AdmissionController
	Decoder             runtime.Decoder
}

// NewAdmissionControllerServer instanciates an admission controller server with
// a default codec
func NewAdmissionControllerServer(ac AdmissionController) *AdmissionControllerServer {
	return &AdmissionControllerServer{
		AdmissionController: ac,
		Decoder:             codecs.UniversalDeserializer(),
	}
}

// ServeHTTP implements http.Server method
func (acs *AdmissionControllerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	review, err := parseAdmissionReview(acs.Decoder, r.Body)
	if err != nil {
		klog.Errorf("Unexpected error decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	acs.AdmissionController.HandleAdmission(review)
	if err := writeAdmissionReview(w, review); err != nil {
		klog.Errorf("Unexpected returning admission review: %v", err)
	}
}

func parseAdmissionReview(decoder runtime.Decoder, r io.Reader) (*v1beta1.AdmissionReview, error) {
	review := &v1beta1.AdmissionReview{}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	_, _, err = decoder.Decode(data, nil, review)
	return review, err
}

func writeAdmissionReview(w io.Writer, ar *v1beta1.AdmissionReview) error {
	e := json.NewEncoder(w)
	return e.Encode(ar)
}
