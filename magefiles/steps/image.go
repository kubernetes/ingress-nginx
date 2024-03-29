/*
Copyright 2024 The Kubernetes Authors.

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
	"runtime"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	utils "k8s.io/ingress-nginx/magefiles/utils"
)

type Image mg.Namespace

var TAG string = "0.0"
var ARCH string = "arm64"
var PKG string = "k8s.io/ingress-nginx"
var PLATFORM string = "arm64"
var BUILD_ID string = "UNSET"

// COMMIT_SHA ?= git-$(shell git rev-parse --short HEAD)
var COMMIT_SHA string = ""
var REGISTRY string = "gcr.io/k8s-staging-ingress-nginx"

// BASE_IMAGE ?= $(shell cat NGINX_BASE)
var BASE_IMAGE string = "registry.k8s.io/ingress-nginx/nginx-1.25:v0.0.5"

func (Image) Create() error {
	if ARCH != runtime.GOARCH {
		ARCH = runtime.GOARCH
	}
	
	COMMIT_SHA, err := getCommitSHA()
	utils.CheckIfError(err, "Error Getting Commit sha")

	err = sh.RunV("docker", "build",
		"--no-cache",
		"--build-arg", fmt.Sprintf("BASE_IMAGE=%s", BASE_IMAGE),
		"--build-arg", fmt.Sprintf("VERSION=%s", TAG),
		"--build-arg", fmt.Sprintf("TARGETARCH=%s", ARCH),
		"--build-arg", fmt.Sprintf("COMMIT_SHA=%s", COMMIT_SHA),
		"--build-arg", fmt.Sprintf("BUILD_ID=%s", BUILD_ID),
		"-t", fmt.Sprintf("%s/controller:%s", REGISTRY, TAG),
		"rootfs",
	)
	if err != nil {
		return err
	}
	return nil
}

func (Image) Load() error {
	workers, err := getWorkers()
	utils.CheckIfError(err, "Error getWorkers")

	err = sh.RunV("kind", "load", "docker-image", fmt.Sprintf("--name=%s", KIND_CLUSTER_NAME), fmt.Sprintf("--nodes=%s", workers), fmt.Sprintf("%s/controller:%s", REGISTRY, TAG))
	utils.CheckIfError(err, "Error Loading controller onto kind cluster")
	return nil
}
