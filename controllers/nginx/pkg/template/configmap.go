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
	"github.com/mitchellh/mapstructure"

	"k8s.io/ingress/controllers/nginx/pkg/config"
)

const (
	customHTTPErrors     = "custom-http-errors"
	skipAccessLogUrls    = "skip-access-log-urls"
	whitelistSourceRange = "whitelist-source-range"
)

// ReadConfig obtains the configuration defined by the user merged with the defaults.
func ReadConfig(src map[string]string) config.Configuration {
	conf := map[string]string{}
	if src != nil {
		// we need to copy the configmap data because the content is altered
		for k, v := range src {
			conf[k] = v
		}
	}

	errors := make([]int, 0)
	skipUrls := make([]string, 0)
	whitelist := make([]string, 0)

	if val, ok := conf[customHTTPErrors]; ok {
		delete(conf, customHTTPErrors)
		for _, i := range strings.Split(val, ",") {
			j, err := strconv.Atoi(i)
			if err != nil {
				glog.Warningf("%v is not a valid http code: %v", i, err)
			} else {
				errors = append(errors, j)
			}
		}
	}
	if val, ok := conf[skipAccessLogUrls]; ok {
		delete(conf, skipAccessLogUrls)
		skipUrls = strings.Split(val, ",")
	}
	if val, ok := conf[whitelistSourceRange]; ok {
		delete(conf, whitelistSourceRange)
		whitelist = append(whitelist, strings.Split(val, ",")...)
	}

	to := config.NewDefault()
	to.CustomHTTPErrors = filterErrors(errors)
	to.SkipAccessLogURLs = skipUrls
	to.WhitelistSourceRange = whitelist

	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           &to,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		glog.Warningf("unexpected error merging defaults: %v", err)
	}
	err = decoder.Decode(conf)
	if err != nil {
		glog.Warningf("unexpected error merging defaults: %v", err)
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
