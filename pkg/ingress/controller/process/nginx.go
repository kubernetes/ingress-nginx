package process

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/golang/glog"
	ps "github.com/mitchellh/go-ps"
	"github.com/ncabatoff/process-exporter/proc"
)

func IsRespawnIfRequired(err error) bool {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	waitStatus := exitError.Sys().(syscall.WaitStatus)
	glog.Warningf(`
-------------------------------------------------------------------------------
NGINX master process died (%v): %v
-------------------------------------------------------------------------------
`, waitStatus.ExitStatus(), err)
	return true
}

func WaitUntilPortIsAvailable(port int) {
	// we wait until the workers are killed
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("0.0.0.0:%v", port), 1*time.Second)
		if err != nil {
			break
		}
		conn.Close()
		// kill nginx worker processes
		fs, err := proc.NewFS("/proc")
		procs, _ := fs.FS.AllProcs()
		for _, p := range procs {
			pn, err := p.Comm()
			if err != nil {
				glog.Errorf("unexpected error obtaining process information: %v", err)
				continue
			}

			if pn == "nginx" {
				osp, err := os.FindProcess(p.PID)
				if err != nil {
					glog.Errorf("unexpected error obtaining process information: %v", err)
					continue
				}
				osp.Signal(syscall.SIGQUIT)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// IsNginxRunning returns true if a process with the name 'nginx' is found
func IsNginxRunning() bool {
	processes, _ := ps.Processes()
	for _, p := range processes {
		if p.Executable() == "nginx" {
			return true
		}
	}
	return false
}
