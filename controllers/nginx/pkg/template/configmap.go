/*
Copyright 2015 The Kubernetes Authors.

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

package template

import (
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/imdario/mergo"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	go_camelcase "github.com/segmentio/go-camelcase"

	"k8s.io/ingress/controllers/nginx/pkg/config"
	"k8s.io/ingress/core/pkg/ingress/defaults"

	"k8s.io/kubernetes/pkg/api"
)

const (
	customHTTPErrors     = "custom-http-errors"
	skipAccessLogUrls    = "skip-access-log-urls"
	whitelistSourceRange = "whitelist-source-range"
)

// StandarizeKeyNames ...
func StandarizeKeyNames(data interface{}) map[string]interface{} {
	return fixKeyNames(structs.Map(data))
}

// ReadConfig obtains the configuration defined by the user merged with the defaults.
func ReadConfig(conf *api.ConfigMap) config.Configuration {
	if len(conf.Data) == 0 {
		return config.NewDefault()
	}

	var errors []int
	var skipUrls []string
	var whitelist []string

	if val, ok := conf.Data[customHTTPErrors]; ok {
		delete(conf.Data, customHTTPErrors)
		for _, i := range strings.Split(val, ",") {
			j, err := strconv.Atoi(i)
			if err != nil {
				glog.Warningf("%v is not a valid http code: %v", i, err)
			} else {
				errors = append(errors, j)
			}
		}
	}
	if val, ok := conf.Data[skipAccessLogUrls]; ok {
		delete(conf.Data, skipAccessLogUrls)
		skipUrls = strings.Split(val, ",")
	}
	if val, ok := conf.Data[whitelistSourceRange]; ok {
		delete(conf.Data, whitelistSourceRange)
		whitelist = append(whitelist, strings.Split(val, ",")...)
	}

	to := config.Configuration{}
	to.Backend = defaults.Backend{
		CustomHTTPErrors:     filterErrors(errors),
		SkipAccessLogURLs:    skipUrls,
		WhitelistSourceRange: whitelist,
	}
	def := config.NewDefault()
	if err := mergo.Merge(&to, def); err != nil {
		glog.Warningf("unexpected error merging defaults: %v", err)
	}

	metadata := &mapstructure.Metadata{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "structs",
		Result:           &to,
		WeaklyTypedInput: true,
		Metadata:         metadata,
	})

	err = decoder.Decode(conf.Data)
	if err != nil {
		glog.Infof("%v", err)
	}
	return to
}

func filterErrors(codes []int) []int {
	var fa []int
	for _, code := range codes {
		if code > 299 && code < 600 {
			fa = append(fa, code)
		} else {
			glog.Warningf("error code %v is not valid for custom error pages", code)
		}
	}

	return fa
}

func fixKeyNames(data map[string]interface{}) map[string]interface{} {
	fixed := make(map[string]interface{})
	for k, v := range data {
		fixed[go_camelcase.Camelcase(k)] = v
	}
	return fixed
}
