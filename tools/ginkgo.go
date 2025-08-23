//go:build ignore

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

// Package tools provides a way to manage tool dependencies for the project.
// This ensures that ginkgo version is managed through go.mod and can be
// automatically updated by dependabot without manual intervention.
package tools

import (
	_ "github.com/onsi/ginkgo/v2/ginkgo"
)

//go:generate go install -modfile=../go.mod github.com/onsi/ginkgo/v2/ginkgo
