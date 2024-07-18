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

/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

// This file is an extraction from https://github.com/nginxinc/nginx-go-crossplane/blob/main/analyze.go
package extramodules

// bit masks for different directive argument styles.
const (
	ngxConfNoArgs = 0x00000001 // 0 args
	ngxConfTake1  = 0x00000002 // 1 args
	ngxConfTake2  = 0x00000004 // 2 args
	ngxConfTake3  = 0x00000008 // 3 args
	ngxConfTake4  = 0x00000010 // 4 args
	ngxConfTake5  = 0x00000020 // 5 args
	ngxConfTake6  = 0x00000040 // 6 args
	// ngxConfTake7  = 0x00000080 // 7 args (currently unused).
	ngxConfBlock = 0x00000100 // followed by block
	ngxConfExpr  = 0x00000200 // directive followed by expression in parentheses `()`
	ngxConfFlag  = 0x00000400 // 'on' or 'off'
	ngxConfAny   = 0x00000800 // >=0 args
	ngxConf1More = 0x00001000 // >=1 args
	ngxConf2More = 0x00002000 // >=2 args

	// some helpful argument style aliases.
	ngxConfTake12   = ngxConfTake1 | ngxConfTake2
	ngxConfTake13   = ngxConfTake1 | ngxConfTake3
	ngxConfTake23   = ngxConfTake2 | ngxConfTake3
	ngxConfTake34   = ngxConfTake3 | ngxConfTake4
	ngxConfTake123  = ngxConfTake12 | ngxConfTake3
	ngxConfTake1234 = ngxConfTake123 | ngxConfTake4

	// bit masks for different directive locations.
	ngxDirectConf     = 0x00010000 // main file (not used)
	ngxMgmtMainConf   = 0x00020000 // mgmt // unique bitmask that may not match NGINX source
	ngxMainConf       = 0x00040000 // main context
	ngxEventConf      = 0x00080000 // events
	ngxMailMainConf   = 0x00100000 // mail
	ngxMailSrvConf    = 0x00200000 // mail > server
	ngxStreamMainConf = 0x00400000 // stream
	ngxStreamSrvConf  = 0x00800000 // stream > server
	ngxStreamUpsConf  = 0x01000000 // stream > upstream
	ngxHTTPMainConf   = 0x02000000 // http
	ngxHTTPSrvConf    = 0x04000000 // http > server
	ngxHTTPLocConf    = 0x08000000 // http > location
	ngxHTTPUpsConf    = 0x10000000 // http > upstream
	ngxHTTPSifConf    = 0x20000000 // http > server > if
	ngxHTTPLifConf    = 0x40000000 // http > location > if
	ngxHTTPLmtConf    = 0x80000000 // http > location > limit_except
)

// helpful directive location alias describing "any" context
// doesn't include ngxHTTPSifConf, ngxHTTPLifConf, ngxHTTPLmtConf, or ngxMgmtMainConf.
//
//nolint:unused // This file is generated
const ngxAnyConf = ngxMainConf | ngxEventConf | ngxMailMainConf | ngxMailSrvConf |
	ngxStreamMainConf | ngxStreamSrvConf | ngxStreamUpsConf |
	ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPUpsConf |
	ngxHTTPSifConf | ngxHTTPLifConf | ngxHTTPLmtConf
