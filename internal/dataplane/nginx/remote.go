package nginx

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// NginxCommand stores context around a given nginx executable path
type NginxRemote struct {
	host string
}

// NewNginxCommand returns a new NginxCommand from which path
// has been detected from environment variable NGINX_BINARY or default
func NewNginxRemote(host string) NginxExecutor {	
	return NginxRemote{
		host: host,
	}
}

func (nc NginxRemote) Start(errch chan error) error {
	/*getStart, err := url.JoinPath(nc.host, "start") // TODO: Turn this path a constant on dataplane
	if err != nil {
		return err
	}
	resp, err := http.Get(getStart)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error executing start: %s", string(body))
	}*/

	// TODO: Add a ping/watcher to backend and populate error channel
	return nil
}

func (nc NginxRemote) Reload() ([]byte, error) {
	getReload, err := url.JoinPath(nc.host, "reload") // TODO: Turn this path a constant on dataplane
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(getReload)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error executing reload: %s", string(body))
	}
	return body, nil
}

func (nc NginxRemote) Stop() error {
	getStop, err := url.JoinPath(nc.host, "stop") // TODO: Turn this path a constant on dataplane
	if err != nil {
		return err
	}
	resp, err := http.Get(getStop)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error executing stop: %s", string(body))
	}
	return nil
}

// Test checks if config file is a syntax valid nginx configuration
func (nc NginxRemote) Test(cfg string) ([]byte, error) {
	form := url.Values{}
	form.Add("testfile", cfg)
	getStop, err := url.JoinPath(nc.host, "test") // TODO: Turn this path a constant on dataplane
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(getStop, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("error executing stop: %s", string(body))
	}
	return body, nil
}