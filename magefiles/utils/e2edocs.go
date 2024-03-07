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
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

//go:embed templates/e2edocs.tpl
var tplContent embed.FS

var skipFiles = []string{
	"test/e2e/framework/framework.go",
	"test/e2e/e2e.go",
	"test/e2e/e2e_test.go",
}

const (
	testDir  = "test/e2e"
	describe = "Describe"
	URL      = "https://github.com/kubernetes/ingress-nginx/tree/main/"
)

var betweenquotes = regexp.MustCompile(`("|\')(?P<TestDescription>.*)("|\')`)

type E2ETemplate struct {
	URL   string
	Tests []string
}

func getDescription(linetext, path, url string, lineN int, isDescription bool) string {
	var descriptionLine string
	prefix := "-"
	if isDescription {
		prefix = "###"
	}

	matches := betweenquotes.FindStringSubmatch(linetext)
	contentIndex := betweenquotes.SubexpIndex("TestDescription")
	if len(matches) < 2 || contentIndex == -1 {
		return ""
	}

	fileName := fmt.Sprintf("%s/%s", url, path)
	descriptionLine = fmt.Sprintf("%s [%s](%s#L%d)", prefix, matches[contentIndex], fileName, lineN)

	return descriptionLine
}

func containsGinkgoTest(line string) bool {
	if !strings.Contains(line, describe) && !strings.Contains(line, "It") {
		return false
	}
	return strings.Contains(line, "func() {")
}

func (t *E2ETemplate) walkE2eDir(path string, d fs.DirEntry, errAggregated error) error {
	if errAggregated != nil {
		return errAggregated
	}
	// Remove ignored files or non .go files
	if d.IsDir() || slices.Contains(skipFiles, path) || !strings.HasSuffix(path, ".go") {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fileScanner := bufio.NewScanner(bytes.NewReader(content))

	fileScanner.Split(bufio.ScanLines)

	tests := make([]string, 0)
	lineN := 0
	for fileScanner.Scan() {
		lineN = lineN + 1
		if !containsGinkgoTest(fileScanner.Text()) {
			continue
		}

		line := getDescription(fileScanner.Text(), path, t.URL, lineN, strings.Contains(fileScanner.Text(), describe))
		if line != "" {
			tests = append(tests, line)
		}
	}
	t.Tests = append(t.Tests, tests...)
	return nil
}

func GenerateE2EDocs() (string, error) {
	e2etpl := &E2ETemplate{URL: URL}
	err := filepath.WalkDir(testDir, e2etpl.walkE2eDir)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("e2edocs.tpl").ParseFS(tplContent, "templates/e2edocs.tpl")
	if err != nil {
		return "", fmt.Errorf("error parsing the template file: %s", err)
	}

	tplBuff := new(bytes.Buffer)
	err = tmpl.Execute(tplBuff, e2etpl)
	if err != nil {
		return "", err
	}
	return tplBuff.String(), nil
}
