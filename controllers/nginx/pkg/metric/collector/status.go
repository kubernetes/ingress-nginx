/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/golang/glog"
)

var (
	ac      = regexp.MustCompile(`Active connections: (\d+)`)
	sahr    = regexp.MustCompile(`(\d+)\s(\d+)\s(\d+)`)
	reading = regexp.MustCompile(`Reading: (\d+)`)
	writing = regexp.MustCompile(`Writing: (\d+)`)
	waiting = regexp.MustCompile(`Waiting: (\d+)`)
)

type basicStatus struct {
	// Active total number of active connections
	Active int
	// Accepted total number of accepted client connections
	Accepted int
	// Handled total number of handled connections. Generally, the parameter value is the same as accepts unless some resource limits have been reached (for example, the worker_connections limit).
	Handled int
	// Requests total number of client requests.
	Requests int
	// Reading current number of connections where nginx is reading the request header.
	Reading int
	// Writing current number of connections where nginx is writing the response back to the client.
	Writing int
	// Waiting current number of idle client connections waiting for a request.
	Waiting int
}

// https://github.com/vozlt/nginx-module-vts
type vts struct {
	NginxVersion string `json:"nginxVersion"`
	LoadMsec     int    `json:"loadMsec"`
	NowMsec      int    `json:"nowMsec"`
	// Total connections and requests(same as stub_status_module in NGINX)
	Connections connections `json:"connections"`
	// Traffic(in/out) and request and response counts and cache hit ratio per each server zone
	ServerZones map[string]serverZone `json:"serverZones"`
	// Traffic(in/out) and request and response counts and cache hit ratio per each server zone filtered through
	// the vhost_traffic_status_filter_by_set_key directive
	FilterZones map[string]map[string]filterZone `json:"filterZones"`
	// Traffic(in/out) and request and response counts per server in each upstream group
	UpstreamZones map[string][]upstreamZone `json:"upstreamZones"`
}

type serverZone struct {
	RequestCounter float64  `json:"requestCounter"`
	InBytes        float64  `json:"inBytes"`
	OutBytes       float64  `json:"outBytes"`
	Responses      response `json:"responses"`
	Cache          cache    `json:"cache"`
}

type filterZone struct {
	RequestCounter float64  `json:"requestCounter"`
	InBytes        float64  `json:"inBytes"`
	OutBytes       float64  `json:"outBytes"`
	Cache          cache    `json:"cache"`
	Responses      response `json:"responses"`
}

type upstreamZone struct {
	Responses      response      `json:"responses"`
	Server         string        `json:"server"`
	RequestCounter float64       `json:"requestCounter"`
	InBytes        float64       `json:"inBytes"`
	OutBytes       float64       `json:"outBytes"`
	ResponseMsec   float64       `json:"responseMsec"`
	Weight         float64       `json:"weight"`
	MaxFails       float64       `json:"maxFails"`
	FailTimeout    float64       `json:"failTimeout"`
	Backup         BoolToFloat64 `json:"backup"`
	Down           BoolToFloat64 `json:"down"`
}

type cache struct {
	Miss        float64 `json:"miss"`
	Bypass      float64 `json:"bypass"`
	Expired     float64 `json:"expired"`
	Stale       float64 `json:"stale"`
	Updating    float64 `json:"updating"`
	Revalidated float64 `json:"revalidated"`
	Hit         float64 `json:"hit"`
	Scarce      float64 `json:"scarce"`
}

type response struct {
	OneXx  float64 `json:"1xx"`
	TwoXx  float64 `json:"2xx"`
	TheeXx float64 `json:"3xx"`
	FourXx float64 `json:"4xx"`
	FiveXx float64 `json:"5xx"`
}

type connections struct {
	Active   float64 `json:"active"`
	Reading  float64 `json:"reading"`
	Writing  float64 `json:"writing"`
	Waiting  float64 `json:"waiting"`
	Accepted float64 `json:"accepted"`
	Handled  float64 `json:"handled"`
	Requests float64 `json:"requests"`
}

// BoolToFloat64 ...
type BoolToFloat64 float64

// UnmarshalJSON ...
func (bit BoolToFloat64) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "1" || asString == "true" {
		bit = 1
	} else if asString == "0" || asString == "false" {
		bit = 0
	} else {
		return fmt.Errorf(fmt.Sprintf("boolean unmarshal error: invalid input %s", asString))
	}
	return nil
}

func getNginxStatus(ngxHealthPort int, ngxStatusPath string) (*basicStatus, error) {
	url := fmt.Sprintf("http://localhost:%v%v", ngxHealthPort, ngxStatusPath)
	glog.V(3).Infof("start scrapping url: %v", url)

	data, err := httpBody(url)

	if err != nil {
		return nil, fmt.Errorf("unexpected error scraping nginx status page: %v", err)
	}

	return parse(string(data)), nil
}

func httpBody(url string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unexpected error scraping nginx : %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unexpected error scraping nginx (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected error scraping nginx (status %v)", resp.StatusCode)
	}

	return data, nil
}

func getNginxVtsMetrics(ngxHealthPort int, ngxVtsPath string) (*vts, error) {
	url := fmt.Sprintf("http://localhost:%v%v", ngxHealthPort, ngxVtsPath)
	glog.V(3).Infof("start scrapping url: %v", url)

	data, err := httpBody(url)

	if err != nil {
		return nil, fmt.Errorf("unexpected error scraping nginx vts (%v)", err)
	}

	var vts *vts
	err = json.Unmarshal(data, &vts)
	if err != nil {
		return nil, fmt.Errorf("unexpected error json unmarshal (%v)", err)
	}
	glog.V(3).Infof("scrap returned : %v", vts)
	return vts, nil
}

func parse(data string) *basicStatus {
	acr := ac.FindStringSubmatch(data)
	sahrr := sahr.FindStringSubmatch(data)
	readingr := reading.FindStringSubmatch(data)
	writingr := writing.FindStringSubmatch(data)
	waitingr := waiting.FindStringSubmatch(data)

	return &basicStatus{
		toInt(acr, 1),
		toInt(sahrr, 1),
		toInt(sahrr, 2),
		toInt(sahrr, 3),
		toInt(readingr, 1),
		toInt(writingr, 1),
		toInt(waitingr, 1),
	}
}

func toInt(data []string, pos int) int {
	if len(data) == 0 {
		return 0
	}
	if pos > len(data) {
		return 0
	}
	if v, err := strconv.Atoi(data[pos]); err == nil {
		return v
	}
	return 0
}
