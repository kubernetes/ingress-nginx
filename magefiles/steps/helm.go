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

package steps

import (
	"bytes"
	"os"

	semver "github.com/blang/semver/v4"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	yamlpath "github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"

	utils "k8s.io/ingress-nginx/magefiles/utils"
)

const (
	HelmChartPath   = "charts/ingress-nginx/Chart.yaml"
	HelmChartValues = "charts/ingress-nginx/values.yaml"
)

type Helm mg.Namespace

// UpdateVersion Update Helm Version of the Chart
func (Helm) UpdateVersion(version string) {
	updateVersion(version)
}

func currentChartVersion() string {
	chart, err := chartutil.LoadChartfile(HelmChartPath)
	utils.CheckIfError(err, "HELM Could not Load Chart")
	return chart.Version
}

func updateVersion(version string) {
	utils.Info("HELM Reading File %v", HelmChartPath)

	chart, err := chartutil.LoadChartfile(HelmChartPath)
	utils.CheckIfError(err, "HELM Could not Load Chart")

	// Get the current tag
	// appVersionV, err := getIngressNGINXVersion()
	// utils.CheckIfError(err, "HELM Issue Retrieving the Current Ingress Nginx Version")

	// remove the v from TAG
	appVersion := version

	utils.Info("HELM Ingress-Nginx App Version: %s Chart AppVersion: %s", appVersion, chart.AppVersion)
	if appVersion == chart.AppVersion {
		utils.Warning("HELM Ingress NGINX Version didnt change Ingress-Nginx App Version: %s Chart AppVersion: %s", appVersion, chart.AppVersion)
		return
	}

	controllerSemVer, err := semver.Parse(version)
	utils.CheckIfError(err, "error parsing semver of new app")
	isPreRelease := len(controllerSemVer.Pre) > 0
	oldControllerSemVer, err := semver.Parse(chart.AppVersion)
	utils.CheckIfError(err, "error parsing semver of old chart")
	isBreakingChange := controllerSemVer.Major > oldControllerSemVer.Major || controllerSemVer.Minor > oldControllerSemVer.Minor

	// Update the helm chart
	chart.AppVersion = appVersion
	cTag, err := semver.Make(chart.Version)
	utils.CheckIfError(err, "HELM Creating Chart Version: %v", err)

	incrFunc := cTag.IncrementPatch
	if isBreakingChange {
		cTag.Patch = 0
		incrFunc = cTag.IncrementMinor

	}

	if isPreRelease {
		chart.Annotations["artifacthub.io/prerelease"] = "true"
		cTag.Pre = controllerSemVer.Pre
	}

	if err = incrFunc(); err != nil {
		utils.ErrorF("HELM Incrementing Chart Version: %v", err)
		os.Exit(1)
	}
	chart.Version = cTag.String()
	utils.Debug("HELM Updated Chart Version: %v", chart.Version)

	err = chartutil.SaveChartfile(HelmChartPath, chart)
	utils.CheckIfError(err, "HELM Saving new Chart")
}

func updateChartReleaseNotes(releaseNotes []string) {
	utils.Info("HELM Updating chart release notes")
	chart, err := chartutil.LoadChartfile(HelmChartPath)
	utils.CheckIfError(err, "HELM Failed to load chart manifest: %s", HelmChartPath)

	releaseNotesBytes, err := yaml.Marshal(releaseNotes)
	utils.CheckIfError(err, "HELM Failed to marshal release notes")

	releaseNotesString := string(releaseNotesBytes)
	utils.Info("HELM Chart release notes:\n%s", releaseNotesString)
	chart.Annotations["artifacthub.io/changes"] = releaseNotesString

	utils.Info("HELM Saving chart release notes")
	err = chartutil.SaveChartfile(HelmChartPath, chart)
	utils.CheckIfError(err, "HELM Failed to save chart manifest: %s", HelmChartPath)
}

// UpdateChartValue Updates the Helm ChartValue
func (Helm) UpdateChartValue(key, value string) {
	updateChartValue(key, value)
}

func updateChartValue(key, value string) {
	utils.Info("HELM Updating Chart %s %s:%s", HelmChartValues, key, value)

	// read current values.yaml
	data, err := os.ReadFile(HelmChartValues)
	utils.CheckIfError(err, "HELM Could not Load Helm Chart Values files %s", HelmChartValues)

	// var valuesStruct IngressChartValue
	var n yaml.Node
	utils.CheckIfError(yaml.Unmarshal(data, &n), "HELM Could not Unmarshal %s", HelmChartValues)

	// update value
	// keyParse := parsePath(key)
	p, err := yamlpath.NewPath(key)
	utils.CheckIfError(err, "HELM cannot create path")

	q, err := p.Find(&n)
	utils.CheckIfError(err, "HELM unexpected error finding path")

	for _, i := range q {
		utils.Info("HELM Found %s at %s", i.Value, key)
		i.Value = value
		utils.Info("HELM Updated %s at %s", i.Value, key)
	}

	//// write to file
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&n)
	utils.CheckIfError(err, "HELM Could not Marshal new Values file")
	err = os.WriteFile(HelmChartValues, b.Bytes(), 0o644)
	utils.CheckIfError(err, "HELM Could not write new Values file to %s", HelmChartValues)

	utils.Info("HELM Ingress Nginx Helm Chart update %s %s", key, value)
}

func (Helm) Helmdocs() error {
	return runHelmDocs()
}

func runHelmDocs() error {
	err := installHelmDocs()
	if err != nil {
		return err
	}
	err = sh.RunV("helm-docs", "--chart-search-root", "${PWD}/charts")
	if err != nil {
		return err
	}
	return nil
}

func installHelmDocs() error {
	utils.Info("HELM Install HelmDocs")
	g0 := sh.RunCmd("go")

	err := g0("install", "github.com/norwoodj/helm-docs/cmd/helm-docs@latest")
	if err != nil {
		return err
	}
	return nil
}
