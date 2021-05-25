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
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	networking "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

const testIngressName = "testIngressName"

type failTestChecker struct {
	t *testing.T
}

func (ftc failTestChecker) CheckIngress(ing *networking.Ingress) error {
	ftc.t.Error("checker should not be called")
	return nil
}

type testChecker struct {
	t   *testing.T
	err error
}

func (tc testChecker) CheckIngress(ing *networking.Ingress) error {
	if ing.ObjectMeta.Name != testIngressName {
		tc.t.Errorf("CheckIngress should be called with %v ingress, but got %v", testIngressName, ing.ObjectMeta.Name)
	}
	return tc.err
}

func TestHandleAdmission(t *testing.T) {
	adm := &IngressAdmission{
		Checker: failTestChecker{t: t},
	}

	result, err := adm.HandleAdmission(&admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			Kind: v1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		},
	})
	if err == nil {
		t.Fatalf("with a non ingress resource, the check should not pass")
	}

	result, err = adm.HandleAdmission(nil)
	if err == nil {
		t.Fatalf("with a nil AdmissionReview request, the check should not pass")
	}

	result, err = adm.HandleAdmission(&admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			Kind: v1.GroupVersionKind{Group: networking.GroupName, Version: "v1", Kind: "Ingress"},
			Object: runtime.RawExtension{
				Raw: []byte{0xff},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	review, isV1 := (result).(*admissionv1.AdmissionReview)
	if !isV1 {
		t.Fatalf("expected AdmissionReview V1 object but %T returned", result)
	}

	if review.Response.Allowed {
		t.Fatalf("when the request object is not decodable, the request should not be allowed")
	}

	raw, err := json.Marshal(networking.Ingress{ObjectMeta: v1.ObjectMeta{Name: testIngressName}})
	if err != nil {
		t.Fatalf("failed to prepare test ingress data: %v", err.Error())
	}

	review.Request.Object.Raw = raw

	adm.Checker = testChecker{
		t:   t,
		err: fmt.Errorf("this is a test error"),
	}

	adm.HandleAdmission(review)
	if review.Response.Allowed {
		t.Fatalf("when the checker returns an error, the request should not be allowed")
	}

	adm.Checker = testChecker{
		t:   t,
		err: nil,
	}

	adm.HandleAdmission(review)
	if !review.Response.Allowed {
		t.Fatalf("when the checker returns no error, the request should be allowed")
	}
}
