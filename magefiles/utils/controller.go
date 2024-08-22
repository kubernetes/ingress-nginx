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

	"github.com/google/go-github/v48/github"
)

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

func (i ControllerImage) print() string {
	return fmt.Sprintf("%s/%s:%s@%s", i.Registry, i.Name, i.Tag, i.Digest)
}

func FindImageDigest(yaml ImageYamls, image, version string) string {
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
