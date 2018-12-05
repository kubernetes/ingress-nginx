/*
Copyright 2016 The Kubernetes Authors.

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

package luarestywaf

import (
	"reflect"
	"strings"

	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var luaRestyWAFModes = map[string]bool{"ACTIVE": true, "INACTIVE": true, "SIMULATE": true}

// Config returns lua-resty-waf configuration for an Ingress rule
type Config struct {
	Mode                     string   `json:"mode"`
	Debug                    bool     `json:"debug"`
	IgnoredRuleSets          []string `json:"ignored-rulesets"`
	ExtraRulesetString       string   `json:"extra-ruleset-string"`
	ScoreThreshold           int      `json:"score-threshold"`
	AllowUnknownContentTypes bool     `json:"allow-unknown-content-types"`
	ProcessMultipartBody     bool     `json:"process-multipart-body"`
}

// Equal tests for equality between two Config types
func (e1 *Config) Equal(e2 *Config) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.Mode != e2.Mode {
		return false
	}
	if e1.Debug != e2.Debug {
		return false
	}
	if !reflect.DeepEqual(e1.IgnoredRuleSets, e2.IgnoredRuleSets) {
		return false
	}
	if e1.ExtraRulesetString != e2.ExtraRulesetString {
		return false
	}
	if e1.ScoreThreshold != e2.ScoreThreshold {
		return false
	}
	if e1.AllowUnknownContentTypes != e2.AllowUnknownContentTypes {
		return false
	}
	if e1.ProcessMultipartBody != e2.ProcessMultipartBody {
		return false
	}

	return true
}

type luarestywaf struct {
	r resolver.Resolver
}

// NewParser creates a new CORS annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return luarestywaf{r}
}

// Parse parses the annotations contained in the ingress rule
// used to indicate if the location/s contains a fragment of
// configuration to be included inside the paths of the rules
func (a luarestywaf) Parse(ing *extensions.Ingress) (interface{}, error) {
	var err error
	config := &Config{}

	mode, err := parser.GetStringAnnotation("lua-resty-waf", ing)
	if err != nil {
		return &Config{}, err
	}

	config.Mode = strings.ToUpper(mode)
	if _, ok := luaRestyWAFModes[config.Mode]; !ok {
		return &Config{}, errors.NewInvalidAnnotationContent("lua-resty-waf", mode)
	}

	config.Debug, _ = parser.GetBoolAnnotation("lua-resty-waf-debug", ing)

	ignoredRuleSetsStr, _ := parser.GetStringAnnotation("lua-resty-waf-ignore-rulesets", ing)
	config.IgnoredRuleSets = strings.FieldsFunc(ignoredRuleSetsStr, func(c rune) bool {
		strC := string(c)
		return strC == "," || strC == " "
	})

	// TODO(elvinefendi) maybe validate the ruleset string here
	config.ExtraRulesetString, _ = parser.GetStringAnnotation("lua-resty-waf-extra-rules", ing)

	config.ScoreThreshold, _ = parser.GetIntAnnotation("lua-resty-waf-score-threshold", ing)

	config.AllowUnknownContentTypes, _ = parser.GetBoolAnnotation("lua-resty-waf-allow-unknown-content-types", ing)

	config.ProcessMultipartBody, err = parser.GetBoolAnnotation("lua-resty-waf-process-multipart-body", ing)
	if err != nil {
		config.ProcessMultipartBody = true
	}

	return config, nil
}
