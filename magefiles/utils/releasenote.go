/*
Copyright 2023 The Kubernetes Authors.

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

package utils

import (
	"fmt"
	"os"
	"text/template"
)

// ReleaseNote - All the pieces of information/documents that get updated during a release
type ReleaseNote struct {
	Version                   string
	NewControllerVersion      string
	PreviousControllerVersion string
	ControllerImages          []ControllerImage
	DepUpdates                []string
	Updates                   []string
	HelmUpdates               []string
	NewHelmChartVersion       string
	PreviousHelmChartVersion  string
}

func (r ReleaseNote) Template() {
	// Files are provided as a slice of strings.
	changelogTemplate, err := os.ReadFile("Changelog.md.gotmpl")
	if err != nil {
		ErrorF("Could not read changelog template file %s", err)
	}
	Debug("ChangeLog Templates %s", string(changelogTemplate))
	t := template.Must(template.New("changelog").Parse(string(changelogTemplate)))
	// create a new file
	file, err := os.Create(fmt.Sprintf("changelog/Changelog-%s.md", r.Version))
	if err != nil {
		ErrorF("Could not create changelog file %s", err)
	}
	defer file.Close()

	err = t.Execute(file, r)
	if err != nil {
		ErrorF("executing template: %s", err)
	}
}

func (r ReleaseNote) HelmTemplate() {
	// Files are provided as a slice of strings.
	changelogTemplate, err := os.ReadFile("charts/ingress-nginx/changelog.md.gotmpl")
	if err != nil {
		ErrorF("Could not read changelog template file %s", err)
	}
	Debug("ChangeLog Templates %s", string(changelogTemplate))
	t := template.Must(template.New("changelog").Parse(string(changelogTemplate)))
	// create a new file
	file, err := os.Create(fmt.Sprintf("charts/ingress-nginx/changelog/Changelog-%s.md", r.NewHelmChartVersion))
	if err != nil {
		ErrorF("Could not create changelog file %s", err)
	}
	defer file.Close()

	err = t.Execute(file, r)
	if err != nil {
		ErrorF("executing template: %s", err)
	}
}

func (r ReleaseNote) PrintRelease() {
	Info("Release Version: %v", r.NewControllerVersion)
	Info("Previous Version: %v", r.PreviousControllerVersion)
	Info("Controller Image: %v", r.ControllerImages[0].print())
	Info("Controller Chroot Image: %v", r.ControllerImages[1].print())
	for i := range r.Updates {
		Info("Update #%v - %v", i, r.Updates[i])
	}
	for j := range r.DepUpdates {
		Info("Dependabot Update #%v - %v", j, r.DepUpdates[j])
	}
}
