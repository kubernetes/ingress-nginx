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

	"k8s.io/api/admission/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	review := &v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Resource: v1.GroupVersionResource{Group: "", Version: "v1", Resource: "pod"},
		},
	}

	adm.HandleAdmission(review)
	if review.Response.Allowed {
		t.Fatalf("with a non ingress resource, the check should not pass")
	}

	review.Request.Resource = v1.GroupVersionResource{Group: networking.GroupName, Version: "v1beta1", Resource: "ingresses"}
	review.Request.Object.Raw = []byte{0xff}

	adm.HandleAdmission(review)
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
