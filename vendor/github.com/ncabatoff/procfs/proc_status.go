// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package procfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// ProcStatus provides status information about the process,
// read from /proc/[pid]/status.
type (
	ProcStatus struct {
		TID                      int
		TracerPid                int
		UIDReal                  int
		UIDEffective             int
		UIDSavedSet              int
		UIDFileSystem            int
		GIDReal                  int
		GIDEffective             int
		GIDSavedSet              int
		GIDFileSystem            int
		FDSize                   int
		VmPeakKB                 int
		VmSizeKB                 int
		VmLckKB                  int
		VmHWMKB                  int
		VmRSSKB                  int
		VmDataKB                 int
		VmStkKB                  int
		VmExeKB                  int
		VmLibKB                  int
		VmPTEKB                  int
		VmSwapKB                 int
		VoluntaryCtxtSwitches    int
		NonvoluntaryCtxtSwitches int
	}

	procStatusFiller  func(*ProcStatus, string) error
	procStatusBuilder struct {
		scanners map[string]procStatusFiller
	}
)

func (ps *ProcStatus) refTID() []interface{} {
	return []interface{}{&ps.TID}
}
func (ps *ProcStatus) refTracerPid() []interface{} {
	return []interface{}{&ps.TracerPid}
}
func (ps *ProcStatus) refUID() []interface{} {
	return []interface{}{&ps.UIDReal, &ps.UIDEffective, &ps.UIDSavedSet, &ps.UIDFileSystem}
}
func (ps *ProcStatus) refGID() []interface{} {
	return []interface{}{&ps.GIDReal, &ps.GIDEffective, &ps.GIDSavedSet, &ps.GIDFileSystem}
}
func (ps *ProcStatus) refFDSize() []interface{} {
	return []interface{}{&ps.FDSize}
}
func (ps *ProcStatus) refVmPeakKB() []interface{} {
	return []interface{}{&ps.VmPeakKB}
}
func (ps *ProcStatus) refVmSizeKB() []interface{} {
	return []interface{}{&ps.VmSizeKB}
}
func (ps *ProcStatus) refVmLckKB() []interface{} {
	return []interface{}{&ps.VmLckKB}
}
func (ps *ProcStatus) refVmHWMKB() []interface{} {
	return []interface{}{&ps.VmHWMKB}
}
func (ps *ProcStatus) refVmRSSKB() []interface{} {
	return []interface{}{&ps.VmRSSKB}
}
func (ps *ProcStatus) refVmDataKB() []interface{} {
	return []interface{}{&ps.VmDataKB}
}
func (ps *ProcStatus) refVmStkKB() []interface{} {
	return []interface{}{&ps.VmStkKB}
}
func (ps *ProcStatus) refVmExeKB() []interface{} {
	return []interface{}{&ps.VmExeKB}
}
func (ps *ProcStatus) refVmLibKB() []interface{} {
	return []interface{}{&ps.VmLibKB}
}
func (ps *ProcStatus) refVmPTEKB() []interface{} {
	return []interface{}{&ps.VmPTEKB}
}
func (ps *ProcStatus) refVmSwapKB() []interface{} {
	return []interface{}{&ps.VmSwapKB}
}
func (ps *ProcStatus) refVoluntaryCtxtSwitches() []interface{} {
	return []interface{}{&ps.VoluntaryCtxtSwitches}
}
func (ps *ProcStatus) refNonvoluntaryCtxtSwitches() []interface{} {
	return []interface{}{&ps.NonvoluntaryCtxtSwitches}
}

func newFiller(format string, ref func(ps *ProcStatus) []interface{}) procStatusFiller {
	return procStatusFiller(func(ps *ProcStatus, s string) error {
		_, err := fmt.Sscanf(s, format, ref(ps)...)
		return err
	})
}

func newProcStatusBuilder() *procStatusBuilder {
	return &procStatusBuilder{
		scanners: map[string]procStatusFiller{
			"Pid":                        newFiller("%d", (*ProcStatus).refTID),
			"TracerPid":                  newFiller("%d", (*ProcStatus).refTracerPid),
			"Uid":                        newFiller("%d %d %d %d", (*ProcStatus).refUID),
			"Gid":                        newFiller("%d %d %d %d", (*ProcStatus).refGID),
			"FDSize":                     newFiller("%d", (*ProcStatus).refFDSize),
			"VmPeak":                     newFiller("%d kB", (*ProcStatus).refVmPeakKB),
			"VmSize":                     newFiller("%d kB", (*ProcStatus).refVmSizeKB),
			"VmLck":                      newFiller("%d kB", (*ProcStatus).refVmLckKB),
			"VmHWM":                      newFiller("%d kB", (*ProcStatus).refVmHWMKB),
			"VmRSS":                      newFiller("%d kB", (*ProcStatus).refVmRSSKB),
			"VmData":                     newFiller("%d kB", (*ProcStatus).refVmDataKB),
			"VmStk":                      newFiller("%d kB", (*ProcStatus).refVmStkKB),
			"VmExe":                      newFiller("%d kB", (*ProcStatus).refVmExeKB),
			"VmLib":                      newFiller("%d kB", (*ProcStatus).refVmLibKB),
			"VmPTE":                      newFiller("%d kB", (*ProcStatus).refVmPTEKB),
			"VmSwap":                     newFiller("%d kB", (*ProcStatus).refVmSwapKB),
			"voluntary_ctxt_switches":    newFiller("%d", (*ProcStatus).refVoluntaryCtxtSwitches),
			"nonvoluntary_ctxt_switches": newFiller("%d", (*ProcStatus).refNonvoluntaryCtxtSwitches),
		},
	}
}

func (b *procStatusBuilder) readStatus(r io.Reader) (ProcStatus, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return ProcStatus{}, err
	}
	s := string(contents)

	var ps ProcStatus
	for lineno := 0; s != ""; lineno++ {
		crpos := strings.IndexByte(s, '\n')
		if crpos == -1 {
			return ProcStatus{}, fmt.Errorf("line %d from status file without newline: %s", lineno, s)
		}
		line := strings.TrimSpace(s[:crpos])
		s = s[crpos+1:]
		if line == "" {
			if s == "" {
				break
			}
			continue
		}

		pos := strings.IndexByte(line, ':')
		if pos == -1 {
			return ProcStatus{}, fmt.Errorf("line %d from status file without ':': %s", lineno, line)
		}

		field := line[:pos]
		scanner, ok := b.scanners[field]
		if !ok {
			continue
		}

		err = scanner(&ps, line[pos+1:])
		// TODO: flag parse errors with some kind of "warning" error.
		if err != nil {
			// Be lenient about parse errors, because otherwise we miss out on some interesting
			// procs.  For example, my Ubuntu kernel (4.4.0-130-generic #156-Ubuntu) is showing
			// a Chromium status file with "VmLib:	18446744073709442944 kB".
			continue
		}
	}
	return ps, nil
}

var psb = newProcStatusBuilder()

// NewStatus returns the current status information of the process.
func (p Proc) NewStatus() (ProcStatus, error) {
	f, err := os.Open(p.path("status"))
	if err != nil {
		return ProcStatus{}, err
	}
	defer f.Close()

	return psb.readStatus(f)
}
