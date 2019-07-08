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
	"github.com/google/uuid"
	"k8s.io/api/admission/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/klog"
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

// HandleAdmission populates the admission Response
// with Allowed=false if the Object is an ingress that would prevent nginx to reload the configuration
// with Allowed=true otherwise
func (ia *IngressAdmission) HandleAdmission(ar *v1beta1.AdmissionReview) error {
	if ar.Request == nil {
		klog.Infof("rejecting nil request")
		ar.Response = &v1beta1.AdmissionResponse{
			UID:     types.UID(uuid.New().String()),
			Allowed: false,
		}
		return nil
	}
	klog.V(3).Infof("handling ingress admission webhook request for {%s}  %s in namespace %s", ar.Request.Resource.String(), ar.Request.Name, ar.Request.Namespace)

	ingressResource := v1.GroupVersionResource{Group: networking.SchemeGroupVersion.Group, Version: networking.SchemeGroupVersion.Version, Resource: "ingresses"}

	if ar.Request.Resource == ingressResource {
		ar.Response = &v1beta1.AdmissionResponse{
			UID:     types.UID(uuid.New().String()),
			Allowed: false,
		}
		ingress := networking.Ingress{}
		deserializer := codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, &ingress); err != nil {
			ar.Response.Result = &v1.Status{Message: err.Error()}
			ar.Response.AuditAnnotations = map[string]string{
				parser.GetAnnotationWithPrefix("error"): err.Error(),
			}
			klog.Errorf("failed to decode ingress %s in namespace %s: %s, refusing it", ar.Request.Name, ar.Request.Namespace, err.Error())
			return err
		}

		err := ia.Checker.CheckIngress(&ingress)
		if err != nil {
			ar.Response.Result = &v1.Status{Message: err.Error()}
			ar.Response.AuditAnnotations = map[string]string{
				parser.GetAnnotationWithPrefix("error"): err.Error(),
			}
			klog.Errorf("failed to generate configuration for ingress %s in namespace %s: %s, refusing it", ar.Request.Name, ar.Request.Namespace, err.Error())
			return err
		}
		ar.Response.Allowed = true
		klog.Infof("successfully validated configuration, accepting ingress %s in namespace %s", ar.Request.Name, ar.Request.Namespace)
		return nil
	}

	klog.Infof("accepting non ingress %s in namespace %s %s", ar.Request.Name, ar.Request.Namespace, ar.Request.Resource.String())
	ar.Response = &v1beta1.AdmissionResponse{
		UID:     types.UID(uuid.New().String()),
		Allowed: true,
	}
	return nil
}
