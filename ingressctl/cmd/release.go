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
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"text/template"

	"github.com/codeskyblue/go-sh"
	"github.com/google/go-github/v48/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"strings"
	"time"
)

var INGRESS_ORG = "strongjz"                                                // the owner so we can test from forks
var INGRESS_REPO = "ingress-nginx"                                          // the repo to pull from
var RELEASE_BRANCH = "main"                                                 //we only release from main
var GITHUB_TOKEN string                                                     // the Google/gogithub lib needs an PAT to access the GitHub API
var K8S_IO_REPO = "k8s.io"                                                  //the repo that holds the images yaml for production promotion
var K8S_IO_BRANCH = "testing"                                               //branch to pull the k8s.io promoter yaml for controller tags
var INGRESS_REGISTRY = "registry.k8s.io"                                    //Container registry for storage Ingress-nginx images
var KUSTOMIZE_INSTALL_VERSION = "sigs.k8s.io/kustomize/kustomize/v4@v4.5.4" //static deploys needs kustomize to generate the template

// ingress-nginx releases start with a TAG then a cloudbuild, then a promotion through a PR, this the location of that PR
var IMAGES_YAML = "https://raw.githubusercontent.com/" + INGRESS_ORG + "/" + K8S_IO_REPO + "/testing/registry.k8s.io/images/k8s-staging-ingress-nginx/images.yaml"
var ctx = context.Background() // Context used for GitHub Client

// version - the version of the ingress-nginx controller to release
var version string
var path = "../"

// Documents that get updated for a controller release
var INDEX_DOCS = path + "docs/deploy/index.md"        //index.md has a version of the controller and needs to updated
var CHANGELOG = path + "Changelog.md"                 //Name of the changelog
var CHANGELOG_PATH = path + "changelog"               //folder for saving the new changelogs
var CHANGELOG_TEMPLATE = path + "Changelog.md.gotmpl" //path to go template for controller change log

// Documents that get updated for the ingress-nginx helm chart release
var CHART_PATH = path + "charts/ingress-nginx/"                   //path to the ingress-nginx helm chart
var CHART_CHANGELOG_PATH = CHART_PATH + "changelog"               //folder path to the helm chart changelog
var CHART_CHANGELOG_TEMPLATE = CHART_PATH + "changelog.md.gotmpl" //go template for the ingress-nginx helm chart

//Scripts listing

// releaseCmd - release root command for controller and helm charts
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Start a release",
	Long:  "Start a new release for ingress-nginx",
}

// helmReleaseCmd - release subcommand to release a new version of the ingress-nginx helm chart
var helmReleaseCmd = &cobra.Command{
	Use:   "helm",
	Short: "Start a new helm chart release",
	Long:  "Start a new helm chart release",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// controllerReleaseCmd - release subcommand to update all the files for a controller release
var controllerReleaseCmd = &cobra.Command{
	Use:   "controller",
	Short: "Release Ingress-nginx Controller",
	Long:  "Release a new version of ingress-nginx controller",
	Run: func(cmd *cobra.Command, args []string) {
		ControllerNewRelease(version)
	},
}

func init() {

	GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().StringVar(&path, "path", "../", "path to root ingress-nginx repo")
	releaseCmd.AddCommand(helmReleaseCmd)
	releaseCmd.AddCommand(controllerReleaseCmd)
	controllerReleaseCmd.Flags().StringVar(&version, "version", "v1.0.0-dev", "version of the controller to update")
}

// ControllerImage - struct with info about controllers
type ControllerImage struct {
	Tag      string
	Digest   string
	Registry string
	Name     string
}

// IngressRelease All the information about an ingress-nginx release that gets updated
type IngressRelease struct {
	ControllerVersion string
	ControllerImage   ControllerImage
	ReleaseNote       ReleaseNote
	Release           *github.RepositoryRelease
}

// IMAGES_YAML returns this data structure
type ImageYamls []ImageElement

// ImageElement - a specific image and it's data structure the dmap is a list of shas and container versions
type ImageElement struct {
	Name string              `json:"name"`
	Dmap map[string][]string `json:"dmap"`
}

// PromoteImage Creates PR into the k8s.io repo for promotion of ingress from staging to production
func PromoteImage(version, sha string) {

	//TODO

}

// Release Create a new release of ingress nginx controller
func ControllerNewRelease(version string) {

	//update ingress-nginx version
	//This is the step that kicks all the release process
	//it is already done, so it kicks off the gcloud build of the controller images

	//the version to release and the current version in TAG should match
	tag, err := getIngressNGINXVersion()
	CheckIfError(err, "RELEASE Retrieving the current Ingress Nginx Version")

	Info("RELEASE Checking Current Version %s to New Version %s", tag[1:], version)
	//if the version were upgrading does not match the TAG file, lets update the TAG file
	if tag[1:] != version {
		Warning("RELEASE Ingress Nginx TAG %s and new version %s do not match", tag, version)
		BumpNginx(fmt.Sprintf("v%s", version))
	}

	//update git controller tag controller-v$version
	NewControllerTag(version)

	//make release notes
	releaseNotes, err := ControllerReleaseNotes(version)
	CheckIfError(err, "RELEASE Creating Release Notes for version %s", version)
	Info("RELEASE Release Notes %s completed", releaseNotes.Version)

	Debug("releaseNotes.ControllerImages[0].Name %s", releaseNotes.ControllerImages[0].Name)
	Debug("releaseNotes.ControllerImages[1].Name %s", releaseNotes.ControllerImages[1].Name)

	//Record the ingress-nginx controller digests
	if releaseNotes.ControllerImages[0].Name == "ingress-nginx/controller" {
		Debug("RELEASE Updating Chart Value %s with %s", "controller.image.digest", releaseNotes.ControllerImages[0].Digest)
		updateChartValue("controller.image.digest", releaseNotes.ControllerImages[0].Digest)
	}

	//Record the ingress-nginx controller chroot digest
	if releaseNotes.ControllerImages[1].Name == "ingress-nginx/controller-chroot" {
		Debug("RELEASE Updating Chart Value %s with %s", "controller.image.digestChroot", releaseNotes.ControllerImages[1].Digest)
		updateChartValue("controller.image.digestChroot", releaseNotes.ControllerImages[1].Digest)
	}

	//update the Helm Chart appVersion aka the controller tag
	updateChartValue("controller.image.tag", fmt.Sprintf("v%s", releaseNotes.Version))

	//update chart values.yaml new controller tag and image digest
	releaseNotes.PreviousHelmChartVersion = currentChartVersion()

	//update helm chart app version
	UpdateVersion(version)

	releaseNotes.NewHelmChartVersion = currentChartVersion()

	//update helm chart release notes
	updateChartReleaseNotes(releaseNotes.HelmUpdates)

	//Run helm docs update
	CheckIfError(HelmDocs(), "RELEASE Error Updating Helm Docs ")

	releaseNotes.helmTemplate()

	//update static manifest
	CheckIfError(updateStaticManifest(), "RELEASE Error Updating Static manifests")

	//update e2e docs
	updateE2EDocs()

	//update documentation with ingress-nginx version
	CheckIfError(updateIndexMD(releaseNotes.PreviousControllerVersion, releaseNotes.NewControllerVersion), "Error Updating %s", INDEX_DOCS)

	//keeping these manual for now
	//git commit TODO
	//Create Pull Request TODO
}

// the index.md doc needs the controller version updated
func updateIndexMD(old, new string) error {
	Info("RELEASE Updating Deploy docs with new version")
	data, err := os.ReadFile(INDEX_DOCS)
	CheckIfError(err, "RELEASE Could not read INDEX_DOCS file %s", INDEX_DOCS)
	datString := string(data)
	datString = strings.Replace(datString, old, new, -1)
	err = os.WriteFile(INDEX_DOCS, []byte(datString), 644)
	if err != nil {
		ErrorF("RELEASE Could not write new %s %s", INDEX_DOCS, err)
		return err
	}
	return nil
}

// runs the hack/generate-deploy-scripts.sh
func updateE2EDocs() {
	updates, err := sh.Command(path + "hack/generate-e2e-suite-doc.sh").Output()
	CheckIfError(err, "Could not run update hack script")
	err = os.WriteFile(path+"docs/e2e-tests.md", []byte(updates), 644)
	CheckIfError(err, "Could not write new e2e test file ")
}

// The static deploy scripts use kustomize to generate them, this function ensures kustomize is installed
func installKustomize() error {
	Info("Install Kustomize")
	return sh.Command("go", "install", KUSTOMIZE_INSTALL_VERSION).Run()
}

func updateStaticManifest() error {
	CheckIfError(installKustomize(), "error installing kustomize")
	//hack/generate-deploy-scripts.sh
	return sh.Command(path + "/hack/generate-deploy-scripts.sh").Run()
}

/*
// CreateRelease Creates a new GitHub Release
func CreateRelease(name string) {
	releaser, err := gh_release.NewReleaser(INGRESS_ORG, INGRESS_REPO, GITHUB_TOKEN)
	CheckIfError(err, "GitHub Release Client error")
	newRelease, err := releaser.Create(fmt.Sprintf("controller-%s", name))
	CheckIfError(err, "Create release error")
	Info("New Release: Tag %v, ID: %v", newRelease.TagName, newRelease.ID)
}
*/

// Returns a GitHub client ready for use
func githubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_TOKEN},
	)
	oauthClient := oauth2.NewClient(ctx, ts)
	return github.NewClient(oauthClient)
}

// LatestCommitLogs Retrieves the commit log between the latest two controller versions.
func LatestCommitLogs() {
	commitLog := commitsBetweenTags()
	for i, s := range commitLog {
		Info("#%v Version %v", i, s)
	}
}

func commitsBetweenTags() []string {
	tags := getAllControllerTags()
	Info("Getting Commits between %v and %v", tags[0], tags[1])
	commitLog, err := sh.Command("git", "log", "--full-history", "--pretty", "--oneline", fmt.Sprintf("%v..%v", tags[1], tags[0])).Output()

	if len(commitLog) == 0 {
		Warning("All Controller Tags is empty")
	}
	CheckIfError(err, "Retrieving Commit log")
	return strings.Split(string(commitLog), "\n")
}

func (i ControllerImage) print() string {
	return fmt.Sprintf("%s/%s:%s@%s", i.Registry, i.Name, i.Tag, i.Digest)
}

func (r ReleaseNote) template() {
	// Files are provided as a slice of strings.
	changelogTemplate, err := os.ReadFile(CHANGELOG_TEMPLATE)
	if err != nil {
		ErrorF("Could not read changelog template file %s", err)
	}
	Debug("ChangeLog Templates %s", string(changelogTemplate))
	t := template.Must(template.New("changelog").Parse(string(changelogTemplate)))
	// create a new file
	file, err := os.Create(fmt.Sprintf("%s/Changelog-%s.md", CHANGELOG_PATH, r.Version))
	if err != nil {
		ErrorF("Could not create changelog file %s", err)
	}
	defer file.Close()

	err = t.Execute(file, r)
	if err != nil {
		ErrorF("executing template:", err)
	}
}

func (r ReleaseNote) helmTemplate() {
	// Files are provided as a slice of strings.
	changelogTemplate, err := os.ReadFile(CHART_CHANGELOG_TEMPLATE)
	if err != nil {
		ErrorF("Could not read changelog template file %s", err)
	}
	Debug("ChangeLog Templates %s", string(changelogTemplate))
	t := template.Must(template.New("changelog").Parse(string(changelogTemplate)))
	// create a new file
	fileName := fmt.Sprintf("%s/Changelog-%s.md", CHART_CHANGELOG_PATH, r.NewHelmChartVersion)
	file, err := os.Create(fileName)
	if err != nil {
		ErrorF("Could not create changelog file %s", err)
	}
	defer file.Close()

	err = t.Execute(file, r)
	if err != nil {
		ErrorF("executing template:", err)
	}
}

func (r ReleaseNote) printRelease() {
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

func findImageDigest(yaml ImageYamls, image, version string) string {
	version = fmt.Sprintf("v%s", version)
	Info("Searching Digest for %s:%s", image, version)
	for i := range yaml {
		if yaml[i].Name == image {
			for k, v := range yaml[i].Dmap {
				if v[0] == version {
					return k
				}
			}
			return ""
		}
	}
	return ""
}

func downloadFile(url string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Could not retrieve file, response from server %s for file %s", resp.StatusCode, url))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	return string(bodyBytes), nil
}

// Latest returns latest Github Release
func ReleaseLatest() error {
	r, _, err := latestRelease()
	if err != nil {
		ErrorF("Latest Release error %s", err)
		return err
	}
	Info("Latest Release %v", r.String())
	return nil
}

func ReleaseByTag(tag string) error {
	r, _, err := releaseByTag(tag)
	if err != nil {
		ErrorF("Release retrieve tag error %s", tag, err)
		return err
	}

	Info("Latest Release %v", r.String())

	return nil
}

func releaseByTag(tag string) (*github.RepositoryRelease, *github.Response, error) {
	ghClient := githubClient()
	return ghClient.Repositories.GetReleaseByTag(ctx, INGRESS_ORG, INGRESS_REPO, tag)
}

func latestRelease() (*github.RepositoryRelease, *github.Response, error) {
	ghClient := githubClient()
	return ghClient.Repositories.GetLatestRelease(ctx, INGRESS_ORG, INGRESS_REPO)
}

// Copy Test function to copy a release
func ReleaseCopy() error {
	ghClient := githubClient()
	kRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, INGRESS_ORG, INGRESS_REPO)
	if err != nil {
		ErrorF("Get Release from kubernetes %s", err)
		return err
	}

	sRelease := &github.RepositoryRelease{
		TagName:                kRelease.TagName,
		Name:                   kRelease.Name,
		Body:                   kRelease.Body,
		Draft:                  kRelease.Draft,
		Prerelease:             kRelease.GenerateReleaseNotes,
		DiscussionCategoryName: kRelease.DiscussionCategoryName,
		GenerateReleaseNotes:   kRelease.GenerateReleaseNotes,
	}

	sRelease, _, err = ghClient.Repositories.CreateRelease(ctx, INGRESS_ORG, INGRESS_REPO, sRelease)
	if err != nil {
		ErrorF("Creating %s/%s release %s", INGRESS_ORG, INGRESS_REPO, err)
		return err
	}
	Info("Copied over Kubernetes Release %v to Strongjz %v", &kRelease.Name, &sRelease.Name)
	return nil
}
