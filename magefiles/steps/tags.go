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
	"fmt"
	"os"
	"strings"

	semver "github.com/blang/semver/v4"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	utils "k8s.io/ingress-nginx/magefiles/utils"
)

type Tag mg.Namespace

var git = sh.OutCmd("git")

// Nginx returns the ingress-nginx current version
func (Tag) Nginx() {
	tag, err := getIngressNGINXVersion()
	utils.CheckIfError(err, "")
	fmt.Printf("%v", tag)
}

func getIngressNGINXVersion() (string, error) {
	dat, err := os.ReadFile("TAG")
	utils.CheckIfError(err, "Could not read TAG file")
	datString := string(dat)
	// remove newline
	datString = strings.Replace(datString, "\n", "", -1)
	return datString, nil
}

func checkSemVer(currentVersion, newVersion string) bool {
	utils.Info("Checking Sem Ver between current %s and new %s", currentVersion, newVersion)
	cVersion, err := semver.Make(currentVersion[1:])
	if err != nil {
		utils.ErrorF("TAG Error Current Tag %v Making Semver : %v", currentVersion[1:], err)
		return false
	}
	nVersion, err := semver.Make(newVersion)
	if err != nil {
		utils.ErrorF("TAG %v Error Making Semver %v \n", newVersion, err)
		return false
	}

	err = nVersion.Validate()
	if err != nil {
		utils.ErrorF("TAG %v not a valid Semver %v \n", newVersion, err)
		return false
	}

	//The result will be
	//0 if newVersion == currentVersion
	//-1 if newVersion < currentVersion
	//+1 if newVersion > currentVersion.
	utils.Info("TAG Comparing Old %s to New %s", cVersion.String(), nVersion.String())
	comp := nVersion.Compare(cVersion)
	if comp <= 0 {
		utils.Warning("SemVer:%v is not an update\n", newVersion)
		return false
	}
	return true
}

// BumpNginx will update the nginx TAG
func (Tag) BumpNginx(newTag string) {
	utils.Info("TAG BumpNginx version %v", newTag)
	currentTag, err := getIngressNGINXVersion()
	utils.CheckIfError(err, "Getting Ingress-nginx Version")
	bump(currentTag, newTag)
}

func bump(currentTag, newTag string) {
	// check if semver is valid
	if !checkSemVer(currentTag, newTag) {
		utils.ErrorF("ERROR: Semver is not valid %v", newTag)
		os.Exit(1)
	}

	utils.Info("Updating Tag %v to %v", currentTag, newTag)
	err := os.WriteFile("TAG", []byte(newTag), 0o666)
	utils.CheckIfError(err, "Error Writing New Tag File")
}

// Git Returns the latest git tag
func (Tag) Git() {
	tag, err := getGitTag()
	utils.CheckIfError(err, "Retrieving Git Tag")
	utils.Info("Git tag: %v", tag)
}

func getGitTag() (string, error) {
	return git("describe", "--tags", "--match", "controller-v*", "--abbrev=0")
}

// ControllerTag Creates a new Git Tag for the ingress controller
func (Tag) NewControllerTag(version string) {
	utils.Info("Create Ingress Nginx Controller Tag v%s", version)
	tag, err := controllerTag(version)
	utils.CheckIfError(err, "Creating git tag")
	utils.Debug("Git Tag: %s", tag)
}

func controllerTag(version string) (string, error) {
	return git("tag", "-a", "-m", fmt.Sprintf("-m \"Automated Controller release %v\"", version), fmt.Sprintf("controller-v%s", version))
}

func (Tag) AllControllerTags() {
	tags := getAllControllerTags()
	for i, s := range tags {
		utils.Info("#%v Version %v", i, s)
	}
}

func getAllControllerTags() []string {
	allControllerTags, err := git("tag", "-l", "--sort=-v:refname", "controller-v*")
	utils.CheckIfError(err, "Retrieving git tags")
	if !sh.CmdRan(err) {
		utils.Warning("Issue Running Command")
	}
	if allControllerTags == "" {
		utils.Warning("All Controller Tags is empty")
	}
	utils.Debug("Controller Tags: %v", allControllerTags)

	temp := strings.Split(allControllerTags, "\n")
	utils.Debug("There are %v controller tags", len(temp))
	return temp
}
