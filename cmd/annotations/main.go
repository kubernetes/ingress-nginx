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

package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/template"

	anns "k8s.io/ingress-nginx/internal/ingress/annotations"
)

type Documentation struct {
	Group      string
	Annotation string
	Risk       string
	Scope      string
}

var output string

//go:embed annotations.tmpl
var content embed.FS

func main() {
	flag.StringVar(&output, "output", "", "where to write documentation")
	flag.Parse()
	if output == "" {
		panic(fmt.Errorf("output field is required"))
	}
	docEntries := make([]Documentation, 0)
	annotationFactory := anns.NewAnnotationFactory(nil)
	for group, val := range annotationFactory {
		annotations := val.GetDocumentation()
		intermediateDocs := make([]Documentation, len(annotations))
		i := 0
		for annotation, values := range annotations {
			doc := Documentation{
				Group:      group,
				Annotation: annotation,
				Scope:      string(values.Scope),
				Risk:       values.Risk.ToString(),
			}
			intermediateDocs[i] = doc
			i++
		}
		slices.SortStableFunc(intermediateDocs, func(a, b Documentation) int {
			return strings.Compare(a.Annotation, b.Annotation)
		})
		docEntries = append(docEntries, intermediateDocs...)
	}
	slices.SortStableFunc(docEntries, func(a, b Documentation) int {
		return strings.Compare(a.Group, b.Group)
	})

	tmpl, err := template.New("annotations.tmpl").ParseFS(content, "annotations.tmpl")
	if err != nil {
		panic(fmt.Errorf("error parsing template: %s", err))
	}

	tplBuffer := new(bytes.Buffer)
	err = tmpl.Execute(tplBuffer, docEntries)
	if err != nil {
		panic(err)
	}
	tplBuffer.WriteString("\n")

	//nolint:gosec // no need to check file permission here
	if err := os.WriteFile(output, tplBuffer.Bytes(), 0o755); err != nil {
		panic(err)
	}
}
