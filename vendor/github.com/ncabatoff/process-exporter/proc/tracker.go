package proc

import (
	"fmt"
	"os"
	"time"
)

type (
	Counts struct {
		Cpu        float64
		ReadBytes  uint64
		WriteBytes uint64
	}

	Memory struct {
		Resident uint64
		Virtual  uint64
	}

	Filedesc struct {
		Open  uint64
		Limit uint64
	}

	// Tracker tracks processes and records metrics.
	Tracker struct {
		// Tracked holds the processes are being monitored.  Processes
		// may be blacklisted such that they no longer get tracked by
		// setting their value in the Tracked map to nil.
		Tracked map[ProcId]*TrackedProc
		// ProcIds is a map from pid to ProcId.  This is a convenience
		// to allow finding the Tracked entry of a parent process.
		ProcIds map[int]ProcId
	}

	// TrackedProc accumulates metrics for a process, as well as
	// remembering an optional GroupName tag associated with it.
	TrackedProc struct {
		// lastUpdate is used internally during the update cycle to find which procs have exited
		lastUpdate time.Time
		// info is the most recently obtained info for this proc
		info ProcInfo
		// accum is the total CPU and IO accrued since we started tracking this proc
		accum Counts
		// lastaccum is the CPU and IO accrued in the last Update()
		lastaccum Counts
		// GroupName is an optional tag for this proc.
		GroupName string
	}

	trackedStats struct {
		aggregate, latest Counts
		Memory
		Filedesc
		start time.Time
	}
)

func (tp *TrackedProc) GetName() string {
	return tp.info.Name
}

func (tp *TrackedProc) GetCmdLine() []string {
	return tp.info.Cmdline
}

func (tp *TrackedProc) GetStats() trackedStats {
	mem := Memory{Resident: tp.info.ResidentBytes, Virtual: tp.info.VirtualBytes}
	fd := Filedesc{Open: tp.info.OpenFDs, Limit: tp.info.MaxFDs}
	return trackedStats{
		aggregate: tp.accum,
		latest:    tp.lastaccum,
		Memory:    mem,
		Filedesc:  fd,
		start:     tp.info.StartTime,
	}
}

func NewTracker() *Tracker {
	return &Tracker{Tracked: make(map[ProcId]*TrackedProc), ProcIds: make(map[int]ProcId)}
}

func (t *Tracker) Track(groupName string, idinfo ProcIdInfo) {
	info := ProcInfo{idinfo.ProcStatic, idinfo.ProcMetrics}
	t.Tracked[idinfo.ProcId] = &TrackedProc{GroupName: groupName, info: info}
}

func (t *Tracker) Ignore(id ProcId) {
	t.Tracked[id] = nil
}

// Scan procs and update metrics for those which are tracked.  Processes that have gone
// away get removed from the Tracked map.  New processes are returned, along with the count
// of permission errors.
func (t *Tracker) Update(procs ProcIter) ([]ProcIdInfo, int, error) {
	now := time.Now()
	var newProcs []ProcIdInfo
	var permissionErrors int

	for procs.Next() {
		procId, err := procs.GetProcId()
		if err != nil {
			continue
		}

		last, known := t.Tracked[procId]

		// Are we ignoring this proc?
		if known && last == nil {
			continue
		}

		// TODO if just the io file is unreadable, should we still return the other metrics?
		metrics, err := procs.GetMetrics()
		if err != nil {
			if os.IsPermission(err) {
				permissionErrors++
				t.Ignore(procId)
			}
			continue
		}

		if known {
			var newaccum, lastaccum Counts
			dcpu := metrics.CpuTime - last.info.CpuTime
			drbytes := metrics.ReadBytes - last.info.ReadBytes
			dwbytes := metrics.WriteBytes - last.info.WriteBytes

			lastaccum = Counts{Cpu: dcpu, ReadBytes: drbytes, WriteBytes: dwbytes}
			newaccum = Counts{
				Cpu:        last.accum.Cpu + lastaccum.Cpu,
				ReadBytes:  last.accum.ReadBytes + lastaccum.ReadBytes,
				WriteBytes: last.accum.WriteBytes + lastaccum.WriteBytes,
			}

			last.info.ProcMetrics = metrics
			last.lastUpdate = now
			last.accum = newaccum
			last.lastaccum = lastaccum
		} else {
			static, err := procs.GetStatic()
			if err != nil {
				continue
			}
			newProcs = append(newProcs, ProcIdInfo{procId, static, metrics})

			// Is this a new process with the same pid as one we already know?
			if oldProcId, ok := t.ProcIds[procId.Pid]; ok {
				// Delete it from known, otherwise the cleanup below will remove the
				// ProcIds entry we're about to create
				delete(t.Tracked, oldProcId)
			}
			t.ProcIds[procId.Pid] = procId
		}

	}
	err := procs.Close()
	if err != nil {
		return nil, permissionErrors, fmt.Errorf("Error reading procs: %v", err)
	}

	// Rather than allocating a new map each time to detect procs that have
	// disappeared, we bump the last update time on those that are still
	// present.  Then as a second pass we traverse the map looking for
	// stale procs and removing them.
	for procId, pinfo := range t.Tracked {
		if pinfo == nil {
			// TODO is this a bug? we're not tracking the proc so we don't see it go away so ProcIds
			// and Tracked are leaking?
			continue
		}
		if pinfo.lastUpdate != now {
			delete(t.Tracked, procId)
			delete(t.ProcIds, procId.Pid)
		}
	}

	return newProcs, permissionErrors, nil
}
