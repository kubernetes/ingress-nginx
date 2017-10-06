package proc

import (
	. "gopkg.in/check.v1"
)

// Verify that the tracker accurately reports new procs that aren't ignored or tracked.
func (s MySuite) TestTrackerBasic(c *C) {
	// create a new proc with zero metrics, cmdline, starttime, ppid
	newProc := func(pid int, startTime uint64, name string) ProcIdInfo {
		pis := newProcIdStatic(pid, 0, 0, name, nil)
		return ProcIdInfo{
			ProcId:      pis.ProcId,
			ProcStatic:  pis.ProcStatic,
			ProcMetrics: ProcMetrics{},
		}
	}
	tr := NewTracker()

	// Test that p1 is seen as new
	p1 := newProc(1, 1, "p1")
	want1 := []ProcIdInfo{p1}
	got1, _, err := tr.Update(procInfoIter(want1...))
	c.Assert(err, IsNil)
	c.Check(got1, DeepEquals, want1)

	// Test that p1 is no longer seen as new once tracked
	tr.Track("g1", p1)
	got2, _, err := tr.Update(procInfoIter(want1...))
	c.Assert(err, IsNil)
	c.Check(got2, DeepEquals, []ProcIdInfo(nil))

	// Test that p2 is new now, but p1 is still not
	p2 := newProc(2, 2, "p2")
	want2 := []ProcIdInfo{p2}
	got3, _, err := tr.Update(procInfoIter(p1, p2))
	c.Assert(err, IsNil)
	c.Check(got3, DeepEquals, want2)

	// Test that p2 stops being new once ignored
	tr.Ignore(p2.ProcId)
	got4, _, err := tr.Update(procInfoIter(p1, p2))
	c.Assert(err, IsNil)
	c.Check(got4, DeepEquals, []ProcIdInfo(nil))

	// TODO test that starttime is taken into account, i.e. pid recycling is handled.
}

// Verify that the tracker accurately reports metric changes.
func (s MySuite) TestTrackerCounts(c *C) {
	// create a new proc with cmdline, starttime, ppid
	newProc := func(pid int, startTime uint64, name string, m ProcMetrics) ProcIdInfo {
		pis := newProcIdStatic(pid, 0, 0, name, nil)
		return ProcIdInfo{
			ProcId:      pis.ProcId,
			ProcStatic:  pis.ProcStatic,
			ProcMetrics: m,
		}
	}
	tr := NewTracker()

	// Test that p1 is seen as new
	p1 := newProc(1, 1, "p1", ProcMetrics{1, 2, 3, 4, 5, 6, 4096})
	want1 := []ProcIdInfo{p1}
	got1, _, err := tr.Update(procInfoIter(p1))
	c.Assert(err, IsNil)
	c.Check(got1, DeepEquals, want1)

	// Test that p1 is no longer seen as new once tracked
	tr.Track("g1", p1)
	got2, _, err := tr.Update(procInfoIter(p1))
	c.Assert(err, IsNil)
	c.Check(got2, DeepEquals, []ProcIdInfo(nil))

	// Now update p1's metrics
	p1.ProcMetrics = ProcMetrics{2, 3, 4, 5, 6, 7, 4096}
	got3, _, err := tr.Update(procInfoIter(p1))
	c.Assert(err, IsNil)
	c.Check(got3, DeepEquals, []ProcIdInfo(nil))

	// Test that counts are correct
	c.Check(tr.Tracked[p1.ProcId].accum, Equals, Counts{1, 1, 1})
	c.Check(tr.Tracked[p1.ProcId].info, DeepEquals, ProcInfo{p1.ProcStatic, p1.ProcMetrics})

	// Now update p1's metrics again
	p1.ProcMetrics = ProcMetrics{4, 6, 8, 9, 10, 11, 4096}
	got4, _, err := tr.Update(procInfoIter(p1))
	c.Assert(err, IsNil)
	c.Check(got4, DeepEquals, []ProcIdInfo(nil))

	// Test that counts are correct
	c.Check(tr.Tracked[p1.ProcId].accum, Equals, Counts{3, 4, 5})
	c.Check(tr.Tracked[p1.ProcId].info, DeepEquals, ProcInfo{p1.ProcStatic, p1.ProcMetrics})
}
