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

package fastcgi

import (
	"fmt"
	"reflect"
	"regexp"

	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	fastCGIIndexAnnotation  = "fastcgi-index"
	fastCGIParamsAnnotation = "fastcgi-params-configmap" //#nosec G101
)

// fast-cgi valid parameters is just a single file name (like index.php)
var (
	regexValidIndexAnnotationAndKey = regexp.MustCompile(`^[A-Za-z0-9.\-\_]+$`)
	validFCGIValue                  = regexp.MustCompile(`^[A-Za-z0-9\-\_\$\{\}/.]*$`)
)

var fastCGIAnnotations = parser.Annotation{
	Group: "fastcgi",
	Annotations: parser.AnnotationFields{
		fastCGIIndexAnnotation: {
			Validator:     parser.ValidateRegex(regexValidIndexAnnotationAndKey, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium,
			Documentation: `This annotation can be used to specify an index file`,
		},
		fastCGIParamsAnnotation: {
			Validator: parser.ValidateRegex(parser.BasicCharsRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskMedium,
			Documentation: `This annotation can be used to specify a ConfigMap containing the fastcgi parameters as a key/value.
			Only ConfigMaps on the same namespace of ingress can be used. They key and value from ConfigMap are validated for unauthorized characters.`,
		},
	},
}

type fastcgi struct {
	r                resolver.Resolver
	annotationConfig parser.Annotation
}

// Config describes the per location fastcgi config
type Config struct {
	Index  string            `json:"index"`
	Params map[string]string `json:"params"`
}

// Equal tests for equality between two Configuration types
func (l1 *Config) Equal(l2 *Config) bool {
	if l1 == l2 {
		return true
	}

	if l1 == nil || l2 == nil {
		return false
	}

	if l1.Index != l2.Index {
		return false
	}

	return reflect.DeepEqual(l1.Params, l2.Params)
}

// NewParser creates a new fastcgiConfig protocol annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return fastcgi{
		r:                r,
		annotationConfig: fastCGIAnnotations,
	}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate the fastcgiConfig.
func (a fastcgi) Parse(ing *networking.Ingress) (interface{}, error) {
	fcgiConfig := Config{}

	if ing.GetAnnotations() == nil {
		return fcgiConfig, nil
	}

	index, err := parser.GetStringAnnotation(fastCGIIndexAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return fcgiConfig, err
		}
		index = ""
	}

	fcgiConfig.Index = index

	cm, err := parser.GetStringAnnotation(fastCGIParamsAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return fcgiConfig, err
		}
		return fcgiConfig, nil
	}

	cmns, cmn, err := cache.SplitMetaNamespaceKey(cm)
	if err != nil {
		return fcgiConfig, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("error reading configmap name from annotation: %w", err),
		}
	}
	secCfg := a.r.GetSecurityConfiguration()

	// We don't accept different namespaces for secrets.
	if cmns != "" && !secCfg.AllowCrossNamespaceResources && cmns != ing.Namespace {
		return fcgiConfig, fmt.Errorf("different namespace is not supported on fast_cgi param configmap")
	}

	cm = fmt.Sprintf("%v/%v", ing.Namespace, cmn)
	cmap, err := a.r.GetConfigMap(cm)
	if err != nil {
		return fcgiConfig, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("unexpected error reading configmap %s: %w", cm, err),
		}
	}

	for k, v := range cmap.Data {
		if !regexValidIndexAnnotationAndKey.MatchString(k) || !validFCGIValue.MatchString(v) {
			klog.ErrorS(fmt.Errorf("fcgi contains invalid key or value"), "fcgi annotation error", "configmap", cmap.Name, "namespace", cmap.Namespace, "key", k, "value", v)
			return fcgiConfig, ing_errors.NewValidationError(fastCGIParamsAnnotation)
		}
	}

	fcgiConfig.Index = index
	fcgiConfig.Params = cmap.Data

	return fcgiConfig, nil
}

func (a fastcgi) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a fastcgi) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, fastCGIAnnotations.Annotations)
}
