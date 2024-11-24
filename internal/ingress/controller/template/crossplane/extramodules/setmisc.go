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

// As opposite to the other files, this wasn't auto generated but hand crafted.
// Please do not change it

package extramodules

var setMiscDirectives = map[string][]uint{
	"set_escape_uri": {
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPSifConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConfTake12,
	},
}

func SetMiscMatchFn(directive string) ([]uint, bool) {
	m, ok := setMiscDirectives[directive]
	return m, ok
}
