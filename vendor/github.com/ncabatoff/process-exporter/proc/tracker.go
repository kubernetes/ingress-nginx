package proc

import (
	"fmt"
	"log"
	"os/user"
	"strconv"
	"time"

	seq "github.com/ncabatoff/go-seq/seq"
	common "github.com/ncabatoff/process-exporter"
)

type (
	// Tracker tracks processes and records metrics.
	Tracker struct {
		// namer determines what processes to track and names them
		namer common.MatchNamer
		// tracked holds the processes are being monitored.  Processes
		// may be blacklisted such that they no longer get tracked by
		// setting their value in the tracked map to nil.
		tracked map[ID]*trackedProc
		// procIds is a map from pid to ProcId.  This is a convenience
		// to allow finding the Tracked entry of a parent process.
		procIds map[int]ID
		// trackChildren makes Tracker track descendants of procs the
		// namer wanted tracked.
		trackChildren bool
		// never ignore processes, i.e. always re-check untracked processes in case comm has changed
		alwaysRecheck bool
		username      map[int]string
		debug         bool
	}

	// Delta is an alias of Counts used to signal that its contents are not
	// totals, but rather the result of subtracting two totals.
	Delta Counts

	trackedThread struct {
		name       string
		accum      Counts
		latest     Delta
		lastUpdate time.Time
		wchan      string
	}

	// trackedProc accumulates metrics for a process, as well as
	// remembering an optional GroupName tag associated with it.
	trackedProc struct {
		// lastUpdate is used internally during the update cycle to find which procs have exited
		lastUpdate time.Time
		// static
		static  Static
		metrics Metrics
		// lastaccum is the increment to the counters seen in the last update.
		lastaccum Delta
		// groupName is the tag for this proc given by the namer.
		groupName string
		threads   map[ThreadID]trackedThread
	}

	// ThreadUpdate describes what's changed for a thread since the last cycle.
	ThreadUpdate struct {
		// ThreadName is the name of the thread based on field of stat.
		ThreadName string
		// Latest is how much the counts increased since last cycle.
		Latest Delta
	}

	// Update reports on the latest stats for a process.
	Update struct {
		// GroupName is the name given by the namer to the process.
		GroupName string
		// Latest is how much the counts increased since last cycle.
		Latest Delta
		// Memory is the current memory usage.
		Memory
		// Filedesc is the current fd usage/limit.
		Filedesc
		// Start is the time the process started.
		Start time.Time
		// NumThreads is the number of threads.
		NumThreads uint64
		// States is how many processes are in which run state.
		States
		// Wchans is how many threads are in each non-zero wchan.
		Wchans map[string]int
		// Threads are the thread updates for this process.
		Threads []ThreadUpdate
	}

	// CollectErrors describes non-fatal errors found while collecting proc
	// metrics.
	CollectErrors struct {
		// Read is incremented every time GetMetrics() returns an error.
		// This means we failed to load even the basics for the process,
		// and not just because it disappeared on us.
		Read int
		// Partial is incremented every time we're unable to collect
		// some metrics (e.g. I/O) for a tracked proc, but we're still able
		// to get the basic stuff like cmdline and core stats.
		Partial int
	}
)

func lessUpdateGroupName(x, y Update) bool { return x.GroupName < y.GroupName }

func lessThreadUpdate(x, y ThreadUpdate) bool { return seq.Compare(x, y) < 0 }

func lessCounts(x, y Counts) bool { return seq.Compare(x, y) < 0 }

func (tp *trackedProc) getUpdate() Update {
	u := Update{
		GroupName:  tp.groupName,
		Latest:     tp.lastaccum,
		Memory:     tp.metrics.Memory,
		Filedesc:   tp.metrics.Filedesc,
		Start:      tp.static.StartTime,
		NumThreads: tp.metrics.NumThreads,
		States:     tp.metrics.States,
		Wchans:     make(map[string]int),
	}
	if tp.metrics.Wchan != "" {
		u.Wchans[tp.metrics.Wchan] = 1
	}
	if len(tp.threads) > 1 {
		for _, tt := range tp.threads {
			u.Threads = append(u.Threads, ThreadUpdate{tt.name, tt.latest})
			if tt.wchan != "" {
				u.Wchans[tt.wchan]++
			}
		}
	}
	return u
}

// NewTracker creates a Tracker.
func NewTracker(namer common.MatchNamer, trackChildren, alwaysRecheck, debug bool) *Tracker {
	return &Tracker{
		namer:         namer,
		tracked:       make(map[ID]*trackedProc),
		procIds:       make(map[int]ID),
		trackChildren: trackChildren,
		alwaysRecheck: alwaysRecheck,
		username:      make(map[int]string),
		debug:         debug,
	}
}

func (t *Tracker) track(groupName string, idinfo IDInfo) {
	tproc := trackedProc{
		groupName: groupName,
		static:    idinfo.Static,
		metrics:   idinfo.Metrics,
	}
	if len(idinfo.Threads) > 0 {
		tproc.threads = make(map[ThreadID]trackedThread)
		for _, thr := range idinfo.Threads {
			tproc.threads[thr.ThreadID] = trackedThread{
				thr.ThreadName, thr.Counts, Delta{}, time.Time{}, thr.Wchan}
		}
	}
	t.tracked[idinfo.ID] = &tproc
}

func (t *Tracker) ignore(id ID) {
	// only ignore ID if we didn't set recheck to true
	if t.alwaysRecheck == false {
		t.tracked[id] = nil
	}
}

func (tp *trackedProc) update(metrics Metrics, now time.Time, cerrs *CollectErrors, threads []Thread) {
	// newcounts: resource consumption since last cycle
	newcounts := metrics.Counts
	tp.lastaccum = newcounts.Sub(tp.metrics.Counts)
	tp.metrics = metrics
	tp.lastUpdate = now
	if len(threads) > 1 {
		if tp.threads == nil {
			tp.threads = make(map[ThreadID]trackedThread)
		}
		for _, thr := range threads {
			tt := trackedThread{thr.ThreadName, thr.Counts, Delta{}, now, thr.Wchan}
			if old, ok := tp.threads[thr.ThreadID]; ok {
				tt.latest, tt.accum = thr.Counts.Sub(old.accum), thr.Counts
			}
			tp.threads[thr.ThreadID] = tt
		}
		for id, tt := range tp.threads {
			if tt.lastUpdate != now {
				delete(tp.threads, id)
			}
		}
	} else {
		tp.threads = nil
	}
}

// handleProc updates the tracker if it's a known and not ignored proc.
// If it's neither known nor ignored, newProc will be non-nil.
// It is not an error if the process disappears while we are reading
// its info out of /proc, it just means nothing will be returned and
// the tracker will be unchanged.
func (t *Tracker) handleProc(proc Proc, updateTime time.Time) (*IDInfo, CollectErrors) {
	var cerrs CollectErrors
	procID, err := proc.GetProcID()
	if err != nil {
		return nil, cerrs
	}

	// Do nothing if we're ignoring this proc.
	last, known := t.tracked[procID]
	if known && last == nil {
		return nil, cerrs
	}

	metrics, softerrors, err := proc.GetMetrics()
	if err != nil {
		if t.debug {
			log.Printf("error reading metrics for %+v: %v", procID, err)
		}
		// This usually happens due to the proc having exited, i.e.
		// we lost the race.  We don't count that as an error.
		if err != ErrProcNotExist {
			cerrs.Read++
		}
		return nil, cerrs
	}

	var threads []Thread
	threads, err = proc.GetThreads()
	if err != nil {
		softerrors |= 1
	}
	cerrs.Partial += softerrors

	if len(threads) > 0 {
		metrics.Counts.CtxSwitchNonvoluntary, metrics.Counts.CtxSwitchVoluntary = 0, 0
		for _, thread := range threads {
			metrics.Counts.CtxSwitchNonvoluntary += thread.Counts.CtxSwitchNonvoluntary
			metrics.Counts.CtxSwitchVoluntary += thread.Counts.CtxSwitchVoluntary
			metrics.States.Add(thread.States)
		}
	}

	var newProc *IDInfo
	if known {
		last.update(metrics, updateTime, &cerrs, threads)
	} else {
		static, err := proc.GetStatic()
		if err != nil {
			if t.debug {
				log.Printf("error reading static details for %+v: %v", procID, err)
			}
			return nil, cerrs
		}
		newProc = &IDInfo{procID, static, metrics, threads}
		if t.debug {
			log.Printf("found new proc: %s", newProc)
		}

		// Is this a new process with the same pid as one we already know?
		// Then delete it from the known map, otherwise the cleanup in Update()
		// will remove the ProcIds entry we're creating here.
		if oldProcID, ok := t.procIds[procID.Pid]; ok {
			delete(t.tracked, oldProcID)
		}
		t.procIds[procID.Pid] = procID
	}
	return newProc, cerrs
}

// update scans procs and updates metrics for those which are tracked. Processes
// that have gone away get removed from the Tracked map. New processes are
// returned, along with the count of nonfatal errors.
func (t *Tracker) update(procs Iter) ([]IDInfo, CollectErrors, error) {
	var newProcs []IDInfo
	var colErrs CollectErrors
	var now = time.Now()

	for procs.Next() {
		newProc, cerrs := t.handleProc(procs, now)
		if newProc != nil {
			newProcs = append(newProcs, *newProc)
		}
		colErrs.Read += cerrs.Read
		colErrs.Partial += cerrs.Partial
	}

	err := procs.Close()
	if err != nil {
		return nil, colErrs, fmt.Errorf("Error reading procs: %v", err)
	}

	// Rather than allocating a new map each time to detect procs that have
	// disappeared, we bump the last update time on those that are still
	// present.  Then as a second pass we traverse the map looking for
	// stale procs and removing them.
	for procID, pinfo := range t.tracked {
		if pinfo == nil {
			// TODO is this a bug? we're not tracking the proc so we don't see it go away so ProcIds
			// and Tracked are leaking?
			continue
		}
		if pinfo.lastUpdate != now {
			delete(t.tracked, procID)
			delete(t.procIds, procID.Pid)
		}
	}

	return newProcs, colErrs, nil
}

// checkAncestry walks the process tree recursively towards the root,
// stopping at pid 1 or upon finding a parent that's already tracked
// or ignored.  If we find a tracked parent track this one too; if not,
// ignore this one.
func (t *Tracker) checkAncestry(idinfo IDInfo, newprocs map[ID]IDInfo) string {
	ppid := idinfo.ParentPid
	pProcID := t.procIds[ppid]
	if pProcID.Pid < 1 {
		if t.debug {
			log.Printf("ignoring unmatched proc with no matched parent: %+v", idinfo)
		}
		// Reached root of process tree without finding a tracked parent.
		t.ignore(idinfo.ID)
		return ""
	}

	// Is the parent already known to the tracker?
	if ptproc, ok := t.tracked[pProcID]; ok {
		if ptproc != nil {
			if t.debug {
				log.Printf("matched as %q because child of %+v: %+v",
					ptproc.groupName, pProcID, idinfo)
			}
			// We've found a tracked parent.
			t.track(ptproc.groupName, idinfo)
			return ptproc.groupName
		}
		// We've found an untracked parent.
		t.ignore(idinfo.ID)
		return ""
	}

	// Is the parent another new process?
	if pinfoid, ok := newprocs[pProcID]; ok {
		if name := t.checkAncestry(pinfoid, newprocs); name != "" {
			if t.debug {
				log.Printf("matched as %q because child of %+v: %+v",
					name, pProcID, idinfo)
			}
			// We've found a tracked parent, which implies this entire lineage should be tracked.
			t.track(name, idinfo)
			return name
		}
	}

	// Parent is dead, i.e. we never saw it, or there's no tracked proc in our ancestry.
	if t.debug {
		log.Printf("ignoring unmatched proc with no matched parent: %+v", idinfo)
	}
	t.ignore(idinfo.ID)
	return ""
}

func (t *Tracker) lookupUid(uid int) string {
	if name, ok := t.username[uid]; ok {
		return name
	}

	var name string
	uidstr := strconv.Itoa(uid)
	u, err := user.LookupId(uidstr)
	if err != nil {
		name = uidstr
	} else {
		name = u.Username
	}
	t.username[uid] = name
	return name
}

// Update modifies the tracker's internal state based on what it reads from
// iter.  Tracks any new procs the namer wants tracked, and updates
// its metrics for existing tracked procs.  Returns nonfatal errors
// and the status of all tracked procs, or an error if fatal.
func (t *Tracker) Update(iter Iter) (CollectErrors, []Update, error) {
	newProcs, colErrs, err := t.update(iter)
	if err != nil {
		return colErrs, nil, err
	}

	// Step 1: track any new proc that should be tracked based on its name and cmdline.
	untracked := make(map[ID]IDInfo)
	for _, idinfo := range newProcs {
		nacl := common.ProcAttributes{
			Name:     idinfo.Name,
			Cmdline:  idinfo.Cmdline,
			Username: t.lookupUid(idinfo.EffectiveUID),
		}
		wanted, gname := t.namer.MatchAndName(nacl)
		if wanted {
			if t.debug {
				log.Printf("matched as %q: %+v", gname, idinfo)
			}
			t.track(gname, idinfo)
		} else {
			untracked[idinfo.ID] = idinfo
		}
	}

	// Step 2: track any untracked new proc that should be tracked because its parent is tracked.
	if t.trackChildren {
		for _, idinfo := range untracked {
			if _, ok := t.tracked[idinfo.ID]; ok {
				// Already tracked or ignored in an earlier iteration
				continue
			}

			t.checkAncestry(idinfo, untracked)
		}
	}

	tp := []Update{}
	for _, tproc := range t.tracked {
		if tproc != nil {
			tp = append(tp, tproc.getUpdate())
		}
	}
	return colErrs, tp, nil
}
