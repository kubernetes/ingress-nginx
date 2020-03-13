/*
Copyright 2019 Google LLC

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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/mmarkdown/mmark/mparser"
)

const (
	startTOC = "<!-- toc -->"
	endTOC   = "<!-- /toc -->"
)

type options struct {
	dryrun     bool
	inplace    bool
	skipPrefix bool
}

var defaultOptions options

func init() {
	flag.BoolVar(&defaultOptions.dryrun, "dryrun", false, "Whether to check for changes to TOC, rather than overwriting. Requires --inplace flag.")
	flag.BoolVar(&defaultOptions.inplace, "inplace", false, "Whether to edit the file in-place, or output to STDOUT. Requires toc tags to be present.")
	flag.BoolVar(&defaultOptions.skipPrefix, "skip-prefix", true, "Whether to ignore any headers before the opening toc tag.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [FILE]...\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Generate a table of contents for a markdown file (github flavor).\n")
		fmt.Fprintf(flag.CommandLine.Output(), "TOC may be wrapped in a pair of tags to allow in-place updates: <!-- toc --><!-- /toc -->\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if err := validateArgs(defaultOptions, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	hadError := false
	for _, file := range flag.Args() {
		toc, err := run(file, defaultOptions)
		if err != nil {
			log.Printf("%s: %v", file, err)
			hadError = true
		} else if !defaultOptions.inplace {
			fmt.Println(toc)
		}
	}

	if hadError {
		os.Exit(1)
	}
}

func validateArgs(opts options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify at least 1 file")
	}
	if !opts.inplace && len(args) > 1 {
		return fmt.Errorf("non-inplace updates require exactly 1 file")
	}
	if opts.dryrun && !opts.inplace {
		return fmt.Errorf("--dryrun requires --inplace")
	}
	return nil
}

// run the TOC generator on file with options.
// Returns the generated toc, and any error.
func run(file string, opts options) (string, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("unable to read %s: %v", file, err)
	}

	start := bytes.Index(raw, []byte(startTOC))
	end := bytes.Index(raw, []byte(endTOC))
	if tocTagRequired(opts) {
		if start == -1 {
			return "", fmt.Errorf("missing opening TOC tag")
		}
		if end == -1 {
			return "", fmt.Errorf("missing closing TOC tag")
		}
		if end < start {
			return "", fmt.Errorf("TOC closing tag before start tag")
		}
	}

	var prefix, doc []byte
	// skipPrefix is only used when toc tags are present.
	if opts.skipPrefix && start != -1 && end != -1 {
		prefix = raw[:start]
		doc = raw[end:]
	} else {
		doc = raw
	}
	toc, err := generateTOC(prefix, doc)
	if err != nil {
		return toc, fmt.Errorf("failed to generate toc: %v", err)
	}

	if !opts.inplace {
		return toc, err
	}

	realStart := start + len(startTOC)
	oldTOC := string(raw[realStart:end])
	if strings.TrimSpace(oldTOC) == strings.TrimSpace(toc) {
		// No changes required.
		return toc, nil
	} else if opts.dryrun {
		return toc, fmt.Errorf("changes found:\n%s", toc)
	}

	err = atomicWrite(file,
		string(raw[:start]),
		startTOC+"\n",
		string(toc),
		string(raw[end:]),
	)
	return toc, err
}

// atomicWrite writes the chunks sequentially to the filePath.
// A temporary file is used so no changes are made to the original in the case of an error.
func atomicWrite(filePath string, chunks ...string) error {
	tmpPath := filePath + "_tmp"
	tmp, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("unable to open tepmorary file %s: %v", tmpPath, err)
	}

	// Cleanup
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	for _, chunk := range chunks {
		if _, err := tmp.WriteString(chunk); err != nil {
			return err
		}
	}

	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filePath)
}

// parse parses a raw markdown document to an AST.
func parse(b []byte) ast.Node {
	p := parser.NewWithExtensions(parser.CommonExtensions)
	p.Opts = parser.Options{
		// mparser is required for parsing the --- title blocks
		ParserHook: mparser.Hook,
	}
	return p.Parse(b)
}

func generateTOC(prefix []byte, doc []byte) (string, error) {
	prefixMd := parse(prefix)
	anchors := make(anchorGen)
	// Start counting anchors from the beginning of the doc.
	walkHeadings(prefixMd, func(heading *ast.Heading) {
		anchors.mkAnchor(asText(heading))
	})

	md := parse(doc)

	baseLvl := headingBase(md)
	toc := &bytes.Buffer{}
	htmlRenderer := html.NewRenderer(html.RendererOptions{})
	walkHeadings(md, func(heading *ast.Heading) {
		anchor := anchors.mkAnchor(asText(heading))
		content := headingBody(htmlRenderer, heading)
		fmt.Fprintf(toc, "%s- [%s](#%s)\n", strings.Repeat("  ", heading.Level-baseLvl), content, anchor)
	})

	return string(toc.Bytes()), nil
}

func tocTagRequired(opts options) bool {
	return opts.inplace
}

type headingFn func(heading *ast.Heading)

// walkHeadings runs the heading function on each heading in the parsed markdown document.
func walkHeadings(doc ast.Node, headingFn headingFn) error {
	var err error
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext // Don't care about closing the heading section.
		}

		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.GoToNext // Ignore non-heading nodes.
		}

		if heading.IsTitleblock {
			return ast.GoToNext // Ignore title blocks (the --- section)
		}

		headingFn(heading)

		return ast.GoToNext
	})
	return err
}

func asText(node ast.Node) string {
	var text string
	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext // Don't care about closing the heading section.
		}
		t, ok := node.(*ast.Text)
		if !ok {
			return ast.GoToNext // Ignore non-text nodes.
		}

		text += string(t.AsLeaf().Literal)
		return ast.GoToNext
	})
	return text
}

// Renders the heading body as HTML
func headingBody(renderer *html.Renderer, heading *ast.Heading) string {
	var buf bytes.Buffer
	for _, child := range heading.Children {
		ast.WalkFunc(child, func(node ast.Node, entering bool) ast.WalkStatus {
			return renderer.RenderNode(&buf, node, entering)
		})
	}
	return strings.TrimSpace(buf.String())
}

// headingBase finds the minimum heading level. This is useful for normalizing indentation, such as
// when a top-level heading is skipped in the prefix.
func headingBase(doc ast.Node) int {
	baseLvl := math.MaxInt32
	walkHeadings(doc, func(heading *ast.Heading) {
		if baseLvl > heading.Level {
			baseLvl = heading.Level
		}
	})
	return baseLvl
}

// Match punctuation that is filtered out from anchor IDs.
var punctuation = regexp.MustCompile(`[^\w\- ]`)

// anchorGen is used to generate heading anchor IDs, using the github-flavored markdown syntax.
type anchorGen map[string]int

func (a anchorGen) mkAnchor(text string) string {
	text = strings.ToLower(text)
	text = punctuation.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, " ", "-")
	idx := a[text]
	a[text] = idx + 1
	if idx > 0 {
		return fmt.Sprintf("%s-%d", text, idx)
	}
	return text
}
