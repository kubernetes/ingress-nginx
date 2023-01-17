//go:build mage

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

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"io"
	"net"
	"net/http"
	"os"
	"text/template"

	"regexp"
	"strings"
	"time"
)

type Release mg.Namespace

var INGRESS_ORG = "kubernetes"                                              // the owner so we can test from forks
var INGRESS_REPO = "ingress-nginx"                                          // the repo to pull from
var RELEASE_BRANCH = "main"                                                 //we only release from main
var GITHUB_TOKEN string                                                     // the Google/gogithub lib needs an PAT to access the GitHub API
var K8S_IO_ORG = "kubernetes"                                               //the owner or organization for the k8s.io repo
var K8S_IO_REPO = "k8s.io"                                                  //the repo that holds the images yaml for production promotion
var INGRESS_REGISTRY = "registry.k8s.io"                                    //Container registry for storage Ingress-nginx images
var KUSTOMIZE_INSTALL_VERSION = "sigs.k8s.io/kustomize/kustomize/v4@v4.5.4" //static deploys needs kustomize to generate the template

// ingress-nginx releases start with a TAG then a cloudbuild, then a promotion through a PR, this the location of that PR
var IMAGES_YAML = "https://raw.githubusercontent.com/kubernetes/k8s.io/main/k8s.gcr.io/images/k8s-staging-ingress-nginx/images.yaml"
var ctx = context.Background() // Context used for GitHub Client

const INDEX_DOCS = "docs/deploy/index.md" //index.md has a version of the controller and needs to updated
const CHANGELOG = "Changelog.md"          //Name of the changelog

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

// IMAGES_YAML returns this data structure
type ImageYamls []ImageElement

// ImageElement - a specific image and it's data structure the dmap is a list of shas and container versions
type ImageElement struct {
	Name string              `json:"name"`
	Dmap map[string][]string `json:"dmap"`
}

// init will set the GitHub token from the committers/releasers env var
func init() {
	GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
}

// PromoteImage Creates PR into the k8s.io repo for promotion of ingress from staging to production
func (Release) PromoteImage(version, sha string) {

}

// Release Create a new release of ingress nginx controller
func (Release) NewRelease(version string) {
	//newRelease := Release{}

	//update ingress-nginx version
	//This is the step that kicks all the release process
	//it is already done, so it kicks off the gcloud build of the controller images
	//mg.Deps(mg.F(Tag.BumpNginx, version))

	tag, err := getIngressNGINXVersion()
	CheckIfError(err, "RELEASE Retrieving the current Ingress Nginx Version")

	Info("RELEASE Checking Current Version %s to New Version %s", tag, version)
	//if the version were upgrading does not match the TAG file, lets update the TAG file
	if tag[1:] != version {
		Warning("RELEASE Ingress Nginx TAG %s and new version %s do not match", tag, version)
		mg.Deps(mg.F(Tag.BumpNginx, fmt.Sprintf("v%s", version)))
	}

	//update git controller tag controller-v$version
	mg.Deps(mg.F(Tag.NewControllerTag, version))

	//make release notes
	releaseNotes, err := makeReleaseNotes(version)
	CheckIfError(err, "RELEASE Creating Release Notes for version %s", version)
	Info("RELEASE Release Notes %s completed", releaseNotes.Version)

	//update chart values.yaml new controller tag and image digest
	releaseNotes.PreviousHelmChartVersion = currentChartVersion()

	//controller tag
	updateChartValue("controller.image.tag", fmt.Sprintf("v%s", releaseNotes.Version))
	//controller digest
	if releaseNotes.ControllerImages[0].Name == "controller" {
		updateChartValue("controller.image.digest", releaseNotes.ControllerImages[0].Digest)
	}
	//controller chroot digest
	if releaseNotes.ControllerImages[1].Name == "controller-chroot" {
		updateChartValue("controller.image.digestChroot", releaseNotes.ControllerImages[1].Digest)
	}

	//update helm chart app version
	mg.Deps(mg.F(Helm.UpdateVersion, version))

	releaseNotes.NewHelmChartVersion = currentChartVersion()

	//update helm chart release notes
	updateChartReleaseNotes(releaseNotes.HelmUpdates)

	//Run helm docs update
	CheckIfError(runHelmDocs(), "Error Updating Helm Docs ")

	releaseNotes.helmTemplate()

	//update static manifest
	CheckIfError(updateStaticManifest(), "Error Updating Static manifests")

	////update e2e docs
	updateE2EDocs()

	//update documentation with ingress-nginx version
	CheckIfError(updateIndexMD(releaseNotes.PreviousControllerVersion, releaseNotes.NewControllerVersion), "Error Updating %s", INDEX_DOCS)

	//keeping these manual for now
	//git commit TODO
	//make Pull Request TODO
	//make release TODO
	//mg.Deps(mg.F(Release.CreateRelease, version))
}

// the index.md doc needs the controller version updated
func updateIndexMD(old, new string) error {
	Info("Updating Deploy docs with new version")
	data, err := os.ReadFile(INDEX_DOCS)
	CheckIfError(err, "Could not read INDEX_DOCS file %s", INDEX_DOCS)
	datString := string(data)
	datString = strings.Replace(datString, old, new, -1)
	err = os.WriteFile(INDEX_DOCS, []byte(datString), 644)
	if err != nil {
		ErrorF("Could not write new %s %s", INDEX_DOCS, err)
		return err
	}
	return nil
}

// runs the hack/generate-deploy-scripts.sh
func updateE2EDocs() {
	updates, err := sh.Output("./hack/generate-e2e-suite-doc.sh")
	CheckIfError(err, "Could not run update hack script")
	err = os.WriteFile("docs/e2e-tests.md", []byte(updates), 644)
	CheckIfError(err, "Could not write new e2e test file ")
}

// The static deploy scripts use kustomize to generate them, this function ensures kustomize is installed
func installKustomize() error {
	Info("Install Kustomize")
	var g0 = sh.RunCmd("go")
	// somewhere in your main code
	err := g0("install", KUSTOMIZE_INSTALL_VERSION)
	if err != nil {
		return err
	}
	return nil
}

func updateStaticManifest() error {
	CheckIfError(installKustomize(), "error installing kustomize")
	//hack/generate-deploy-scripts.sh
	err := sh.RunV("./hack/generate-deploy-scripts.sh")
	if err != nil {
		return err
	}
	return nil
}

//// CreateRelease Creates a new GitHub Release
//func (Release) CreateRelease(name string) {
//	releaser, err := gh_release.NewReleaser(INGRESS_ORG, INGRESS_REPO, GITHUB_TOKEN)
//	CheckIfError(err, "GitHub Release Client error")
//	newRelease, err := releaser.Create(fmt.Sprintf("controller-%s", name))
//	CheckIfError(err, "Create release error")
//	Info("New Release: Tag %v, ID: %v", newRelease.TagName, newRelease.ID)
//}

// Returns a GitHub client ready for use
func githubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_TOKEN},
	)
	oauthClient := oauth2.NewClient(ctx, ts)
	return github.NewClient(oauthClient)
}

// LatestCommitLogs Retrieves the commit log between the latest two controller versions.
func (Release) LatestCommitLogs() {
	commitLog := commitsBetweenTags()
	for i, s := range commitLog {
		Info("#%v Version %v", i, s)
	}
}

func commitsBetweenTags() []string {
	tags := getAllControllerTags()
	Info("Getting Commits between %v and %v", tags[0], tags[1])
	commitLog, err := git("log", "--full-history", "--pretty", "--oneline", fmt.Sprintf("%v..%v", tags[1], tags[0]))

	if commitLog == "" {
		Warning("All Controller Tags is empty")
	}
	CheckIfError(err, "Retrieving Commit log")
	return strings.Split(commitLog, "\n")
}

// Generate Release Notes
func (Release) ReleaseNotes(newVersion string) error {
	notes, err := makeReleaseNotes(newVersion)
	CheckIfError(err, "Creating Release Notes for version %s", newVersion)
	Info("Release Notes %s completed", notes.Version)
	return nil
}

func makeReleaseNotes(newVersion string) (*ReleaseNote, error) {
	var newReleaseNotes = ReleaseNote{}

	newReleaseNotes.Version = newVersion
	allControllerTags := getAllControllerTags()

	//new version
	newReleaseNotes.NewControllerVersion = allControllerTags[0]
	newControllerVersion := fmt.Sprintf("controller-v%s", newVersion)

	//the newControllerVersion should match the latest tag
	if newControllerVersion != allControllerTags[0] {
		return nil, errors.New(fmt.Sprintf("Generating release new version %s didnt match the current latest tag %s", newControllerVersion, allControllerTags[0]))
	}
	//previous version
	newReleaseNotes.PreviousControllerVersion = allControllerTags[1]

	Info("New Version: %s Old Version: %s", newReleaseNotes.NewControllerVersion, newReleaseNotes.PreviousControllerVersion)

	commits := commitsBetweenTags()

	//dependency_updates
	//all_updates
	var allUpdates []string
	var depUpdates []string
	var helmUpdates []string
	prRegex := regexp.MustCompile("\\(#\\d+\\)")
	depBot := regexp.MustCompile("^(\\w){1,10} Bump ")
	helmRegex := regexp.MustCompile("helm|chart")
	for i, s := range commits {
		//matches on PR
		if prRegex.Match([]byte(s)) {
			//matches a dependant bot update
			if depBot.Match([]byte(s)) { //
				Debug("#%v DEPENDABOT %v", i, s)
				u := strings.SplitN(s, " ", 2)
				depUpdates = append(depUpdates, u[1])
			} else { // add it to the all updates slice
				Debug("#%v ALL UPDATES %v", i, s)
				u := strings.SplitN(s, " ", 2)
				allUpdates = append(allUpdates, u[1])

				//helm chart updates
				if helmRegex.Match([]byte(s)) {
					u := strings.SplitN(s, " ", 2)
					helmUpdates = append(helmUpdates, u[1])
				}
			}

		}
	}
	helmUpdates = append(helmUpdates, fmt.Sprintf("Update Ingress-Nginx version %s", newReleaseNotes.NewControllerVersion))

	newReleaseNotes.Updates = allUpdates
	newReleaseNotes.DepUpdates = depUpdates
	newReleaseNotes.HelmUpdates = helmUpdates

	//controller_image_digests
	imagesYaml, err := downloadFile(IMAGES_YAML)
	if err != nil {
		ErrorF("Could not download file %s : %s", IMAGES_YAML, err)
		return nil, err
	}
	Debug("%s", imagesYaml)

	data := ImageYamls{}

	err = yaml.Unmarshal([]byte(imagesYaml), &data)
	if err != nil {
		ErrorF("Could not unmarshal images yaml %s", err)
		return nil, err
	}

	//controller
	controllerDigest := findImageDigest(data, "controller", newVersion)
	if len(controllerDigest) == 0 {
		ErrorF("Controller Digest could not be found")
		return nil, errors.New("Controller digest could not be found")
	}

	controllerChrootDigest := findImageDigest(data, "controller-chroot", newVersion)
	if len(controllerChrootDigest) == 0 {
		ErrorF("Controller Chroot Digest could not be found")
		return nil, errors.New("Controller Chroot digest could not be found")
	}

	Debug("Latest Controller Digest %v", controllerDigest)
	Debug("Latest Controller Chroot Digest %v", controllerChrootDigest)
	c1 := ControllerImage{
		Digest:   controllerDigest,
		Registry: INGRESS_REGISTRY,
		Name:     "controller",
		Tag:      newReleaseNotes.NewControllerVersion,
	}
	c2 := ControllerImage{
		Digest:   controllerChrootDigest,
		Registry: INGRESS_REGISTRY,
		Name:     "controller-chroot",
		Tag:      newReleaseNotes.NewControllerVersion,
	}
	newReleaseNotes.ControllerImages = append(newReleaseNotes.ControllerImages, c1)
	newReleaseNotes.ControllerImages = append(newReleaseNotes.ControllerImages, c2)
	Debug("New Release Controller Images %s %s", newReleaseNotes.ControllerImages[0].Digest, newReleaseNotes.ControllerImages[1].Digest)

	if DEBUG {
		newReleaseNotes.printRelease()
	}

	//write it all out to the changelog file
	newReleaseNotes.template()

	return &newReleaseNotes, nil
}

func (i ControllerImage) print() string {
	return fmt.Sprintf("%s/%s:%s@%s", i.Registry, i.Name, i.Tag, i.Digest)
}

func (r ReleaseNote) template() {
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
		ErrorF("executing template:", err)
	}
}

func (r ReleaseNote) helmTemplate() {
	// Files are provided as a slice of strings.
	changelogTemplate, err := os.ReadFile("charts/ingress-nginx/changelog.md.gotmpl")
	if err != nil {
		ErrorF("Could not read changelog template file %s", err)
	}
	Debug("ChangeLog Templates %s", string(changelogTemplate))
	t := template.Must(template.New("changelog").Parse(string(changelogTemplate)))
	// create a new file
	file, err := os.Create(fmt.Sprintf("charts/ingress-nginx/changelog/Changelog-%s.md", r.Version))
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
func (Release) Latest() error {
	r, _, err := latestRelease()
	if err != nil {
		ErrorF("Latest Release error %s", err)
		return err
	}
	Info("Latest Release %v", r.String())
	return nil
}

func (Release) ReleaseByTag(tag string) error {
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
func (Release) Copy() error {
	ghClient := githubClient()
	kRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, "kubernetes", "ingress-nginx")
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

	sRelease, _, err = ghClient.Repositories.CreateRelease(ctx, "strongjz", "ingress-nginx", sRelease)
	if err != nil {
		ErrorF("Creating Strongjz release %s", err)
		return err
	}
	Info("Copied over Kubernetes Release %v to Strongjz %v", &kRelease.Name, &sRelease.Name)
	return nil
}
