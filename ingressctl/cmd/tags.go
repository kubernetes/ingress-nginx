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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codeskyblue/go-sh"

	semver "github.com/blang/semver/v4"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "taggin information",
	Long:  "Retieves tagging information about the controller",
}

var ingressVersionCmd = &cobra.Command{
	Use:   "controller",
	Short: "Ingress-Nginx controller version",
	Run: func(cmd *cobra.Command, args []string) {
		vers, err := getIngressNGINXVersion()
		if err != nil {
			ErrorF("Could not determine ingress-nginx version: %v", err)
		}
		Info("Current Ingress-nginx version: %s", vers)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(ingressVersionCmd)

}
func getIngressNGINXVersion() (string, error) {
	dat, err := os.ReadFile("../TAG")
	CheckIfError(err, "Could not read TAG file")
	datString := string(dat)
	//remove newline
	datString = strings.Replace(datString, "\n", "", -1)
	return datString, nil
}

func checkSemVer(currentVersion, newVersion string) bool {
	Info("Checking Sem Ver between current %s and new %s", currentVersion, newVersion)
	cVersion, err := semver.Make(currentVersion[1:])
	if err != nil {
		ErrorF("TAG Error Current Tag %v Making Semver : %v", currentVersion[1:], err)
		return false
	}
	nVersion, err := semver.Make(newVersion)
	if err != nil {
		ErrorF("TAG %v Error Making Semver %v \n", newVersion, err)
		return false
	}

	err = nVersion.Validate()
	if err != nil {
		ErrorF("TAG %v not a valid Semver %v \n", newVersion, err)
		return false
	}

	//The result will be
	//0 if newVersion == currentVersion
	//-1 if newVersion < currentVersion
	//+1 if newVersion > currentVersion.
	Info("TAG Comparing Old %s to New %s", cVersion.String(), nVersion.String())
	comp := nVersion.Compare(cVersion)
	if comp <= 0 {
		Warning("SemVer:%v is not an update\n", newVersion)
		return false
	}
	return true
}

// BumpNginx will update the nginx TAG
func BumpNginx(newTag string) {
	Info("TAG BumpNginx version %v", newTag)
	currentTag, err := getIngressNGINXVersion()
	CheckIfError(err, "Getting Ingress-nginx Version")
	bump(currentTag, newTag)
}

func bump(currentTag, newTag string) {
	//check if semver is valid
	if !checkSemVer(currentTag, newTag) {
		ErrorF("ERROR: Semver is not valid %v", newTag)
		os.Exit(1)
	}

	Info("Updating Tag %v to %v", currentTag, newTag)
	err := os.WriteFile("../TAG", []byte(newTag), 0666)
	CheckIfError(err, "Error Writing New Tag File")
}

// Git Returns the latest git tag
func Git() {
	tag, err := getGitTag()
	CheckIfError(err, "Retrieving Git Tag")
	Info("Git tag: %v", tag)
}

func getGitTag() (string, error) {
	out, err := sh.Command("git", "describe", "--tags", "--match", "controller-v*", "--abbrev=1").Output()
	return string(out), err
}

// ControllerTag Creates a new Git Tag for the ingress controller
func NewControllerTag(version string) {
	Info("Create Ingress Nginx Controller Tag v%s", version)
	tag, err := controllerTag(version)
	CheckIfError(err, "Creating git tag")
	Debug("Git Tag: %s", tag)
}

func controllerTag(version string) (string, error) {
	out, err := sh.Command("git", "tag", "-a", "-m", fmt.Sprintf("-m \"Automated Controller release %v\"", version), fmt.Sprintf("controller-v%s", version)).Output()
	return string(out), err
}

func AllControllerTags() {
	tags := getAllControllerTags()
	for i, s := range tags {
		Info("#%v Version %v", i, s)
	}
}

func getAllControllerTags() []string {
	out, err := sh.Command("git", "tag", "-l", "--sort=-v:refname", "controller-v*").Output()
	CheckIfError(err, "Retrieving git tags")
	if err != nil {
		Warning("Issue Running Command")
	}
	if len(out) == 0 {
		Warning("All Controller Tags is empty")
	}
	Debug("Controller Tags: %v", out)

	temp := strings.Split(string(out), "\n")
	Debug("There are %v controller tags", len(temp))
	return temp
}
