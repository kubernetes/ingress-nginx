package config

import (
	// "github.com/kylelemons/godebug/pretty"
	common "github.com/ncabatoff/process-exporter"
	. "gopkg.in/check.v1"
)

func (s MySuite) TestConfigBasic(c *C) {
	yml := `
process_names:
  - exe: 
    - bash
  - exe: 
    - sh
  - exe: 
    - /bin/ksh
`
	cfg, err := GetConfig(yml)
	c.Assert(err, IsNil)
	c.Check(cfg.MatchNamers, HasLen, 3)

	bash := common.NameAndCmdline{Name: "bash", Cmdline: []string{"/bin/bash"}}
	sh := common.NameAndCmdline{Name: "sh", Cmdline: []string{"sh"}}
	ksh := common.NameAndCmdline{Name: "ksh", Cmdline: []string{"/bin/ksh"}}

	found, name := cfg.MatchNamers[0].MatchAndName(bash)
	c.Check(found, Equals, true)
	c.Check(name, Equals, "bash")
	found, name = cfg.MatchNamers[0].MatchAndName(sh)
	c.Check(found, Equals, false)
	found, name = cfg.MatchNamers[0].MatchAndName(ksh)
	c.Check(found, Equals, false)

	found, name = cfg.MatchNamers[1].MatchAndName(bash)
	c.Check(found, Equals, false)
	found, name = cfg.MatchNamers[1].MatchAndName(sh)
	c.Check(found, Equals, true)
	c.Check(name, Equals, "sh")
	found, name = cfg.MatchNamers[1].MatchAndName(ksh)
	c.Check(found, Equals, false)

	found, name = cfg.MatchNamers[2].MatchAndName(bash)
	c.Check(found, Equals, false)
	found, name = cfg.MatchNamers[2].MatchAndName(sh)
	c.Check(found, Equals, false)
	found, name = cfg.MatchNamers[2].MatchAndName(ksh)
	c.Check(found, Equals, true)
	c.Check(name, Equals, "ksh")
}

func (s MySuite) TestConfigTemplates(c *C) {
	yml := `
process_names:
  - exe: 
    - postmaster
    cmdline: 
    - "-D\\s+.+?(?P<Path>[^/]+)(?:$|\\s)"
    name: "{{.ExeBase}}:{{.Matches.Path}}"
  - exe: 
    - prometheus
    name: "{{.ExeFull}}"
`
	cfg, err := GetConfig(yml)
	c.Assert(err, IsNil)
	c.Check(cfg.MatchNamers, HasLen, 2)

	postgres := common.NameAndCmdline{Name: "postmaster", Cmdline: []string{"/usr/bin/postmaster", "-D", "/data/pg"}}
	found, name := cfg.MatchNamers[0].MatchAndName(postgres)
	c.Check(found, Equals, true)
	c.Check(name, Equals, "postmaster:pg")

	pm := common.NameAndCmdline{Name: "prometheus", Cmdline: []string{"/usr/local/bin/prometheus"}}
	found, name = cfg.MatchNamers[1].MatchAndName(pm)
	c.Check(found, Equals, true)
	c.Check(name, Equals, "/usr/local/bin/prometheus")
}
