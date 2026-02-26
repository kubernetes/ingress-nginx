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
	"strings"

	semver "github.com/blang/semver/v4"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.yaml.in/yaml/v3"
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

// Updates a Helm chart value by path and value.
func (Helm) UpdateChartValue(path, value string) {
	updateChartValue(path, value)
}

// Updates a Helm chart value by path and value.
func updateChartValue(path, value string) {
	utils.Info("HELM Updating path %q to value %q in file %q", path, value, HelmChartValues)

	// Read file.
	file, err := os.ReadFile(HelmChartValues)
	utils.CheckIfError(err, "HELM Failed to read file %q", HelmChartValues)

	// Unmarshal values.
	var values yaml.Node
	err = yaml.Unmarshal(file, &values)
	utils.CheckIfError(err, "HELM Failed to unmarshal values %q", HelmChartValues)

	// Variable to track if we updated the value in the values.
	updated := false

	// Iterate nodes.
	for _, node := range values.Content {
		// Variable to track if we found the path in the values.
		found := false

		// Split path into keys and iterate over them to find the node to update.
		for _, key := range strings.Split(path, ".") {
			// Reset found variable for each key, since it might happen a single key of a path is not a mapping or not found.
			found = false

			// Check if node is a mapping node.
			if node.Kind != yaml.MappingNode {
				break
			}

			// Iterate over mapping content to find the key.
			// Each key and its value are stored in consecutive positions in the content slice, so we need to iterate with a step of 2.
			for i := 0; i < len(node.Content); i += 2 {
				// Check if the current key matches the desired key.
				if node.Content[i].Value == key {
					// If we found the key, we need to update the mapping variable to point to its value.
					node = node.Content[i+1]
					found = true
					break
				}
			}

			// Check if we found the key in the mapping.
			if !found {
				break
			}
		}

		// Check if we found the path in the node.
		if !found {
			continue
		}

		// Update the value of the node.
		node.SetString(value)
		updated = true
	}

	// Check if we updated the value in the values.
	if !updated {
		utils.ErrorF("HELM Could not find path %q in values %q", path, HelmChartValues)
		os.Exit(1)
	}

	// Setup encoder with buffer and indent.
	var encodedValues bytes.Buffer
	encoder := yaml.NewEncoder(&encodedValues)
	encoder.SetIndent(2)

	// Encode the values.
	err = encoder.Encode(&values)
	utils.CheckIfError(err, "HELM Failed to encode values")

	// Write the values to file.
	//nolint:gosec // We need to write to the file with 644 permissions.
	err = os.WriteFile(HelmChartValues, encodedValues.Bytes(), 0o644)
	utils.CheckIfError(err, "HELM Failed to write values to file %q", HelmChartValues)

	utils.Info("HELM Updated path %q to value %q in file %q", path, value, HelmChartValues)
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
