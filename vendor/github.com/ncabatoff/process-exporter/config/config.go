package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	common "github.com/ncabatoff/process-exporter"
	"gopkg.in/yaml.v2"
)

type (
	Matcher interface {
		// Match returns empty string for no match, or the group name on success.
		Match(common.NameAndCmdline) bool
	}

	FirstMatcher []common.MatchNamer
	Config       struct {
		MatchNamers FirstMatcher
	}

	commMatcher struct {
		comms map[string]struct{}
	}

	exeMatcher struct {
		exes map[string]string
	}

	cmdlineMatcher struct {
		regexes  []*regexp.Regexp
		captures map[string]string
	}

	andMatcher []Matcher

	templateNamer struct {
		template *template.Template
	}

	matchNamer struct {
		andMatcher
		templateNamer
	}

	templateParams struct {
		Comm    string
		ExeBase string
		ExeFull string
		Matches map[string]string
	}
)

func (f FirstMatcher) MatchAndName(nacl common.NameAndCmdline) (bool, string) {
	for _, m := range f {
		if matched, name := m.MatchAndName(nacl); matched {
			return true, name
		}
	}
	return false, ""
}

func (m *matchNamer) MatchAndName(nacl common.NameAndCmdline) (bool, string) {
	if !m.Match(nacl) {
		return false, ""
	}

	matches := make(map[string]string)
	for _, m := range m.andMatcher {
		if mc, ok := m.(*cmdlineMatcher); ok {
			for k, v := range mc.captures {
				matches[k] = v
			}
		}
	}

	exebase, exefull := nacl.Name, nacl.Name
	if len(nacl.Cmdline) > 0 {
		exefull = nacl.Cmdline[0]
		exebase = filepath.Base(exefull)
	}

	var buf bytes.Buffer
	m.template.Execute(&buf, &templateParams{
		Comm:    nacl.Name,
		ExeBase: exebase,
		ExeFull: exefull,
		Matches: matches,
	})
	return true, buf.String()
}

func (m *commMatcher) Match(nacl common.NameAndCmdline) bool {
	_, found := m.comms[nacl.Name]
	return found
}

func (m *exeMatcher) Match(nacl common.NameAndCmdline) bool {
	if len(nacl.Cmdline) == 0 {
		return false
	}
	thisbase := filepath.Base(nacl.Cmdline[0])
	fqpath, found := m.exes[thisbase]
	if !found {
		return false
	}
	if fqpath == "" {
		return true
	}

	return fqpath == nacl.Cmdline[0]
}

func (m *cmdlineMatcher) Match(nacl common.NameAndCmdline) bool {
	for _, regex := range m.regexes {
		captures := regex.FindStringSubmatch(strings.Join(nacl.Cmdline, " "))
		if m.captures == nil {
			return false
		}
		subexpNames := regex.SubexpNames()
		if len(subexpNames) != len(captures) {
			return false
		}

		for i, name := range subexpNames {
			m.captures[name] = captures[i]
		}
	}
	return true
}

func (m andMatcher) Match(nacl common.NameAndCmdline) bool {
	for _, matcher := range m {
		if !matcher.Match(nacl) {
			return false
		}
	}
	return true
}

// ReadRecipesFile opens the named file and extracts recipes from it.
func ReadFile(cfgpath string) (*Config, error) {
	content, err := ioutil.ReadFile(cfgpath)
	if err != nil {
		return nil, err
	}
	return GetConfig(string(content))
}

// GetConfig extracts Config from content by parsing it as YAML.
func GetConfig(content string) (*Config, error) {
	var yamldata map[string]interface{}

	err := yaml.Unmarshal([]byte(content), &yamldata)
	if err != nil {
		return nil, err
	}
	yamlProcnames, ok := yamldata["process_names"]
	if !ok {
		return nil, fmt.Errorf("error parsing YAML config: no top-level 'process_names' key")
	}
	procnames, ok := yamlProcnames.([]interface{})
	if !ok {
		return nil, fmt.Errorf("error parsing YAML config: 'process_names' is not a list")
	}

	var cfg Config
	for i, procname := range procnames {
		mn, err := getMatchNamer(procname)
		if err != nil {
			return nil, fmt.Errorf("unable to parse process_name entry %d: %v", i, err)
		}
		cfg.MatchNamers = append(cfg.MatchNamers, mn)
	}

	return &cfg, nil
}

func getMatchNamer(yamlmn interface{}) (common.MatchNamer, error) {
	nm, ok := yamlmn.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("not a map")
	}

	var smap = make(map[string][]string)
	var nametmpl string
	for k, v := range nm {
		key, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("non-string key %v", k)
		}

		if key == "name" {
			value, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("non-string value %v for key %q", v, key)
			}
			nametmpl = value
		} else {
			vals, ok := v.([]interface{})
			if !ok {
				return nil, fmt.Errorf("non-string array value %v for key %q", v, key)
			}
			var strs []string
			for i, si := range vals {
				s, ok := si.(string)
				if !ok {
					return nil, fmt.Errorf("non-string value %v in list[%d] for key %q", v, i, key)
				}
				strs = append(strs, s)
			}
			smap[key] = strs
		}
	}

	var matchers andMatcher
	if comm, ok := smap["comm"]; ok {
		comms := make(map[string]struct{})
		for _, c := range comm {
			comms[c] = struct{}{}
		}
		matchers = append(matchers, &commMatcher{comms})
	}
	if exe, ok := smap["exe"]; ok {
		exes := make(map[string]string)
		for _, e := range exe {
			if strings.Contains(e, "/") {
				exes[filepath.Base(e)] = e
			} else {
				exes[e] = ""
			}
		}
		matchers = append(matchers, &exeMatcher{exes})
	}
	if cmdline, ok := smap["cmdline"]; ok {
		var rs []*regexp.Regexp
		for _, c := range cmdline {
			r, err := regexp.Compile(c)
			if err != nil {
				return nil, fmt.Errorf("bad cmdline regex %q: %v", c, err)
			}
			rs = append(rs, r)
		}
		matchers = append(matchers, &cmdlineMatcher{
			regexes:  rs,
			captures: make(map[string]string),
		})
	}
	if len(matchers) == 0 {
		return nil, fmt.Errorf("no matchers provided")
	}

	if nametmpl == "" {
		nametmpl = "{{.ExeBase}}"
	}
	tmpl := template.New("cmdname")
	tmpl, err := tmpl.Parse(nametmpl)
	if err != nil {
		return nil, fmt.Errorf("bad name template %q: %v", nametmpl, err)
	}

	return &matchNamer{matchers, templateNamer{tmpl}}, nil
}
