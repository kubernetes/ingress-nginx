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
	"fmt"

	networking "k8s.io/api/networking/v1"
)

// InspectIngress is used to do the deep inspection of an ingress object, walking through all
// of the spec fields and checking for matching strings and configurations that may represent
// an attempt to escape configs
func InspectIngress(ingress *networking.Ingress) error {
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			if err := CheckRegex(rule.Host); err != nil {
				return fmt.Errorf("invalid host in ingress %s/%s: %s", ingress.Namespace, ingress.Name, err)
			}
		}
		if rule.HTTP != nil {
			if err := inspectIngressRule(rule.HTTP); err != nil {
				return fmt.Errorf("invalid rule in ingress %s/%s: %s", ingress.Namespace, ingress.Name, err)
			}
		}
	}

	for _, tls := range ingress.Spec.TLS {
		if err := CheckRegex(tls.SecretName); err != nil {
			return fmt.Errorf("invalid secret in ingress %s/%s: %s", ingress.Namespace, ingress.Name, err)
		}
		for _, host := range tls.Hosts {
			if err := CheckRegex(host); err != nil {
				return fmt.Errorf("invalid host in ingress tls config %s/%s: %s", ingress.Namespace, ingress.Name, err)
			}
		}
	}
	return nil
}

func inspectIngressRule(httprule *networking.HTTPIngressRuleValue) error {
	for _, path := range httprule.Paths {
		if err := CheckRegex(path.Path); err != nil {
			return fmt.Errorf("invalid http path: %s", err)
		}
	}
	return nil
}
