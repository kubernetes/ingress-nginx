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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

var (
	ac      = regexp.MustCompile(`Active connections: (\d+)`)
	sahr    = regexp.MustCompile(`(\d+)\s(\d+)\s(\d+)`)
	reading = regexp.MustCompile(`Reading: (\d+)`)
	writing = regexp.MustCompile(`Writing: (\d+)`)
	waiting = regexp.MustCompile(`Waiting: (\d+)`)
)

type nginxStatus struct {
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

type Vts struct {
	NginxVersion  string                    `json:"nginxVersion"`
	LoadMsec      int                       `json:"loadMsec"`
	NowMsec       int                       `json:"nowMsec"`
	Connections   Connections               `json:"connections"`
	ServerZones   map[string]ServerZones    `json:"serverZones"`
	FilterZones   map[string]FilterZone     `json:"filterZones"`
	UpstreamZones map[string][]UpstreamZone `json:"upstreamZones"`
}

type ServerZones struct {
	RequestCounter float64    `json:"requestCounter"`
	InBytes        float64    `json:"inBytes"`
	OutBytes       float64    `json:"outBytes"`
	Responses      Response   `json:"responses"`
	OverCounts     OverCounts `json:"overCounts"`
}

type OverCounts struct {
	RequestCounter float64 `json:"requestCounter"`
	InBytes        float64 `json:"inBytes"`
	OutBytes       float64 `json:"outBytes"`
	OneXx          float64 `json:"1xx"`
	TwoXx          float64 `json:"2xx"`
	TheeXx         float64 `json:"3xx"`
	FourXx         float64 `json:"4xx"`
	FiveXx         float64 `json:"5xx"`
}

type FilterZone struct {
}

type UpstreamZone struct {
	Server         string        `json:"server"`
	RequestCounter float64       `json:"requestCounter"`
	InBytes        float64       `json:"inBytes"`
	OutBytes       float64       `json:"outBytes"`
	Responses      Response      `json:"responses"`
	OverCounts     OverCounts    `json:"overcounts"`
	ResponseMsec   float64       `json:"responseMsec"`
	Weight         float64       `json:"weight"`
	MaxFails       float64       `json:"maxFails"`
	FailTimeout    float64       `json:"failTimeout"`
	Backup         BoolToFloat64 `json:"backup"`
	Down           BoolToFloat64 `json:"down"`
}

type Response struct {
	OneXx            float64 `json:"1xx"`
	TwoXx            float64 `json:"2xx"`
	TheeXx           float64 `json:"3xx"`
	FourXx           float64 `json:"4xx"`
	FiveXx           float64 `json:"5xx"`
	CacheMiss        float64 `json:"miss"`
	CacheBypass      float64 `json:"bypass"`
	CacheExpired     float64 `json:"expired"`
	CacheStale       float64 `json:"stale"`
	CacheUpdating    float64 `json:"updating"`
	CacheRevalidated float64 `json:"revalidated"`
	CacheHit         float64 `json:"hit"`
	CacheScarce      float64 `json:"scarce"`
}

type Connections struct {
	Active   float64 `json:"active"`
	Reading  float64 `json:"reading"`
	Writing  float64 `json:"writing"`
	Waiting  float64 `json:"waiting"`
	Accepted float64 `json:"accepted"`
	Handled  float64 `json:"handled"`
	Requests float64 `json:"requests"`
}

type BoolToFloat64 float64

func (bit BoolToFloat64) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "1" || asString == "true" {
		bit = 1
	} else if asString == "0" || asString == "false" {
		bit = 0
	} else {
		return fmt.Errorf(fmt.Sprintf("Boolean unmarshal error: invalid input %s", asString))
	}
	return nil
}

func getNginxStatus() (*nginxStatus, error) {
	data, err := httpBody(fmt.Sprintf("http://localhost:%v%v", ngxHealthPort, ngxStatusPath))

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
func getNginxVtsMetrics() (*Vts, error) {
	data, err := httpBody(fmt.Sprintf("http://localhost:%v%v", ngxHealthPort, ngxVtsPath))

	if err {
		return nil, fmt.Errorf("unexpected error scraping nginx vts (%v)", err)
	}

	var vts Vts
	err = json.Unmarshal(data, &vts)
	if err != nil {
		return nil, fmt.Errorf("unexpected error json unmarshal (%v)", err)
	}

	return &vts, nil
}

func parse(data string) *nginxStatus {
	acr := ac.FindStringSubmatch(data)
	sahrr := sahr.FindStringSubmatch(data)
	readingr := reading.FindStringSubmatch(data)
	writingr := writing.FindStringSubmatch(data)
	waitingr := waiting.FindStringSubmatch(data)

	return &nginxStatus{
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
