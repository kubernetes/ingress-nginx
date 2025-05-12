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
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	utils "k8s.io/ingress-nginx/magefiles/utils"
)

type Release mg.Namespace

var (
	INGRESS_ORG               = "kubernetes"                                // the owner so we can test from forks
	INGRESS_REPO              = "ingress-nginx"                             // the repo to pull from
	RELEASE_BRANCH            = "main"                                      // we only release from main
	GITHUB_TOKEN              string                                        // the Google/gogithub lib needs an PAT to access the GitHub API
	K8S_IO_ORG                = "kubernetes"                                // the owner or organization for the k8s.io repo
	K8S_IO_REPO               = "k8s.io"                                    // the repo that holds the images yaml for production promotion
	INGRESS_REGISTRY          = "registry.k8s.io"                           // Container registry for storage Ingress-nginx images
	KUSTOMIZE_INSTALL_VERSION = "sigs.k8s.io/kustomize/kustomize/v4@v4.5.4" // static deploys needs kustomize to generate the template
)

// ingress-nginx releases start with a TAG then a cloudbuild, then a promotion through a PR, this the location of that PR
var (
	IMAGES_YAML = "https://raw.githubusercontent.com/kubernetes/k8s.io/main/registry.k8s.io/images/k8s-staging-ingress-nginx/images.yaml"
	ctx         = context.Background() // Context used for GitHub Client
)

const (
	INDEX_DOCS = "docs/deploy/index.md" // index.md has a version of the controller and needs to updated
	CHANGELOG  = "Changelog.md"         // Name of the changelog
)

// init will set the GitHub token from the committers/releasers env var
func init() {
	GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
}

// Release Create a new release of ingress nginx controller
func (Release) NewRelease(version string) {
	newRelease(version, "")
}

func (Release) NewReleaseFromOld(version, oldversion string) {
	newRelease(version, oldversion)
}

func (Release) E2EDocs() {
	e2edocs, err := utils.GenerateE2EDocs()
	utils.CheckIfError(err, "error on template")
	err = os.WriteFile("docs/e2e-tests.md", []byte(e2edocs), 0644)
	utils.CheckIfError(err, "Could not write new e2e test file ")
}

func newRelease(version, oldversion string) {
	// newRelease := Release{}

	// update ingress-nginx version
	// This is the step that kicks all the release process
	// it is already done, so it kicks off the gcloud build of the controller images
	// mg.Deps(mg.F(Tag.BumpNginx, version))

	tag, err := getIngressNGINXVersion()
	utils.CheckIfError(err, "RELEASE Retrieving the current Ingress Nginx Version")

	utils.Info("RELEASE Checking Current Version %s to New Version %s", tag, version)
	// if the version were upgrading does not match the TAG file, lets update the TAG file
	if tag[1:] != version {
		utils.Warning("RELEASE Ingress Nginx TAG %s and new version %s do not match", tag, version)
		mg.Deps(mg.F(Tag.BumpNginx, fmt.Sprintf("v%s", version)))
	}

	// update git controller tag controller-v$version
	mg.Deps(mg.F(Tag.NewControllerTag, version))

	// make release notes
	releaseNotes, err := makeReleaseNotes(version, oldversion)
	utils.CheckIfError(err, "RELEASE Creating Release Notes for version %s", version)
	utils.Info("RELEASE Release Notes %s completed", releaseNotes.Version)

	// update chart values.yaml new controller tag and image digest
	releaseNotes.PreviousHelmChartVersion = currentChartVersion()

	// controller tag
	updateChartValue("controller.image.tag", fmt.Sprintf("v%s", releaseNotes.Version))
	utils.Debug("releaseNotes.ControllerImages[0].Name %s", releaseNotes.ControllerImages[0].Name)
	utils.Debug("releaseNotes.ControllerImages[1].Name %s", releaseNotes.ControllerImages[1].Name)
	// controller digest
	if releaseNotes.ControllerImages[0].Name == "ingress-nginx/controller" {
		utils.Debug("Updating Chart Value %s with %s", "controller.image.digest", releaseNotes.ControllerImages[0].Digest)
		updateChartValue("controller.image.digest", releaseNotes.ControllerImages[0].Digest)
	}
	// controller chroot digest
	if releaseNotes.ControllerImages[1].Name == "ingress-nginx/controller-chroot" {
		utils.Debug("Updating Chart Value %s with %s", "controller.image.digestChroot", releaseNotes.ControllerImages[1].Digest)
		updateChartValue("controller.image.digestChroot", releaseNotes.ControllerImages[1].Digest)
	}

	// update helm chart app version
	mg.Deps(mg.F(Helm.UpdateVersion, version))

	releaseNotes.NewHelmChartVersion = currentChartVersion()

	// update helm chart release notes
	updateChartReleaseNotes(releaseNotes.HelmUpdates)

	// Run helm docs update
	utils.CheckIfError(runHelmDocs(), "Error Updating Helm Docs ")

	releaseNotes.HelmTemplate()

	// update static manifest
	utils.CheckIfError(updateStaticManifest(), "Error Updating Static manifests")

	////update e2e docs
	mg.Deps(mg.F(Release.E2EDocs))

	// update documentation with ingress-nginx version
	utils.CheckIfError(updateIndexMD(releaseNotes.PreviousControllerVersion, releaseNotes.NewControllerVersion), "Error Updating %s", INDEX_DOCS)

	// keeping these manual for now
	// git commit TODO
	// make Pull Request TODO
	// make release TODO
	// mg.Deps(mg.F(Release.CreateRelease, version))
}

// the index.md doc needs the controller version updated
func updateIndexMD(old, new string) error {
	utils.Info("Updating Deploy docs with new version")
	data, err := os.ReadFile(INDEX_DOCS)
	utils.CheckIfError(err, "Could not read INDEX_DOCS file %s", INDEX_DOCS)
	datString := string(data)
	datString = strings.Replace(datString, old, new, -1)
	err = os.WriteFile(INDEX_DOCS, []byte(datString), 0644)
	if err != nil {
		utils.ErrorF("Could not write new %s %s", INDEX_DOCS, err)
		return err
	}
	return nil
}

// The static deploy scripts use kustomize to generate them, this function ensures kustomize is installed
func installKustomize() error {
	utils.Info("Install Kustomize")
	g0 := sh.RunCmd("go")
	// somewhere in your main code
	err := g0("install", KUSTOMIZE_INSTALL_VERSION)
	if err != nil {
		return err
	}
	return nil
}

func updateStaticManifest() error {
	utils.CheckIfError(installKustomize(), "error installing kustomize")
	// hack/generate-deploy-scripts.sh
	err := sh.RunV("./hack/generate-deploy-scripts.sh")
	if err != nil {
		return err
	}
	return nil
}

//// CreateRelease Creates a new GitHub Release
//func (Release) CreateRelease(name string) {
//	releaser, err := gh_release.NewReleaser(INGRESS_ORG, INGRESS_REPO, GITHUB_TOKEN)
//	utils.CheckIfError(err, "GitHub Release Client error")
//	newRelease, err := releaser.Create(fmt.Sprintf("controller-%s", name))
//	utils.CheckIfError(err, "Create release error")
//	utils.Info("New Release: Tag %v, ID: %v", newRelease.TagName, newRelease.ID)
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
	commitLog := commitsBetweenTags("", "")
	for i, s := range commitLog {
		utils.Info("#%v Version %v", i, s)
	}
}

func commitsBetweenTags(newversion, oldversion string) []string {
	var newTag, oldTag string
	tags := getAllControllerTags()
	newTag, oldTag = tags[0], tags[1]
	if newversion != "" {
		newTag = newversion
	}
	if oldversion != "" {
		oldTag = oldversion
	}

	utils.Info("Getting Commits between %v and %v", newTag, oldTag)
	commitLog, err := git("log", "--full-history", "--pretty", "--oneline", fmt.Sprintf("%v..%v", oldTag, newTag))

	if commitLog == "" {
		utils.Warning("All Controller Tags is empty")
	}
	utils.CheckIfError(err, "Retrieving Commit log")
	return strings.Split(commitLog, "\n")
}

// Generate Release Notes
func (Release) ReleaseNotes(newVersion string) error {
	notes, err := makeReleaseNotes(newVersion, "")
	utils.CheckIfError(err, "Creating Release Notes for version %s", newVersion)
	utils.Info("Release Notes %s completed", notes.Version)
	return nil
}

func makeReleaseNotes(newVersion, oldVersion string) (*utils.ReleaseNote, error) {
	newReleaseNotes := utils.ReleaseNote{}

	newReleaseNotes.Version = newVersion
	allControllerTags := getAllControllerTags()

	// new version
	newReleaseNotes.NewControllerVersion = allControllerTags[0]
	newControllerVersion := fmt.Sprintf("controller-v%s", newVersion)

	// the newControllerVersion should match the latest tag
	if newControllerVersion != allControllerTags[0] {
		return nil, fmt.Errorf("generating release new version %s didnt match the current latest tag %s", newControllerVersion, allControllerTags[0])
	}
	// previous version
	newReleaseNotes.PreviousControllerVersion = allControllerTags[1]
	if oldVersion != "" {
		newReleaseNotes.PreviousControllerVersion = oldVersion
	}

	utils.Info("New Version: %s Old Version: %s", newReleaseNotes.NewControllerVersion, newReleaseNotes.PreviousControllerVersion)

	commits := commitsBetweenTags(newReleaseNotes.NewControllerVersion, newReleaseNotes.PreviousControllerVersion)

	// dependency_updates
	// all_updates
	var allUpdates []string
	var depUpdates []string
	var helmUpdates []string
	prRegex := regexp.MustCompile(`\(#\d+\)`)
	depBot := regexp.MustCompile(`^(\w){1,10} Bump `)
	helmRegex := regexp.MustCompile("helm|chart")
	for i, s := range commits {
		// matches on PR
		if prRegex.Match([]byte(s)) {
			// matches a dependant bot update
			if depBot.Match([]byte(s)) { //
				utils.Debug("#%v DEPENDABOT %v", i, s)
				u := strings.SplitN(s, " ", 2)
				depUpdates = append(depUpdates, u[1])
			} else { // add it to the all updates slice
				utils.Debug("#%v ALL UPDATES %v", i, s)
				u := strings.SplitN(s, " ", 2)
				allUpdates = append(allUpdates, u[1])

				// helm chart updates
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

	// controller_image_digests
	imagesYaml, err := utils.DownloadFile(IMAGES_YAML)
	if err != nil {
		utils.ErrorF("Could not download file %s : %s", IMAGES_YAML, err)
		return nil, err
	}
	utils.Debug("%s", imagesYaml)

	data := utils.ImageYamls{}

	err = yaml.Unmarshal([]byte(imagesYaml), &data)
	if err != nil {
		utils.ErrorF("Could not unmarshal images yaml %s", err)
		return nil, err
	}

	// controller
	controllerDigest := utils.FindImageDigest(data, "controller", newVersion)
	if len(controllerDigest) == 0 {
		utils.ErrorF("Controller Digest could not be found")
		return nil, errors.New("controller digest could not be found")
	}

	controllerChrootDigest := utils.FindImageDigest(data, "controller-chroot", newVersion)
	if len(controllerChrootDigest) == 0 {
		utils.ErrorF("Controller Chroot Digest could not be found")
		return nil, errors.New("controller chroot digest could not be found")
	}

	utils.Debug("Latest Controller Digest %v", controllerDigest)
	utils.Debug("Latest Controller Chroot Digest %v", controllerChrootDigest)
	c1 := utils.ControllerImage{
		Digest:   controllerDigest,
		Registry: INGRESS_REGISTRY,
		Name:     "ingress-nginx/controller",
		Tag:      fmt.Sprintf("v%s", newReleaseNotes.Version),
	}

	c2 := utils.ControllerImage{
		Digest:   controllerChrootDigest,
		Registry: INGRESS_REGISTRY,
		Name:     "ingress-nginx/controller-chroot",
		Tag:      fmt.Sprintf("v%s", newReleaseNotes.Version),
	}

	newReleaseNotes.ControllerImages = append(newReleaseNotes.ControllerImages, c1)
	newReleaseNotes.ControllerImages = append(newReleaseNotes.ControllerImages, c2)
	utils.Debug("New Release Controller Images %s %s", newReleaseNotes.ControllerImages[0].Digest, newReleaseNotes.ControllerImages[1].Digest)

	if utils.DEBUG {
		newReleaseNotes.PrintRelease()
	}

	// write it all out to the changelog file
	newReleaseNotes.Template()

	return &newReleaseNotes, nil
}

// Latest returns latest Github Release
func (Release) Latest() error {
	r, _, err := latestRelease()
	if err != nil {
		utils.ErrorF("Latest Release error %s", err)
		return err
	}
	utils.Info("Latest Release %v", r.String())
	return nil
}

func (Release) ReleaseByTag(tag string) error {
	r, _, err := releaseByTag(tag)
	if err != nil {
		utils.ErrorF("Release retrieve tag %s error %s", tag, err)
		return err
	}

	utils.Info("Latest Release %v", r.String())

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
