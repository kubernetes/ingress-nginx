/*
Copyright 2018 The Kubernetes Authors.

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

package influxdb

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

type influxdb struct {
	r resolver.Resolver
}

// Config contains the IfluxDB configuration to be used in the Ingress
type Config struct {
	InfluxDBEnabled     bool   `json:"influxDBEnabled"`
	InfluxDBMeasurement string `json:"influxDBMeasurement"`
	InfluxDBPort        string `json:"influxDBPort"`
	InfluxDBHost        string `json:"influxDBHost"`
	InfluxDBServerName  string `json:"influxDBServerName"`
}

// NewParser creates a new InfluxDB annotation parser
func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return influxdb{r}
}

// Parse parses the annotations to look for InfluxDB configurations
func (c influxdb) Parse(ing *extensions.Ingress) (interface{}, error) {
	influxdbEnabled, err := parser.GetBoolAnnotation("enable-influxdb", ing)
	if err != nil {
		influxdbEnabled = false
	}

	influxdbMeasurement, err := parser.GetStringAnnotation("influxdb-measurement", ing)
	if err != nil {
		influxdbMeasurement = "default"
	}

	influxdbPort, err := parser.GetStringAnnotation("influxdb-port", ing)
	if err != nil {
		// This is not the default 8086 port but the port usually used to expose
		// influxdb in UDP, the module uses UDP to talk to influx via the line protocol.
		influxdbPort = "8089"
	}

	influxdbHost, err := parser.GetStringAnnotation("influxdb-host", ing)
	if err != nil {
		influxdbHost = "127.0.0.1"
	}

	influxdbServerName, err := parser.GetStringAnnotation("influxdb-server-name", ing)
	if err != nil {
		influxdbServerName = "nginx-ingress"
	}

	return &Config{
		InfluxDBEnabled:     influxdbEnabled,
		InfluxDBMeasurement: influxdbMeasurement,
		InfluxDBPort:        influxdbPort,
		InfluxDBHost:        influxdbHost,
		InfluxDBServerName:  influxdbServerName,
	}, nil
}

// Equal tests for equality between two Config types
func (e1 *Config) Equal(e2 *Config) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.InfluxDBEnabled != e2.InfluxDBEnabled {
		return false
	}
	if e1.InfluxDBPort != e2.InfluxDBPort {
		return false
	}
	if e1.InfluxDBHost != e2.InfluxDBHost {
		return false
	}
	if e1.InfluxDBServerName != e2.InfluxDBServerName {
		return false
	}

	return true
}
