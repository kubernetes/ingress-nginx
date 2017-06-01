/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"github.com/golang/glog"
	"os/exec"
	gotemplate "text/template"
)

type template struct {
	tmpl      *gotemplate.Template
	rawConfig *bytes.Buffer
	fmtConfig *bytes.Buffer
}

func newTemplate(name string, file string) *template {
	tmpl, err := gotemplate.New(name).ParseFiles(file)
	if err != nil {
		glog.Fatalf("Cannot read template file: %v", err)
	}
	return &template{
		tmpl:      tmpl,
		rawConfig: bytes.NewBuffer(make([]byte, 0, 16384)),
		fmtConfig: bytes.NewBuffer(make([]byte, 0, 16384)),
	}
}

func (t *template) execute(conf *configuration) ([]byte, error) {
	t.rawConfig.Reset()
	t.fmtConfig.Reset()
	if err := t.tmpl.Execute(t.rawConfig, conf); err != nil {
		return nil, err
	}
	cmd := exec.Command("sed", "/^ *$/d")
	cmd.Stdin = t.rawConfig
	cmd.Stdout = t.fmtConfig
	if err := cmd.Run(); err != nil {
		glog.Errorf("Template cleaning has failed: %v", err)
		// TODO recover and return raw buffer
		return nil, err
	}
	return t.fmtConfig.Bytes(), nil
}
