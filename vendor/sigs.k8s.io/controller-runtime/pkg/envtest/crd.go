/*
Copyright 2018 The Kubernetes Authors.

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

package envtest

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

// CRDInstallOptions are the options for installing CRDs
type CRDInstallOptions struct {
	// Paths is the path to the directory containing CRDs
	Paths []string

	// CRDs is a list of CRDs to install
	CRDs []*apiextensionsv1beta1.CustomResourceDefinition

	// ErrorIfPathMissing will cause an error if a Path does not exist
	ErrorIfPathMissing bool

	// maxTime is the max time to wait
	maxTime time.Duration

	// pollInterval is the interval to check
	pollInterval time.Duration
}

const defaultPollInterval = 100 * time.Millisecond
const defaultMaxWait = 10 * time.Second

// InstallCRDs installs a collection of CRDs into a cluster by reading the crd yaml files from a directory
func InstallCRDs(config *rest.Config, options CRDInstallOptions) ([]*apiextensionsv1beta1.CustomResourceDefinition, error) {
	defaultCRDOptions(&options)

	// Read the CRD yamls into options.CRDs
	if err := readCRDFiles(&options); err != nil {
		return nil, err
	}

	// Create the CRDs in the apiserver
	if err := CreateCRDs(config, options.CRDs); err != nil {
		return options.CRDs, err
	}

	// Wait for the CRDs to appear as Resources in the apiserver
	if err := WaitForCRDs(config, options.CRDs, options); err != nil {
		return options.CRDs, err
	}

	return options.CRDs, nil
}

// readCRDFiles reads the directories of CRDs in options.Paths and adds the CRD structs to options.CRDs
func readCRDFiles(options *CRDInstallOptions) error {
	if len(options.Paths) > 0 {
		for _, path := range options.Paths {
			if _, err := os.Stat(path); !options.ErrorIfPathMissing && os.IsNotExist(err) {
				continue
			}
			new, err := readCRDs(path)
			if err != nil {
				return err
			}
			options.CRDs = append(options.CRDs, new...)
		}
	}
	return nil
}

// defaultCRDOptions sets the default values for CRDs
func defaultCRDOptions(o *CRDInstallOptions) {
	if o.maxTime == 0 {
		o.maxTime = defaultMaxWait
	}
	if o.pollInterval == 0 {
		o.pollInterval = defaultPollInterval
	}
}

// WaitForCRDs waits for the CRDs to appear in discovery
func WaitForCRDs(config *rest.Config, crds []*apiextensionsv1beta1.CustomResourceDefinition, options CRDInstallOptions) error {
	// Add each CRD to a map of GroupVersion to Resource
	waitingFor := map[schema.GroupVersion]*sets.String{}
	for _, crd := range crds {
		gvs := []schema.GroupVersion{}
		if crd.Spec.Version != "" {
			gvs = append(gvs, schema.GroupVersion{Group: crd.Spec.Group, Version: crd.Spec.Version})
		}
		for _, ver := range crd.Spec.Versions {
			if ver.Served {
				gvs = append(gvs, schema.GroupVersion{Group: crd.Spec.Group, Version: ver.Name})
			}
		}
		for _, gv := range gvs {
			log.V(1).Info("adding API in waitlist", "GV", gv)
			if _, found := waitingFor[gv]; !found {
				// Initialize the set
				waitingFor[gv] = &sets.String{}
			}
			// Add the Resource
			waitingFor[gv].Insert(crd.Spec.Names.Plural)
		}
	}

	// Poll until all resources are found in discovery
	p := &poller{config: config, waitingFor: waitingFor}
	return wait.PollImmediate(options.pollInterval, options.maxTime, p.poll)
}

// poller checks if all the resources have been found in discovery, and returns false if not
type poller struct {
	// config is used to get discovery
	config *rest.Config

	// waitingFor is the map of resources keyed by group version that have not yet been found in discovery
	waitingFor map[schema.GroupVersion]*sets.String
}

// poll checks if all the resources have been found in discovery, and returns false if not
func (p *poller) poll() (done bool, err error) {
	// Create a new clientset to avoid any client caching of discovery
	cs, err := clientset.NewForConfig(p.config)
	if err != nil {
		return false, err
	}

	allFound := true
	for gv, resources := range p.waitingFor {
		// All resources found, do nothing
		if resources.Len() == 0 {
			delete(p.waitingFor, gv)
			continue
		}

		// Get the Resources for this GroupVersion
		// TODO: Maybe the controller-runtime client should be able to do this...
		resourceList, err := cs.Discovery().ServerResourcesForGroupVersion(gv.Group + "/" + gv.Version)
		if err != nil {
			return false, nil
		}

		// Remove each found resource from the resources set that we are waiting for
		for _, resource := range resourceList.APIResources {
			resources.Delete(resource.Name)
		}

		// Still waiting on some resources in this group version
		if resources.Len() != 0 {
			allFound = false
		}
	}
	return allFound, nil
}

// CreateCRDs creates the CRDs
func CreateCRDs(config *rest.Config, crds []*apiextensionsv1beta1.CustomResourceDefinition) error {
	cs, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}

	// Create each CRD
	for _, crd := range crds {
		log.V(1).Info("installing CRD", "crd", crd.Name)
		if _, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
			return err
		}
	}
	return nil
}

// readCRDs reads the CRDs from files and Unmarshals them into structs
func readCRDs(path string) ([]*apiextensionsv1beta1.CustomResourceDefinition, error) {
	// Get the CRD files
	var files []os.FileInfo
	var err error
	log.V(1).Info("reading CRDs from path", "path", path)
	if files, err = ioutil.ReadDir(path); err != nil {
		return nil, err
	}

	// White list the file extensions that may contain CRDs
	crdExts := sets.NewString(".json", ".yaml", ".yml")

	var crds []*apiextensionsv1beta1.CustomResourceDefinition
	for _, file := range files {
		// Only parse whitelisted file types
		if !crdExts.Has(filepath.Ext(file.Name())) {
			continue
		}

		// Unmarshal CRDs from file into structs
		docs, err := readDocuments(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		for _, doc := range docs {
			crd := &apiextensionsv1beta1.CustomResourceDefinition{}
			if err = yaml.Unmarshal(doc, crd); err != nil {
				return nil, err
			}

			// Check that it is actually a CRD
			if crd.Spec.Names.Kind == "" || crd.Spec.Group == "" {
				continue
			}
			crds = append(crds, crd)
		}

		log.V(1).Info("read CRDs from file", "file", file.Name())
	}
	return crds, nil
}

// readDocuments reads documents from file
func readDocuments(fp string) ([][]byte, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	docs := [][]byte{}
	reader := k8syaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(b)))
	for {
		// Read document
		doc, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}
