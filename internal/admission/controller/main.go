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
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	networking "k8s.io/api/networking/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog/v2"
)

// Checker must return an error if the ingress provided as argument
// contains invalid instructions
type Checker interface {
	CheckIngress(ing *networking.Ingress) error
}

// IngressAdmission implements the AdmissionController interface
// to handle Admission Reviews and deny requests that are not validated
type IngressAdmission struct {
	Checker Checker
}

var (
	ingressResource = metav1.GroupVersionKind{
		Group:   networking.GroupName,
		Version: "v1",
		Kind:    "Ingress",
	}
)

// HandleAdmission populates the admission Response
// with Allowed=false if the Object is an ingress that would prevent nginx to reload the configuration
// with Allowed=true otherwise
func (ia *IngressAdmission) HandleAdmission(obj runtime.Object) (runtime.Object, error) {

	review, isV1 := obj.(*admissionv1.AdmissionReview)
	if !isV1 {
		return nil, fmt.Errorf("request is not of type AdmissionReview v1 or v1beta1")
	}

	if !apiequality.Semantic.DeepEqual(review.Request.Kind, ingressResource) {
		return nil, fmt.Errorf("rejecting admission review because the request does not contain an Ingress resource but %s with name %s in namespace %s",
			review.Request.Kind.String(), review.Request.Name, review.Request.Namespace)
	}

	status := &admissionv1.AdmissionResponse{}
	status.UID = review.Request.UID

	ingress := networking.Ingress{}

	codec := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Pretty: true,
	})
	_, _, err := codec.Decode(review.Request.Object.Raw, nil, &ingress)
	if err != nil {
		klog.ErrorS(err, "failed to decode ingress")
		status.Allowed = false
		status.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}

		review.Response = status
		return review, nil
	}

	if err := ia.Checker.CheckIngress(&ingress); err != nil {
		klog.ErrorS(err, "invalid ingress configuration", "ingress", fmt.Sprintf("%v/%v", review.Request.Namespace, review.Request.Name))
		status.Allowed = false
		status.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}

		review.Response = status
		return review, nil
	}

	klog.InfoS("successfully validated configuration, accepting", "ingress", fmt.Sprintf("%v/%v", review.Request.Namespace, review.Request.Name))
	status.Allowed = true
	review.Response = status

	return review, nil
}
