package nginx

import (
	"fmt"
	"net/http"

	nginxdataplane "k8s.io/ingress-nginx/internal/dataplane/nginx"
	"k8s.io/klog/v2"
)

type nginxExecutor struct {
	cmd nginxdataplane.NginxExecutor
	errch chan error
	stopch chan bool
	stopdelay int
}

func NewNGINXExecutor(mux *http.ServeMux, stopdelay int, errch chan error, stopch chan bool) *nginxExecutor {
	n := &nginxExecutor{
		cmd: nginxdataplane.NewNginxCommand(),
		stopdelay: stopdelay,
		errch: errch,
		stopch: stopch,
	}
	registerDataplaneHandler(n, mux)
	return n
}

func (n *nginxExecutor) Start() {
	n.cmd.Start(n.errch)
}

func (n *nginxExecutor) Stop() error {
	return n.cmd.Stop()
}

func (n *nginxExecutor) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := n.Stop(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		n.errch <- fmt.Errorf("error stopping: %w", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	n.stopch <- true
}

func (n *nginxExecutor) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	o, err := n.cmd.Reload()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error() + "\n"))
		w.Write(o)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (n *nginxExecutor) handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusForbidden)
		klog.ErrorS(err, "error parsing request", "handler", "test")
		return
	}

	testFile := r.FormValue("testfile")
	if testFile == "" {
		w.WriteHeader(http.StatusForbidden)
		klog.ErrorS(fmt.Errorf("testfile parameter not found"), "error parsing request", "handler", "test")
		return
	}

	o, err := n.cmd.Test(testFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error() + "\n"))
		w.Write(o)
		klog.ErrorS(err, "error testing file", "output", string(o), "handler", "test")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func registerDataplaneHandler(n *nginxExecutor, mux *http.ServeMux) {
	mux.HandleFunc("/stop", n.handleStop)
	mux.HandleFunc("/reload", n.handleReload)
	mux.HandleFunc("/test", n.handleTest)
}