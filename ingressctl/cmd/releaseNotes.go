package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReleaseNote - All the pieces of information/documents that get updated during a release
type ReleaseNote struct {
	Version                   string            //released version
	NewControllerVersion      string            //the new controller version being release
	PreviousControllerVersion string            //the previous controller tag/release
	ControllerImages          []ControllerImage //the full image digests
	DepUpdates                []string          //list of dependabot updates to put in the changelog
	Updates                   []string          //updates with no category
	HelmUpdates               []string          //updates to the ingress-nginx helm chart
	NewHelmChartVersion       string            //update to the helm chart version
	PreviousHelmChartVersion  string            //previous helm chart version
}

// Generate Release Notes
func ControllerReleaseNotes(newVersion string) (*ReleaseNote, error) {
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

	allUpdates, depUpdates, helmUpdates := getCommitUpdates(newVersion)

	newReleaseNotes.Updates = allUpdates
	newReleaseNotes.DepUpdates = depUpdates
	newReleaseNotes.HelmUpdates = helmUpdates

	// Get the latest controller image digests from k8s.io promoter
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
		Name:     "ingress-nginx/controller",
		Tag:      fmt.Sprintf("v%s", newReleaseNotes.Version),
	}

	c2 := ControllerImage{
		Digest:   controllerChrootDigest,
		Registry: INGRESS_REGISTRY,
		Name:     "ingress-nginx/controller-chroot",
		Tag:      fmt.Sprintf("v%s", newReleaseNotes.Version),
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

func getCommitUpdates(newVersion string) ([]string, []string, []string) {

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

	helmUpdates = append(helmUpdates, fmt.Sprintf("Update Ingress-Nginx version %s", newVersion))
	return allUpdates, depUpdates, helmUpdates
}
