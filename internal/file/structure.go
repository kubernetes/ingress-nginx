/*
Copyright 2017 The Kubernetes Authors.

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

package file

import (
	"os"

	"github.com/pkg/errors"
)

const (
	// AuthDirectory default directory used to store files
	// to authenticate request
	AuthDirectory = "/etc/ingress-controller/auth"

	// DefaultSSLDirectory defines the location where the SSL certificates will be generated
	// This directory contains all the SSL certificates that are specified in Ingress rules.
	// The name of each file is <namespace>-<secret name>.pem. The content is the concatenated
	// certificate and key.
	DefaultSSLDirectory = "/etc/ingress-controller/ssl"
)

var (
	directories = []string{
		DefaultSSLDirectory,
		AuthDirectory,
	}
)

// CreateRequiredDirectories verifies if the required directories to
// start the ingress controller exist and creates the missing ones.
func CreateRequiredDirectories() error {
	for _, directory := range directories {
		_, err := os.Stat(directory)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(directory, ReadWriteByUser)
				if err != nil {
					return errors.Wrapf(err, "creating directory '%v'", directory)
				}

				continue
			}

			return errors.Wrapf(err, "checking directory %v", directory)
		}
	}

	return nil
}
