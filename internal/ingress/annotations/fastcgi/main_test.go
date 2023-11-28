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
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-backend",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
			},
		},
	}
}

type mockConfigMap struct {
	resolver.Mock
	extraConfigMap map[string]map[string]string
}

func (m mockConfigMap) GetConfigMap(name string) (*api.ConfigMap, error) {
	if m.extraConfigMap == nil {
		m.extraConfigMap = make(map[string]map[string]string)
	}
	cmdata, ok := m.extraConfigMap[name]

	if name != "default/demo-configmap" && name != "otherns/demo-configmap" && !ok {
		return nil, fmt.Errorf("there is no configmap with name %v", name)
	}

	cmns, cmn, err := cache.SplitMetaNamespaceKey(name)
	if err != nil {
		return nil, fmt.Errorf("invalid configmap name")
	}

	data := map[string]string{"REDIRECT_STATUS": "200", "SERVER_NAME": "$server_name"}
	if ok {
		data = cmdata
	}

	return &api.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: cmns,
			Name:      cmn,
		},
		Data: data,
	}, nil
}

func TestParseEmptyFastCGIAnnotations(t *testing.T) {
	ing := buildIngress()

	i, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress without fastcgi")
	}

	config, ok := i.(Config)
	if !ok {
		t.Errorf("Parse do not return a Config object")
	}

	if config.Index != "" {
		t.Errorf("Index should be an empty string")
	}

	if len(config.Params) != 0 {
		t.Errorf("Params should be an empty slice")
	}
}

func TestParseFastCGIIndexAnnotation(t *testing.T) {
	ing := buildIngress()

	const expectedAnnotation = "index.php"

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("fastcgi-index")] = expectedAnnotation
	ing.SetAnnotations(data)

	i, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress without fastcgi")
	}

	config, ok := i.(Config)
	if !ok {
		t.Errorf("Parse do not return a Config object")
	}

	if config.Index != "index.php" {
		t.Errorf("expected %s but %v returned", expectedAnnotation, config.Index)
	}
}

func TestParseEmptyFastCGIParamsConfigMapAnnotation(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = ""
	ing.SetAnnotations(data)

	i, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress without fastcgi")
	}

	config, ok := i.(Config)
	if !ok {
		t.Errorf("Parse do not return a Config object")
	}

	if len(config.Params) != 0 {
		t.Errorf("Params should be an empty slice")
	}
}

func TestParseFastCGIInvalidParamsConfigMapAnnotation(t *testing.T) {
	ing := buildIngress()

	invalidConfigMapList := []string{"unknown/configMap", "unknown/config/map"}
	for _, configmap := range invalidConfigMapList {
		data := map[string]string{}
		data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = configmap
		ing.SetAnnotations(data)

		i, err := NewParser(&mockConfigMap{}).Parse(ing)
		if err == nil {
			t.Errorf("Reading an unexisting configmap should return an error")
		}

		config, ok := i.(Config)
		if !ok {
			t.Errorf("Parse do not return a Config object")
		}

		if len(config.Params) != 0 {
			t.Errorf("Params should be an empty slice")
		}
	}
}

func TestParseFastCGIParamsConfigMapAnnotationWithoutNS(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = "demo-configmap"
	ing.SetAnnotations(data)

	i, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress without fastcgi")
	}

	config, ok := i.(Config)
	if !ok {
		t.Errorf("Parse do not return a Config object")
	}

	if len(config.Params) != 2 {
		t.Errorf("Params should have a length of 2")
	}

	if config.Params["REDIRECT_STATUS"] != "200" || config.Params["SERVER_NAME"] != "$server_name" {
		t.Errorf("Params value is not the one expected")
	}
}

func TestParseFastCGIParamsConfigMapAnnotationWithNS(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = "default/demo-configmap"
	ing.SetAnnotations(data)

	i, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress without fastcgi")
	}

	config, ok := i.(Config)
	if !ok {
		t.Errorf("Parse do not return a Config object")
	}

	if len(config.Params) != 2 {
		t.Errorf("Params should have a length of 2")
	}

	if config.Params["REDIRECT_STATUS"] != "200" || config.Params["SERVER_NAME"] != "$server_name" {
		t.Errorf("Params value is not the one expected")
	}
}

func TestParseFastCGIParamsConfigMapAnnotationWithDifferentNS(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = "otherns/demo-configmap"
	ing.SetAnnotations(data)

	_, err := NewParser(&mockConfigMap{}).Parse(ing)
	if err == nil {
		t.Errorf("Different namespace configmap should return an error")
	}
}

func TestConfigEquality(t *testing.T) {
	var nilConfig *Config

	config := Config{
		Index:  "index.php",
		Params: map[string]string{"REDIRECT_STATUS": "200", "SERVER_NAME": "$server_name"},
	}

	configCopy := Config{
		Index:  "index.php",
		Params: map[string]string{"REDIRECT_STATUS": "200", "SERVER_NAME": "$server_name"},
	}

	config2 := Config{
		Index:  "index.php",
		Params: map[string]string{"REDIRECT_STATUS": "200"},
	}

	config3 := Config{
		Index:  "index.py",
		Params: map[string]string{"SERVER_NAME": "$server_name", "REDIRECT_STATUS": "200"},
	}

	config4 := Config{
		Index:  "index.php",
		Params: map[string]string{"SERVER_NAME": "$server_name", "REDIRECT_STATUS": "200"},
	}

	if !config.Equal(&config) {
		t.Errorf("config should be equal to itself")
	}

	if nilConfig.Equal(&config) {
		t.Errorf("Foo")
	}

	if !config.Equal(&configCopy) {
		t.Errorf("config should be equal to configCopy")
	}

	if config.Equal(&config2) {
		t.Errorf("config2 should not be equal to config")
	}

	if config.Equal(&config3) {
		t.Errorf("config3 should not be equal to config")
	}

	if !config.Equal(&config4) {
		t.Errorf("config4 should be equal to config")
	}
}

func Test_fastcgi_Parse(t *testing.T) {
	tests := []struct {
		name          string
		index         string
		configmapname string
		configmap     map[string]string
		want          interface{}
		wantErr       bool
	}{
		{
			name:          "valid configuration",
			index:         "indexxpto-92123.php",
			configmapname: "default/fcgiconfig",
			configmap: map[string]string{
				"REQUEST_METHOD":  "$request_method",
				"SCRIPT_FILENAME": "$document_root$fastcgi_script_name",
			},
			want: Config{
				Index: "indexxpto-92123.php",
				Params: map[string]string{
					"REQUEST_METHOD":  "$request_method",
					"SCRIPT_FILENAME": "$document_root$fastcgi_script_name",
				},
			},
		},
		{
			name:          "invalid index name",
			index:         "indexxpto-92123$xx.php",
			configmapname: "default/fcgiconfig",
			configmap: map[string]string{
				"REQUEST_METHOD":  "$request_method",
				"SCRIPT_FILENAME": "$document_root$fastcgi_script_name",
			},
			want:    Config{},
			wantErr: true,
		},
		{
			name:          "invalid configmap namespace",
			index:         "indexxpto-92123.php",
			configmapname: "otherns/fcgiconfig",
			configmap: map[string]string{
				"REQUEST_METHOD":  "$request_method",
				"SCRIPT_FILENAME": "$document_root$fastcgi_script_name",
			},
			want:    Config{Index: "indexxpto-92123.php"},
			wantErr: true,
		},
		{
			name:          "invalid configmap namespace name",
			index:         "indexxpto-92123.php",
			configmapname: "otherns/fcgicon;{fig",
			configmap: map[string]string{
				"REQUEST_METHOD":  "$request_method",
				"SCRIPT_FILENAME": "$document_root$fastcgi_script_name",
			},
			want:    Config{Index: "indexxpto-92123.php"},
			wantErr: true,
		},
		{
			name:          "invalid configmap values key",
			index:         "indexxpto-92123.php",
			configmapname: "default/fcgiconfig",
			configmap: map[string]string{
				"REQUEST_METHOD$XPTO": "$request_method",
			},
			want:    Config{Index: "indexxpto-92123.php"},
			wantErr: true,
		},
		{
			name:          "invalid configmap values val",
			index:         "indexxpto-92123.php",
			configmapname: "default/fcgiconfig",
			configmap: map[string]string{
				"REQUEST_METHOD_XPTO": "$request_method{test};a",
			},
			want:    Config{Index: "indexxpto-92123.php"},
			wantErr: true,
		},
		{
			name:          "invalid configmap values val",
			index:         "indexxpto-92123.php",
			configmapname: "default/fcgiconfig",
			configmap: map[string]string{
				"SCRIPT_FILENAME": "/app/src/index.php",
			},
			want: Config{Index: "indexxpto-92123.php", Params: map[string]string{
				"SCRIPT_FILENAME": "/app/src/index.php",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ing := buildIngress()

			data := map[string]string{}
			data[parser.GetAnnotationWithPrefix("fastcgi-index")] = tt.index
			data[parser.GetAnnotationWithPrefix("fastcgi-params-configmap")] = tt.configmapname
			ing.SetAnnotations(data)

			m := &mockConfigMap{
				extraConfigMap: map[string]map[string]string{
					tt.configmapname: tt.configmap,
				},
			}

			got, err := NewParser(m).Parse(ing)
			if (err != nil) != tt.wantErr {
				t.Errorf("fastcgi.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fastcgi.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
