package nginx

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"k8s.io/klog/v2"
)

const (
	defBinary = "/usr/bin/nginx"
	CfgPath   = "/etc/nginx/conf/nginx.conf"
	TempDir = "/etc/ingress-controller/tempconf"
)

// NginxExecTester defines the interface to execute
// command like reload or test configuration
type NginxExecutor interface {
	Reload() ([]byte, error)
	Test(cfg string) ([]byte, error)
	Stop() error
	Start(chan error) error
}

// NginxCommand stores context around a given nginx executable path
type NginxCommand struct {
	Binary string
}

// NewNginxCommand returns a new NginxCommand from which path
// has been detected from environment variable NGINX_BINARY or default
func NewNginxCommand() NginxCommand {
	command := NginxCommand{
		Binary: defBinary,
	}

	binary := os.Getenv("NGINX_BINARY")
	if binary != "" {
		command.Binary = binary
	}

	return command
}

// ExecCommand instanciates an exec.Cmd object to call nginx program
func (nc NginxCommand) execCommand(args ...string) *exec.Cmd {
	cmdArgs := []string{}

	cmdArgs = append(cmdArgs, "-c", CfgPath)
	cmdArgs = append(cmdArgs, args...)
	//nolint:gosec // Ignore G204 error
	executor := exec.Command(nc.Binary, cmdArgs...)
	executor.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	return executor
}

func (nc NginxCommand) Start(errch chan error) error {
	cmd := nc.execCommand()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		klog.ErrorS(err, "NGINX error")
		return err
	}
	go func() {
		errch <- cmd.Wait()
	}()
	return nil
}

func (nc NginxCommand) Reload() ([]byte, error) {
	cmd := nc.execCommand("-s", "reload")
	return cmd.CombinedOutput()
}

func (nc NginxCommand) Stop() error {
	cmd := nc.execCommand("-s", "quit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Test checks if config file is a syntax valid nginx configuration
func (nc NginxCommand) Test(cfg string) ([]byte, error) {
	//nolint:gosec // Ignore G204 error
	cfg = filepath.Join(TempDir, cfg)
	return exec.Command(nc.Binary, "-c", cfg, "-t").CombinedOutput()
}