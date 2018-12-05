package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ncabatoff/procfs"
)

// ErrProcNotExist indicates a process couldn't be read because it doesn't exist,
// typically because it disappeared while we were reading it.
var ErrProcNotExist = fmt.Errorf("process does not exist")

type (
	// ID uniquely identifies a process.
	ID struct {
		// UNIX process id
		Pid int
		// The time the process started after system boot, the value is expressed
		// in clock ticks.
		StartTimeRel uint64
	}

	ThreadID ID

	// Static contains data read from /proc/pid/*
	Static struct {
		Name         string
		Cmdline      []string
		ParentPid    int
		StartTime    time.Time
		EffectiveUID int
	}

	// Counts are metric counters common to threads and processes and groups.
	Counts struct {
		CPUUserTime           float64
		CPUSystemTime         float64
		ReadBytes             uint64
		WriteBytes            uint64
		MajorPageFaults       uint64
		MinorPageFaults       uint64
		CtxSwitchVoluntary    uint64
		CtxSwitchNonvoluntary uint64
	}

	// Memory describes a proc's memory usage.
	Memory struct {
		ResidentBytes uint64
		VirtualBytes  uint64
		VmSwapBytes   uint64
	}

	// Filedesc describes a proc's file descriptor usage and soft limit.
	Filedesc struct {
		// Open is the count of open file descriptors, -1 if unknown.
		Open int64
		// Limit is the fd soft limit for the process.
		Limit uint64
	}

	// States counts how many threads are in each state.
	States struct {
		Running  int
		Sleeping int
		Waiting  int
		Zombie   int
		Other    int
	}

	// Metrics contains data read from /proc/pid/*
	Metrics struct {
		Counts
		Memory
		Filedesc
		NumThreads uint64
		States
		Wchan string
	}

	// Thread contains per-thread data.
	Thread struct {
		ThreadID
		ThreadName string
		Counts
		Wchan string
		States
	}

	// IDInfo groups all info for a single process.
	IDInfo struct {
		ID
		Static
		Metrics
		Threads []Thread
	}

	// ProcIdInfoThreads struct {
	// 	ProcIdInfo
	// 	Threads []ProcThread
	// }

	// Proc wraps the details of the underlying procfs-reading library.
	// Any of these methods may fail if the process has disapeared.
	// We try to return as much as possible rather than an error, e.g.
	// if some /proc files are unreadable.
	Proc interface {
		// GetPid() returns the POSIX PID (process id).  They may be reused over time.
		GetPid() int
		// GetProcID() returns (pid,starttime), which can be considered a unique process id.
		GetProcID() (ID, error)
		// GetStatic() returns various details read from files under /proc/<pid>/.  Technically
		// name may not be static, but we'll pretend it is.
		GetStatic() (Static, error)
		// GetMetrics() returns various metrics read from files under /proc/<pid>/.
		// It returns an error on complete failure.  Otherwise, it returns metrics
		// and 0 on complete success, 1 if some (like I/O) couldn't be read.
		GetMetrics() (Metrics, int, error)
		GetStates() (States, error)
		GetWchan() (string, error)
		GetCounts() (Counts, int, error)
		GetThreads() ([]Thread, error)
	}

	// proccache implements the Proc interface by acting as wrapper for procfs.Proc
	// that caches results of some reads.
	proccache struct {
		procfs.Proc
		procid  *ID
		stat    *procfs.ProcStat
		status  *procfs.ProcStatus
		cmdline []string
		io      *procfs.ProcIO
		fs      *FS
		wchan   *string
	}

	proc struct {
		proccache
	}

	// procs is a fancier []Proc that saves on some copying.
	procs interface {
		get(int) Proc
		length() int
	}

	// procfsprocs implements procs using procfs.
	procfsprocs struct {
		Procs []procfs.Proc
		fs    *FS
	}

	// Iter is an iterator over a sequence of procs.
	Iter interface {
		// Next returns true if the iterator is not exhausted.
		Next() bool
		// Close releases any resources the iterator uses.
		Close() error
		// The iterator satisfies the Proc interface.
		Proc
	}

	// procIterator implements the Iter interface
	procIterator struct {
		// procs is the list of Proc we're iterating over.
		procs
		// idx is the current iteration, i.e. it's an index into procs.
		idx int
		// err is set with an error when Next() fails.  It is not affected by failures accessing
		// the current iteration variable, e.g. with GetProcId.
		err error
		// Proc is the current iteration variable, or nil if Next() has never been called or the
		// iterator is exhausted.
		Proc
	}

	// Source is a source of procs.
	Source interface {
		// AllProcs returns all the processes in this source at this moment in time.
		AllProcs() Iter
	}

	// FS implements Source.
	FS struct {
		procfs.FS
		BootTime   uint64
		MountPoint string
		debug      bool
	}
)

func (ii IDInfo) String() string {
	return fmt.Sprintf("%+v:%+v", ii.ID, ii.Static)
}

// Add adds c2 to the counts.
func (c *Counts) Add(c2 Delta) {
	c.CPUUserTime += c2.CPUUserTime
	c.CPUSystemTime += c2.CPUSystemTime
	c.ReadBytes += c2.ReadBytes
	c.WriteBytes += c2.WriteBytes
	c.MajorPageFaults += c2.MajorPageFaults
	c.MinorPageFaults += c2.MinorPageFaults
	c.CtxSwitchVoluntary += c2.CtxSwitchVoluntary
	c.CtxSwitchNonvoluntary += c2.CtxSwitchNonvoluntary
}

// Sub subtracts c2 from the counts.
func (c Counts) Sub(c2 Counts) Delta {
	c.CPUUserTime -= c2.CPUUserTime
	c.CPUSystemTime -= c2.CPUSystemTime
	c.ReadBytes -= c2.ReadBytes
	c.WriteBytes -= c2.WriteBytes
	c.MajorPageFaults -= c2.MajorPageFaults
	c.MinorPageFaults -= c2.MinorPageFaults
	c.CtxSwitchVoluntary -= c2.CtxSwitchVoluntary
	c.CtxSwitchNonvoluntary -= c2.CtxSwitchNonvoluntary
	return Delta(c)
}

func (s *States) Add(s2 States) {
	s.Other += s2.Other
	s.Running += s2.Running
	s.Sleeping += s2.Sleeping
	s.Waiting += s2.Waiting
	s.Zombie += s2.Zombie
}

func (p IDInfo) GetThreads() ([]Thread, error) {
	return p.Threads, nil
}

// GetPid implements Proc.
func (p IDInfo) GetPid() int {
	return p.ID.Pid
}

// GetProcID implements Proc.
func (p IDInfo) GetProcID() (ID, error) {
	return p.ID, nil
}

// GetStatic implements Proc.
func (p IDInfo) GetStatic() (Static, error) {
	return p.Static, nil
}

// GetCounts implements Proc.
func (p IDInfo) GetCounts() (Counts, int, error) {
	return p.Metrics.Counts, 0, nil
}

// GetMetrics implements Proc.
func (p IDInfo) GetMetrics() (Metrics, int, error) {
	return p.Metrics, 0, nil
}

// GetStates implements Proc.
func (p IDInfo) GetStates() (States, error) {
	return p.States, nil
}

func (p IDInfo) GetWchan() (string, error) {
	return p.Wchan, nil
}

func (p *proccache) GetPid() int {
	return p.Proc.PID
}

func (p *proccache) getStat() (procfs.ProcStat, error) {
	if p.stat == nil {
		stat, err := p.Proc.NewStat()
		if err != nil {
			return procfs.ProcStat{}, err
		}
		p.stat = &stat
	}

	return *p.stat, nil
}

func (p *proccache) getStatus() (procfs.ProcStatus, error) {
	if p.status == nil {
		status, err := p.Proc.NewStatus()
		if err != nil {
			return procfs.ProcStatus{}, err
		}
		p.status = &status
	}

	return *p.status, nil
}

// GetProcID implements Proc.
func (p *proccache) GetProcID() (ID, error) {
	if p.procid == nil {
		stat, err := p.getStat()
		if err != nil {
			return ID{}, err
		}
		p.procid = &ID{Pid: p.GetPid(), StartTimeRel: stat.Starttime}
	}

	return *p.procid, nil
}

func (p *proccache) getCmdLine() ([]string, error) {
	if p.cmdline == nil {
		cmdline, err := p.Proc.CmdLine()
		if err != nil {
			return nil, err
		}
		p.cmdline = cmdline
	}
	return p.cmdline, nil
}

func (p *proccache) getWchan() (string, error) {
	if p.wchan == nil {
		wchan, err := p.Proc.Wchan()
		if err != nil {
			return "", err
		}
		p.wchan = &wchan
	}
	return *p.wchan, nil
}

func (p *proccache) getIo() (procfs.ProcIO, error) {
	if p.io == nil {
		io, err := p.Proc.NewIO()
		if err != nil {
			return procfs.ProcIO{}, err
		}
		p.io = &io
	}
	return *p.io, nil
}

// GetStatic returns the ProcStatic corresponding to this proc.
func (p *proccache) GetStatic() (Static, error) {
	// /proc/<pid>/cmdline is normally world-readable.
	cmdline, err := p.getCmdLine()
	if err != nil {
		return Static{}, err
	}

	// /proc/<pid>/stat is normally world-readable.
	stat, err := p.getStat()
	if err != nil {
		return Static{}, err
	}
	startTime := time.Unix(int64(p.fs.BootTime), 0).UTC()
	startTime = startTime.Add(time.Second / userHZ * time.Duration(stat.Starttime))

	// /proc/<pid>/status is normally world-readable.
	status, err := p.getStatus()
	if err != nil {
		return Static{}, err
	}

	return Static{
		Name:         stat.Comm,
		Cmdline:      cmdline,
		ParentPid:    stat.PPID,
		StartTime:    startTime,
		EffectiveUID: status.UIDEffective,
	}, nil
}

func (p proc) GetCounts() (Counts, int, error) {
	stat, err := p.getStat()
	if err != nil {
		if err == os.ErrNotExist {
			err = ErrProcNotExist
		}
		return Counts{}, 0, err
	}

	status, err := p.getStatus()
	if err != nil {
		if err == os.ErrNotExist {
			err = ErrProcNotExist
		}
		return Counts{}, 0, err
	}

	io, err := p.getIo()
	softerrors := 0
	if err != nil {
		softerrors++
	}
	return Counts{
		CPUUserTime:           float64(stat.UTime) / userHZ,
		CPUSystemTime:         float64(stat.STime) / userHZ,
		ReadBytes:             io.ReadBytes,
		WriteBytes:            io.WriteBytes,
		MajorPageFaults:       uint64(stat.MajFlt),
		MinorPageFaults:       uint64(stat.MinFlt),
		CtxSwitchVoluntary:    uint64(status.VoluntaryCtxtSwitches),
		CtxSwitchNonvoluntary: uint64(status.NonvoluntaryCtxtSwitches),
	}, softerrors, nil
}

func (p proc) GetWchan() (string, error) {
	return p.getWchan()
}

func (p proc) GetStates() (States, error) {
	stat, err := p.getStat()
	if err != nil {
		return States{}, err
	}

	var s States
	switch stat.State {
	case "R":
		s.Running++
	case "S":
		s.Sleeping++
	case "D":
		s.Waiting++
	case "Z":
		s.Zombie++
	default:
		s.Other++
	}
	return s, nil
}

// GetMetrics returns the current metrics for the proc.  The results are
// not cached.
func (p proc) GetMetrics() (Metrics, int, error) {
	counts, softerrors, err := p.GetCounts()
	if err != nil {
		return Metrics{}, 0, err
	}

	// We don't need to check for error here because p will have cached
	// the successful result of calling getStat in GetCounts.
	// Since GetMetrics isn't a pointer receiver method, our callers
	// won't see the effect of the caching between calls.
	stat, _ := p.getStat()

	// Ditto for states
	states, _ := p.GetStates()

	status, err := p.getStatus()
	if err != nil {
		return Metrics{}, 0, err
	}

	numfds, err := p.Proc.FileDescriptorsLen()
	if err != nil {
		numfds = -1
		softerrors |= 1
	}

	limits, err := p.Proc.NewLimits()
	if err != nil {
		return Metrics{}, 0, err
	}

	wchan, err := p.getWchan()
	if err != nil {
		softerrors |= 1
	}

	return Metrics{
		Counts: counts,
		Memory: Memory{
			ResidentBytes: uint64(stat.ResidentMemory()),
			VirtualBytes:  uint64(stat.VirtualMemory()),
			VmSwapBytes:   uint64(status.VmSwapKB * 1024),
		},
		Filedesc: Filedesc{
			Open:  int64(numfds),
			Limit: uint64(limits.OpenFiles),
		},
		NumThreads: uint64(stat.NumThreads),
		States:     states,
		Wchan:      wchan,
	}, softerrors, nil
}

func (p proc) GetThreads() ([]Thread, error) {
	fs, err := p.fs.threadFs(p.PID)
	if err != nil {
		return nil, err
	}

	threads := []Thread{}
	iter := fs.AllProcs()
	for iter.Next() {
		var id ID
		id, err = iter.GetProcID()
		if err != nil {
			continue
		}

		var static Static
		static, err = iter.GetStatic()
		if err != nil {
			continue
		}

		var counts Counts
		counts, _, err = iter.GetCounts()
		if err != nil {
			continue
		}

		wchan, _ := iter.GetWchan()
		states, _ := iter.GetStates()

		threads = append(threads, Thread{
			ThreadID:   ThreadID(id),
			ThreadName: static.Name,
			Counts:     counts,
			Wchan:      wchan,
			States:     states,
		})
	}
	err = iter.Close()
	if err != nil {
		return nil, err
	}
	if len(threads) < 2 {
		return nil, nil
	}

	return threads, nil
}

// See https://github.com/prometheus/procfs/blob/master/proc_stat.go for details on userHZ.
const userHZ = 100

// NewFS returns a new FS mounted under the given mountPoint. It will error
// if the mount point can't be read.
func NewFS(mountPoint string, debug bool) (*FS, error) {
	fs, err := procfs.NewFS(mountPoint)
	if err != nil {
		return nil, err
	}
	stat, err := fs.NewStat()
	if err != nil {
		return nil, err
	}
	return &FS{fs, stat.BootTime, mountPoint, debug}, nil
}

func (fs *FS) threadFs(pid int) (*FS, error) {
	mountPoint := filepath.Join(fs.MountPoint, strconv.Itoa(pid), "task")
	tfs, err := procfs.NewFS(mountPoint)
	if err != nil {
		return nil, err
	}
	return &FS{tfs, fs.BootTime, mountPoint, false}, nil
}

// AllProcs implements Source.
func (fs *FS) AllProcs() Iter {
	procs, err := fs.FS.AllProcs()
	if err != nil {
		err = fmt.Errorf("Error reading procs: %v", err)
	}
	return &procIterator{procs: procfsprocs{procs, fs}, err: err, idx: -1}
}

// get implements procs.
func (p procfsprocs) get(i int) Proc {
	return &proc{proccache{Proc: p.Procs[i], fs: p.fs}}
}

// length implements procs.
func (p procfsprocs) length() int {
	return len(p.Procs)
}

// Next implements Iter.
func (pi *procIterator) Next() bool {
	pi.idx++
	if pi.idx < pi.procs.length() {
		pi.Proc = pi.procs.get(pi.idx)
	} else {
		pi.Proc = nil
	}
	return pi.idx < pi.procs.length()
}

// Close implements Iter.
func (pi *procIterator) Close() error {
	pi.Next()
	pi.procs = nil
	pi.Proc = nil
	return pi.err
}
