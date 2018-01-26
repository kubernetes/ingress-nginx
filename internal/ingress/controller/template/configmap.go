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
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"github.com/mitchellh/mapstructure"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
)

const (
	customHTTPErrors     = "custom-http-errors"
	skipAccessLogUrls    = "skip-access-log-urls"
	whitelistSourceRange = "whitelist-source-range"
	proxyRealIPCIDR      = "proxy-real-ip-cidr"
	bindAddress          = "bind-address"
	httpRedirectCode     = "http-redirect-code"
	proxyStreamResponses = "proxy-stream-responses"
	hideHeaders          = "hide-headers"
)

var (
	validRedirectCodes = sets.NewInt([]int{301, 302, 307, 308}...)
)

// ReadConfig obtains the configuration defined by the user merged with the defaults.
func ReadConfig(src map[string]string) config.Configuration {
	conf := map[string]string{}
	// we need to copy the configmap data because the content is altered
	for k, v := range src {
		conf[k] = v
	}

	errors := make([]int, 0)
	skipUrls := make([]string, 0)
	whitelist := make([]string, 0)
	proxylist := make([]string, 0)
	hideHeaderslist := make([]string, 0)

	bindAddressIpv4List := make([]string, 0)
	bindAddressIpv6List := make([]string, 0)
	redirectCode := 308

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
	if val, ok := conf[hideHeaders]; ok {
		delete(conf, hideHeaders)
		hideHeaderslist = strings.Split(val, ",")
	}
	if val, ok := conf[skipAccessLogUrls]; ok {
		delete(conf, skipAccessLogUrls)
		skipUrls = strings.Split(val, ",")
	}
	if val, ok := conf[whitelistSourceRange]; ok {
		delete(conf, whitelistSourceRange)
		whitelist = append(whitelist, strings.Split(val, ",")...)
	}
	if val, ok := conf[proxyRealIPCIDR]; ok {
		delete(conf, proxyRealIPCIDR)
		proxylist = append(proxylist, strings.Split(val, ",")...)
	} else {
		proxylist = append(proxylist, "0.0.0.0/0")
	}
	if val, ok := conf[bindAddress]; ok {
		delete(conf, bindAddress)
		for _, i := range strings.Split(val, ",") {
			ns := net.ParseIP(i)
			if ns != nil {
				if ing_net.IsIPV6(ns) {
					bindAddressIpv6List = append(bindAddressIpv6List, fmt.Sprintf("[%v]", ns))
				} else {
					bindAddressIpv4List = append(bindAddressIpv4List, fmt.Sprintf("%v", ns))
				}
			} else {
				glog.Warningf("%v is not a valid textual representation of an IP address", i)
			}
		}
	}

	if val, ok := conf[httpRedirectCode]; ok {
		delete(conf, httpRedirectCode)
		j, err := strconv.Atoi(val)
		if err != nil {
			glog.Warningf("%v is not a valid HTTP code: %v", val, err)
		} else {
			if validRedirectCodes.Has(j) {
				redirectCode = j
			} else {
				glog.Warningf("The code %v is not a valid as HTTP redirect code. Using the default.", val)
			}
		}
	}

	streamResponses := 1
	if val, ok := conf[proxyStreamResponses]; ok {
		delete(conf, proxyStreamResponses)
		j, err := strconv.Atoi(val)
		if err != nil {
			glog.Warningf("%v is not a valid number: %v", val, err)
		} else {
			streamResponses = j
		}
	}

	to := config.NewDefault()
	to.CustomHTTPErrors = filterErrors(errors)
	to.SkipAccessLogURLs = skipUrls
	to.WhitelistSourceRange = whitelist
	to.ProxyRealIPCIDR = proxylist
	to.BindAddressIpv4 = bindAddressIpv4List
	to.BindAddressIpv6 = bindAddressIpv6List
	to.HideHeaders = hideHeaderslist
	to.HTTPRedirectCode = redirectCode
	to.ProxyStreamResponses = streamResponses
	to.DisableIpv6DNS = !ing_net.IsIPv6Enabled()

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
