/*
Copyright 2020 The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"strings"

	networking "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress"
)

var (
	pathTypeExact  = networking.PathTypeExact
	pathTypePrefix = networking.PathTypePrefix
)

// updateServerLocations inspects the generated locations configuration for a server
// normalizing the path and adding an additional exact location when is possible
func updateServerLocations(locations []*ingress.Location) []*ingress.Location {
	newLocations := []*ingress.Location{}

	// get Exact locations to check if one already exists
	exactLocations := map[string]*ingress.Location{}
	for _, location := range locations {
		if *location.PathType == pathTypeExact {
			exactLocations[location.Path] = location
		}
	}

	for _, location := range locations {
		// location / does not require any update
		if location.Path == rootLocation {
			newLocations = append(newLocations, location)
			continue
		}

		location.IngressPath = location.Path

		// only Prefix locations could require an additional location block
		if *location.PathType != pathTypePrefix {
			newLocations = append(newLocations, location)
			continue
		}

		// locations with rewrite or using regular expressions are not modified
		if needsRewrite(location) || location.Rewrite.UseRegex {
			newLocations = append(newLocations, location)
			continue
		}

		// If exists an Exact location is not possible to create a new one.
		if _, alreadyExists := exactLocations[location.Path]; alreadyExists {
			// normalize path. Must end in /
			location.Path = normalizePrefixPath(location.Path)
			newLocations = append(newLocations, location)
			continue
		}

		var el ingress.Location = *location

		// normalize path. Must end in /
		location.Path = normalizePrefixPath(location.Path)
		newLocations = append(newLocations, location)

		// add exact location
		exactLocation := &el
		exactLocation.PathType = &pathTypeExact

		newLocations = append(newLocations, exactLocation)
	}

	return newLocations
}

func normalizePrefixPath(path string) string {
	if path == rootLocation {
		return rootLocation
	}

	if !strings.HasSuffix(path, "/") {
		return fmt.Sprintf("%v/", path)
	}

	return path
}

func needsRewrite(location *ingress.Location) bool {
	if len(location.Rewrite.Target) > 0 && location.Rewrite.Target != location.Path {
		return true
	}

	return false
}
