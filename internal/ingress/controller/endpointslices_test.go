/*
Copyright 2018 The Kubernetes Authors.

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

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

func TestGetEndpointsFromSlices(t *testing.T) {
	tests := []struct {
		name   string
		svc    *corev1.Service
		port   *corev1.ServicePort
		proto  corev1.Protocol
		zone   string
		fn     func(string) ([]*discoveryv1.EndpointSlice, error)
		result []ingress.Endpoint
	}{
		{
			"no service should return 0 endpoint",
			nil,
			nil,
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return nil, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"no service port should return 0 endpoint",
			&corev1.Service{},
			nil,
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return nil, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service without endpoint should return 0 endpoint",
			&corev1.Service{},
			&corev1.ServicePort{Name: "default"},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName service with an invalid port should return 0 endpoint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
				},
			},
			&corev1.ServicePort{Name: "default"},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName service with localhost in name should return 0 endpoint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "localhost",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName service with 127.0.0.1 in name should return 0 endpoint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "127.0.0.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName with a valid port should return one endpoint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "www.google.com",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "www.google.com",
					Port:    "443",
				},
			},
		},
		{
			"a service type ServiceTypeExternalName with an trailing dot ExternalName value should return one endpoints",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "www.google.com.",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "www.google.com",
					Port:    "443",
				},
			},
		},
		{
			"a service type ServiceTypeExternalName with an invalid ExternalName value should no return endpoints",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "1#invalid.hostname",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoint when there is an error searching for endpoints",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return nil, fmt.Errorf("unexpected error")
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoint when the protocol does not match",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Name:     &[]string{""}[0],
							Port:     &[]int32{80}[0],
							Protocol: &[]corev1.Protocol{corev1.ProtocolUDP}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoint when there is no ready Addresses",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{false}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Name:     &[]string{""}[0],
							Port:     &[]int32{80}[0],
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoint when the name of the port name do not match any port in the endpointPort and TargetPort is string",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"another-name"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return one endpoint when the name of the port name do not match any port in the endpointPort and TargetPort is int",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"another-name"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
			},
		},
		{
			"should return one endpoint when the name of the port name match a port in the endpointPort",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"default"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
			},
		},
		{
			"should return two endpoints when the name of the port name match a port in the endpointPort",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{"1.1.1.1"},
								Conditions: discoveryv1.EndpointConditions{
									Ready: &[]bool{true}[0],
								},
							},
						},
						Ports: []discoveryv1.EndpointPort{
							{
								Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
								Port:     &[]int32{80}[0],
								Name:     &[]string{"default"}[0],
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{"2.2.2.2"},
								Conditions: discoveryv1.EndpointConditions{
									Ready: &[]bool{true}[0],
								},
							},
						},
						Ports: []discoveryv1.EndpointPort{
							{
								Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
								Port:     &[]int32{80}[0],
								Name:     &[]string{"default"}[0],
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
				{
					Address: "2.2.2.2",
					Port:    "80",
				},
			},
		},
		{
			"should return one endpoints when the name of the port name match a port in the endpointPort",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{"1.1.1.1"},
								Conditions: discoveryv1.EndpointConditions{
									Ready: &[]bool{true}[0],
								},
							},
						},
						Ports: []discoveryv1.EndpointPort{
							{
								Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
								Port:     &[]int32{80}[0],
								Name:     &[]string{"port-1"}[0],
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{"2.2.2.2"},
								Conditions: discoveryv1.EndpointConditions{
									Ready: &[]bool{true}[0],
								},
							},
						},
						Ports: []discoveryv1.EndpointPort{
							{
								Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
								Port:     &[]int32{80}[0],
								Name:     &[]string{"another-name"}[0],
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
			},
		},
		{
			"should return one endpoint when the name of the port name match more than one port in the endpointPort",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"port-1"}[0],
						},
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"port-2"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
			},
		},
		{
			"should return one endpoint which belongs to zone",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"eu-west-1b",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1b",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1a",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.3"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1c",
								}},
							}}[0],
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"port-1"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
			},
		},
		{
			"should return all endpoints because one is missing zone hint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"eu-west-1b",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1b",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1b",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.3"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{}}[0],
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"port-1"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
				{
					Address: "1.1.1.2",
					Port:    "80",
				},
				{
					Address: "1.1.1.3",
					Port:    "80",
				},
			},
		},
		{
			"should return all endpoints because no zone from controller node",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			"",
			func(string) ([]*discoveryv1.EndpointSlice, error) {
				return []*discoveryv1.EndpointSlice{{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{discoveryv1.LabelServiceName: "default"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"1.1.1.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1a",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1b",
								}},
							}}[0],
						},
						{
							Addresses: []string{"1.1.1.3"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: &[]bool{true}[0],
							},
							Hints: &[]discoveryv1.EndpointHints{{
								ForZones: []discoveryv1.ForZone{{
									Name: "eu-west-1c",
								}},
							}}[0],
						},
					},
					Ports: []discoveryv1.EndpointPort{
						{
							Protocol: &[]corev1.Protocol{corev1.ProtocolTCP}[0],
							Port:     &[]int32{80}[0],
							Name:     &[]string{"port-1"}[0],
						},
					},
				}}, nil
			},
			[]ingress.Endpoint{
				{
					Address: "1.1.1.1",
					Port:    "80",
				},
				{
					Address: "1.1.1.2",
					Port:    "80",
				},
				{
					Address: "1.1.1.3",
					Port:    "80",
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := getEndpointsFromSlices(testCase.svc, testCase.port, testCase.proto, testCase.zone, testCase.fn)
			if len(testCase.result) != len(result) {
				t.Errorf("Expected %d Endpoints but got %d", len(testCase.result), len(result))
			}
		})
	}
}
