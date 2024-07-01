/*
Copyright 2021 The Kubernetes Authors.

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

package k8sclient

import (
	"sync"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	client "k8s.io/client-go/kubernetes"
)

var (
	once         sync.Once
	globalClient *client.Clientset
)

func GlobalClient(flags *genericclioptions.ConfigFlags) *client.Clientset {
	once.Do(func() {
		rawConfig, err := flags.ToRESTConfig()
		if err != nil {
			panic(err)
		}
		globalClient, err = client.NewForConfig(rawConfig)
		if err != nil {
			panic(err)
		}
	})
	return globalClient
}
