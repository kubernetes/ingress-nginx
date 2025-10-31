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
	"regexp"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/mitchellh/mapstructure"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/internal/ingress/annotations/authreq"
	"k8s.io/ingress-nginx/internal/ingress/annotations/customheaders"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
	"k8s.io/ingress-nginx/pkg/util/runtime"
)

const (
	customHTTPErrors              = "custom-http-errors"
	skipAccessLogUrls             = "skip-access-log-urls"
	whitelistSourceRange          = "whitelist-source-range"
	denylistSourceRange           = "denylist-source-range"
	proxyRealIPCIDR               = "proxy-real-ip-cidr"
	bindAddress                   = "bind-address"
	httpRedirectCode              = "http-redirect-code"
	blockCIDRs                    = "block-cidrs"
	blockUserAgents               = "block-user-agents"
	blockReferers                 = "block-referers"
	proxyStreamResponses          = "proxy-stream-responses"
	hideHeaders                   = "hide-headers"
	nginxStatusIpv4Whitelist      = "nginx-status-ipv4-whitelist"
	nginxStatusIpv6Whitelist      = "nginx-status-ipv6-whitelist"
	proxyHeaderTimeout            = "proxy-protocol-header-timeout"
	workerProcesses               = "worker-processes"
	globalAllowedResponseHeaders  = "global-allowed-response-headers"
	globalAuthURL                 = "global-auth-url"
	globalAuthMethod              = "global-auth-method"
	globalAuthSignin              = "global-auth-signin"
	globalAuthSigninRedirectParam = "global-auth-signin-redirect-param"
	globalAuthResponseHeaders     = "global-auth-response-headers"
	globalAuthRequestRedirect     = "global-auth-request-redirect"
	globalAuthSnippet             = "global-auth-snippet"
	globalAuthCacheKey            = "global-auth-cache-key"
	globalAuthCacheDuration       = "global-auth-cache-duration"
	globalAuthAlwaysSetCookie     = "global-auth-always-set-cookie"
	luaSharedDictsKey             = "lua-shared-dicts"
	debugConnections              = "debug-connections"
	workerSerialReloads           = "enable-serial-reloads"
)

var (
	validRedirectCodes    = sets.NewInt([]int{301, 302, 307, 308}...)
	dictSizeRegex         = regexp.MustCompile(`^(\d+)([kKmM])?$`)
	defaultLuaSharedDicts = map[string]int{
		"configuration_data":            20480,
		"certificate_data":              20480,
		"balancer_ewma":                 10240,
		"balancer_ewma_last_touched_at": 10240,
		"balancer_ewma_locks":           1024,
		"balancer_ewma_total":           10240,
		"balancer_ewma_failed":          10240,
		"certificate_servers":           5120,
		"ocsp_response_cache":           5120, // keep this same as certificate_servers
	}
	defaultGlobalAuthRedirectParam = "rd"
)

const (
	maxAllowedLuaDictSize = 204800
	maxNumberOfLuaDicts   = 100
)

// ReadConfig obtains the configuration defined by the user merged with the defaults.
//
//nolint:gocyclo // Ignore function complexity error
func ReadConfig(src map[string]string) config.Configuration {
	conf := map[string]string{}
	// we need to copy the configmap data because the content is altered
	for k, v := range src {
		conf[k] = v
	}

	to := config.NewDefault()
	errors := make([]int, 0)
	skipUrls := make([]string, 0)
	denyList := make([]string, 0)
	whiteList := make([]string, 0)
	proxyList := make([]string, 0)
	hideHeadersList := make([]string, 0)

	bindAddressIpv4List := make([]string, 0)
	bindAddressIpv6List := make([]string, 0)

	blockCIDRList := make([]string, 0)
	blockUserAgentList := make([]string, 0)
	blockRefererList := make([]string, 0)
	responseHeaders := make([]string, 0)
	allowedResponseHeaders := make([]string, 0)
	luaSharedDicts := make(map[string]int)
	debugConnectionsList := make([]string, 0)

	// parse lua shared dict values
	if val, ok := conf[luaSharedDictsKey]; ok {
		delete(conf, luaSharedDictsKey)
		lsd := splitAndTrimSpace(val, ",")
		for _, v := range lsd {
			results := strings.SplitN(strings.ReplaceAll(v, " ", ""), ":", 2)
			if len(results) != 2 {
				klog.Errorf("Ignoring poorly formatted Lua dictionary %v", v)
				continue
			}
			dictName := results[0]
			size := dictStrToKb(results[1])
			if size < 0 {
				klog.Errorf("Ignoring poorly formatted value %v for Lua dictionary %v", results[1], dictName)
				continue
			}
			if size > maxAllowedLuaDictSize {
				klog.Errorf("Ignoring %v for Lua dictionary %v: maximum size is %vk.", results[1], dictName, maxAllowedLuaDictSize)
				continue
			}
			if len(luaSharedDicts)+1 > maxNumberOfLuaDicts {
				klog.Errorf("Ignoring %v for Lua dictionary %v: can not configure more than %v dictionaries.",
					results[1], dictName, maxNumberOfLuaDicts)
				continue
			}

			luaSharedDicts[dictName] = size
		}
	}
	// set default Lua shared dicts
	for k, v := range defaultLuaSharedDicts {
		if _, ok := luaSharedDicts[k]; !ok {
			luaSharedDicts[k] = v
		}
	}

	if val, ok := conf[customHTTPErrors]; ok {
		delete(conf, customHTTPErrors)
		for _, i := range splitAndTrimSpace(val, ",") {
			j, err := strconv.Atoi(i)
			if err != nil {
				klog.Warningf("%v is not a valid http code: %v", i, err)
			} else {
				errors = append(errors, j)
			}
		}
	}

	if val, ok := conf[hideHeaders]; ok {
		delete(conf, hideHeaders)
		hideHeadersList = splitAndTrimSpace(val, ",")
	}

	if val, ok := conf[skipAccessLogUrls]; ok {
		delete(conf, skipAccessLogUrls)
		skipUrls = splitAndTrimSpace(val, ",")
	}

	if val, ok := conf[denylistSourceRange]; ok {
		delete(conf, denylistSourceRange)
		denyList = append(denyList, splitAndTrimSpace(val, ",")...)
	}

	if val, ok := conf[whitelistSourceRange]; ok {
		delete(conf, whitelistSourceRange)
		whiteList = append(whiteList, splitAndTrimSpace(val, ",")...)
	}

	if val, ok := conf[proxyRealIPCIDR]; ok {
		delete(conf, proxyRealIPCIDR)
		proxyList = append(proxyList, splitAndTrimSpace(val, ",")...)
	} else {
		proxyList = append(proxyList, "0.0.0.0/0")
	}

	if val, ok := conf[bindAddress]; ok {
		delete(conf, bindAddress)
		for _, i := range splitAndTrimSpace(val, ",") {
			ns := net.ParseIP(i)
			if ns != nil {
				if ing_net.IsIPV6(ns) {
					bindAddressIpv6List = append(bindAddressIpv6List, fmt.Sprintf("[%v]", ns))
				} else {
					bindAddressIpv4List = append(bindAddressIpv4List, ns.String())
				}
			} else {
				klog.Warningf("%v is not a valid textual representation of an IP address", i)
			}
		}
	}

	if val, ok := conf[blockCIDRs]; ok {
		delete(conf, blockCIDRs)
		blockCIDRList = splitAndTrimSpace(val, ",")
	}

	if val, ok := conf[blockUserAgents]; ok {
		delete(conf, blockUserAgents)
		blockUserAgentList = splitAndTrimSpace(val, ",")
	}

	if val, ok := conf[blockReferers]; ok {
		delete(conf, blockReferers)
		blockRefererList = splitAndTrimSpace(val, ",")
	}

	if val, ok := conf[httpRedirectCode]; ok {
		delete(conf, httpRedirectCode)
		j, err := strconv.Atoi(val)
		if err != nil {
			klog.Warningf("%v is not a valid HTTP code: %v", val, err)
		} else {
			if validRedirectCodes.Has(j) {
				to.HTTPRedirectCode = j
			} else {
				klog.Warningf("The code %v is not a valid as HTTP redirect code. Using the default.", val)
			}
		}
	}

	// Verify that the configured global external authorization URL is parsable as URL. if not, set the default value
	if val, ok := conf[globalAuthURL]; ok {
		delete(conf, globalAuthURL)

		authURL, message := parser.StringToURL(val)
		if authURL == nil {
			klog.Warningf("Global auth location denied - %v.", message)
		} else {
			to.GlobalExternalAuth.URL = val
			to.GlobalExternalAuth.Host = authURL.Hostname()
		}
	}

	// Verify that the configured global external authorization response headers are valid. if not, set the default value
	if val, ok := conf[globalAllowedResponseHeaders]; ok {
		delete(conf, globalAllowedResponseHeaders)

		if val != "" {
			harr := splitAndTrimSpace(val, ",")
			for _, header := range harr {
				if !customheaders.ValidHeader(header) {
					klog.Warningf("Global allowed response headers denied - %s.", header)
				} else {
					allowedResponseHeaders = append(allowedResponseHeaders, header)
				}
			}
		}
	}

	// Verify that the configured global external authorization method is a valid HTTP method. if not, set the default value
	if val, ok := conf[globalAuthMethod]; ok {
		delete(conf, globalAuthMethod)

		if val != "" && !authreq.ValidMethod(val) {
			klog.Warningf("Global auth location denied - %v.", "invalid HTTP method")
		} else {
			to.GlobalExternalAuth.Method = val
		}
	}

	// Verify that the configured global external authorization error page is set and valid. if not, set the default value
	if val, ok := conf[globalAuthSignin]; ok {
		delete(conf, globalAuthSignin)

		signinURL, err := parser.StringToURL(val)
		if err != nil {
			klog.Errorf("string to URL conversion failed: %v", err)
		}
		if signinURL == nil {
			klog.Warningf("Global auth location denied - %v.", "global-auth-signin setting is undefined and will not be set")
		} else {
			to.GlobalExternalAuth.SigninURL = val
		}
	}

	// Verify that the configured global external authorization error page redirection URL parameter is set and valid. if not, set the default value
	if val, ok := conf[globalAuthSigninRedirectParam]; ok {
		delete(conf, globalAuthSigninRedirectParam)

		redirectParam := strings.TrimSpace(val)
		dummySigninURL, err := parser.StringToURL(fmt.Sprintf("%s?%s=dummy", to.GlobalExternalAuth.SigninURL, redirectParam))
		if err != nil {
			klog.Errorf("string to URL conversion failed: %v", err)
		}
		if dummySigninURL == nil {
			klog.Warningf("Global auth redirect parameter denied - %v.", "global-auth-signin-redirect-param setting is invalid and will not be set")
		} else {
			to.GlobalExternalAuth.SigninURLRedirectParam = redirectParam
		}
	}

	// Verify that the configured global external authorization response headers are valid. if not, set the default value
	if val, ok := conf[globalAuthResponseHeaders]; ok {
		delete(conf, globalAuthResponseHeaders)

		if val != "" {
			harr := splitAndTrimSpace(val, ",")
			for _, header := range harr {
				if !authreq.ValidHeader(header) {
					klog.Warningf("Global auth location denied - %v.", "invalid headers list")
				} else {
					responseHeaders = append(responseHeaders, header)
				}
			}
		}
		to.GlobalExternalAuth.ResponseHeaders = responseHeaders
	}

	if val, ok := conf[globalAuthRequestRedirect]; ok {
		delete(conf, globalAuthRequestRedirect)
		to.GlobalExternalAuth.RequestRedirect = val
	}

	if val, ok := conf[globalAuthSnippet]; ok {
		delete(conf, globalAuthSnippet)
		to.GlobalExternalAuth.AuthSnippet = val
	}

	if val, ok := conf[globalAuthCacheKey]; ok {
		delete(conf, globalAuthCacheKey)
		to.GlobalExternalAuth.AuthCacheKey = val
	}

	// Verify that the configured global external authorization cache duration is valid
	if val, ok := conf[globalAuthCacheDuration]; ok {
		delete(conf, globalAuthCacheDuration)

		cacheDurations, err := authreq.ParseStringToCacheDurations(val)
		if err != nil {
			klog.Warningf("Global auth location denied - %s", err)
		}
		to.GlobalExternalAuth.AuthCacheDuration = cacheDurations
	}

	if val, ok := conf[globalAuthAlwaysSetCookie]; ok {
		delete(conf, globalAuthAlwaysSetCookie)

		alwaysSetCookie, err := strconv.ParseBool(val)
		if err != nil {
			klog.Warningf("Global auth location denied - %s", fmt.Errorf("cannot convert %s to bool: %v", globalAuthAlwaysSetCookie, err))
		}
		to.GlobalExternalAuth.AlwaysSetCookie = alwaysSetCookie
	}

	// Verify that the configured timeout is parsable as a duration. if not, set the default value
	if val, ok := conf[proxyHeaderTimeout]; ok {
		delete(conf, proxyHeaderTimeout)
		duration, err := time.ParseDuration(val)
		if err != nil {
			klog.Warningf("proxy-protocol-header-timeout of %v encountered an error while being parsed %v. Switching to use default value instead.", val, err)
		} else {
			to.ProxyProtocolHeaderTimeout = duration
		}
	}

	streamResponses := 1
	if val, ok := conf[proxyStreamResponses]; ok {
		delete(conf, proxyStreamResponses)
		j, err := strconv.Atoi(val)
		if err != nil {
			klog.Warningf("%v is not a valid number: %v", val, err)
		} else {
			streamResponses = j
		}
	}

	// Nginx Status whitelist
	if val, ok := conf[nginxStatusIpv4Whitelist]; ok {
		to.NginxStatusIpv4Whitelist = splitAndTrimSpace(val, ",")
		delete(conf, nginxStatusIpv4Whitelist)
	}

	if val, ok := conf[nginxStatusIpv6Whitelist]; ok {
		to.NginxStatusIpv6Whitelist = splitAndTrimSpace(val, ",")
		delete(conf, nginxStatusIpv6Whitelist)
	}

	if val, ok := conf[workerProcesses]; ok {
		to.WorkerProcesses = val
		if val == "auto" {
			to.WorkerProcesses = strconv.Itoa(runtime.NumCPU())
		}

		delete(conf, workerProcesses)
	}

	if val, ok := conf[workerSerialReloads]; ok {
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			to.WorkerSerialReloads = false
			klog.Warningf("failed to parse enable-serial-reloads setting, valid values are true or false, found %s", val)
		} else {
			to.WorkerSerialReloads = boolVal
		}
		delete(conf, workerSerialReloads)
	}

	if val, ok := conf[debugConnections]; ok {
		delete(conf, debugConnections)
		for _, i := range splitAndTrimSpace(val, ",") {
			validIP := net.ParseIP(i)
			if validIP != nil {
				debugConnectionsList = append(debugConnectionsList, i)
			} else {
				_, _, err := net.ParseCIDR(i)
				if err == nil {
					debugConnectionsList = append(debugConnectionsList, i)
				} else {
					klog.Warningf("%v is not a valid IP or CIDR address", i)
				}
			}
		}
		to.DebugConnections = debugConnectionsList
	}

	to.CustomHTTPErrors = filterErrors(errors)
	to.SkipAccessLogURLs = skipUrls
	to.DenylistSourceRange = denyList
	to.WhitelistSourceRange = whiteList
	to.ProxyRealIPCIDR = proxyList
	to.BindAddressIpv4 = bindAddressIpv4List
	to.BindAddressIpv6 = bindAddressIpv6List
	to.BlockCIDRs = blockCIDRList
	to.BlockUserAgents = blockUserAgentList
	to.BlockReferers = blockRefererList
	to.HideHeaders = hideHeadersList
	to.ProxyStreamResponses = streamResponses
	to.DisableIpv6DNS = !ing_net.IsIPv6Enabled()
	to.LuaSharedDicts = luaSharedDicts
	to.Backend.AllowedResponseHeaders = allowedResponseHeaders

	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           &to,
		TagName:          "json",
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		klog.Warningf("unexpected error merging defaults: %v", err)
	}
	err = decoder.Decode(conf)
	if err != nil {
		klog.Warningf("unexpected error merging defaults: %v", err)
	}

	hash, err := hashstructure.Hash(to, hashstructure.FormatV1, &hashstructure.HashOptions{
		TagName: "json",
	})
	if err != nil {
		klog.Warningf("unexpected error obtaining hash: %v", err)
	}

	to.Checksum = fmt.Sprintf("%v", hash)

	return to
}

func filterErrors(codes []int) []int {
	var fa []int
	for _, code := range codes {
		if code > 299 && code < 600 {
			fa = append(fa, code)
		} else {
			klog.Warningf("error code %v is not valid for custom error pages", code)
		}
	}

	return fa
}

//nolint:unparam // Ignore `sep` always receives `,` error
func splitAndTrimSpace(s, sep string) []string {
	f := func(c rune) bool {
		return strings.EqualFold(string(c), sep)
	}

	values := strings.FieldsFunc(s, f)
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return values
}

func dictStrToKb(sizeStr string) int {
	sizeMatch := dictSizeRegex.FindStringSubmatch(sizeStr)
	if sizeMatch == nil {
		return -1
	}
	size, err := strconv.Atoi(sizeMatch[1]) // validated already with regex
	if err != nil {
		klog.Errorf("unexpected error converting size string %s to int: %v", sizeStr, err)
	}
	if sizeMatch[2] == "" || strings.EqualFold(sizeMatch[2], "m") {
		size *= 1024
	}
	return size
}

func dictKbToStr(size int) string {
	if size%1024 == 0 {
		return fmt.Sprintf("%dM", size/1024)
	}
	return fmt.Sprintf("%dK", size)
}
