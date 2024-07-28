/*
Copyright 2024 The Kubernetes Authors.

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

package crossplane

import (
	"fmt"
	"net"
	"strconv"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ing_net "k8s.io/ingress-nginx/internal/net"
)

type seconds int

func buildDirective(directive string, args ...any) *ngx_crossplane.Directive {
	argsVal := make([]string, 0)
	for k := range args {
		switch v := args[k].(type) {
		case string:
			argsVal = append(argsVal, v)
		case []string:
			argsVal = append(argsVal, v...)
		case int:
			argsVal = append(argsVal, strconv.Itoa(v))
		case bool:
			argsVal = append(argsVal, boolToStr(v))
		case seconds:
			argsVal = append(argsVal, strconv.Itoa(int(v))+"s")
		}
	}
	return &ngx_crossplane.Directive{
		Directive: directive,
		Args:      argsVal,
	}
}

func buildLuaSharedDictionaries(cfg *config.Configuration) []*ngx_crossplane.Directive {
	out := make([]*ngx_crossplane.Directive, 0, len(cfg.LuaSharedDicts))
	for name, size := range cfg.LuaSharedDicts {
		sizeStr := dictKbToStr(size)
		out = append(out, buildDirective("lua_shared_dict", name, sizeStr))
	}

	return out
}

// TODO: The utils below should be moved to a level where they can be consumed by any template writer

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolversInternal(res []net.IP, disableIpv6 bool) []string {
	r := make([]string, 0)
	for _, ns := range res {
		if ing_net.IsIPV6(ns) {
			if disableIpv6 {
				continue
			}
			r = append(r, fmt.Sprintf("[%s]", ns))
		} else {
			r = append(r, ns.String())
		}
	}
	r = append(r, "valid=30s")

	if disableIpv6 {
		r = append(r, "ipv6=off")
	}

	return r
}

// buildBlockDirective is used to build a block directive
func buildBlockDirective(blockName string, args []string, block ngx_crossplane.Directives) *ngx_crossplane.Directive {
	return &ngx_crossplane.Directive{
		Directive: blockName,
		Args:      args,
		Block:     block,
	}
}

// buildMapDirective is used to build a map directive
func buildMapDirective(name, variable string, block ngx_crossplane.Directives) *ngx_crossplane.Directive {
	return buildBlockDirective("map", []string{name, variable}, block)
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func dictKbToStr(size int) string {
	if size%1024 == 0 {
		return fmt.Sprintf("%dM", size/1024)
	}
	return fmt.Sprintf("%dK", size)
}
