package proc

import (
	. "gopkg.in/check.v1"
	"os"
	"os/exec"
)

var fs *FS

func init() {
	fs, _ = NewFS("/proc")
}

// Basic test of proc reading: does AllProcs return at least two procs, one of which is us.
func (s MySuite) TestAllProcs(c *C) {
	procs := fs.AllProcs()
	count := 0
	for procs.Next() {
		count++
		if procs.GetPid() != os.Getpid() {
			continue
		}
		procid, err := procs.GetProcId()
		c.Assert(err, IsNil)
		c.Check(procid.Pid, Equals, os.Getpid())
		static, err := procs.GetStatic()
		c.Assert(err, IsNil)
		c.Check(static.ParentPid, Equals, os.Getppid())
	}
	err := procs.Close()
	c.Assert(err, IsNil)
	c.Check(count, Not(Equals), 0)
}

// Test that we can observe the absence of a child process before it spawns and after it exits,
// and its presence during its lifetime.
func (s MySuite) TestAllProcsSpawn(c *C) {
	childprocs := func() ([]ProcIdStatic, error) {
		found := []ProcIdStatic{}
		procs := fs.AllProcs()
		mypid := os.Getpid()
		for procs.Next() {
			procid, err := procs.GetProcId()
			if err != nil {
				continue
			}
			static, err := procs.GetStatic()
			if err != nil {
				continue
			}
			if static.ParentPid == mypid {
				found = append(found, ProcIdStatic{procid, static})
			}
		}
		err := procs.Close()
		if err != nil {
			return nil, err
		}
		return found, nil
	}

	children1, err := childprocs()
	c.Assert(err, IsNil)

	cmd := exec.Command("/bin/cat")
	wc, err := cmd.StdinPipe()
	c.Assert(err, IsNil)
	err = cmd.Start()
	c.Assert(err, IsNil)

	children2, err := childprocs()
	c.Assert(err, IsNil)

	err = wc.Close()
	c.Assert(err, IsNil)
	err = cmd.Wait()
	c.Assert(err, IsNil)

	children3, err := childprocs()
	c.Assert(err, IsNil)

	foundcat := func(procs []ProcIdStatic) bool {
		for _, proc := range procs {
			if proc.Name == "cat" {
				return true
			}
		}
		return false
	}

	c.Check(foundcat(children1), Equals, false)
	c.Check(foundcat(children2), Equals, true)
	c.Check(foundcat(children3), Equals, false)
}

func (s MySuite) TestIterator(c *C) {
	// create a new proc with zero metrics, cmdline, starttime, ppid
	newProc := func(pid int, name string) ProcIdInfo {
		pis := newProcIdStatic(pid, 0, 0, name, nil)
		return ProcIdInfo{
			ProcId:      pis.ProcId,
			ProcStatic:  pis.ProcStatic,
			ProcMetrics: ProcMetrics{},
		}
	}
	p1 := newProc(1, "p1")
	want1 := []ProcIdInfo{p1}
	pi1 := procInfoIter(want1...)
	got, err := consumeIter(pi1)
	c.Assert(err, IsNil)
	c.Check(got, DeepEquals, want1)

	p2 := newProc(2, "p2")
	want2 := []ProcIdInfo{p1, p2}
	pi2 := procInfoIter(want2...)
	got2, err := consumeIter(pi2)
	c.Assert(err, IsNil)
	c.Check(got2, DeepEquals, want2)
}
