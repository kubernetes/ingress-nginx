/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"

	"k8s.io/klog/v2"

	"gopkg.in/mcuadros/go-syslog.v2"
)

func logger(address string) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()

	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)
	if err := server.ListenUDP(address); err != nil {
		klog.Fatalf("failed bind internal syslog: %w", err)
	}

	if err := server.Boot(); err != nil {
		klog.Fatalf("failed to boot internal syslog: %w", err)
	}
	klog.Infof("Is Chrooted, starting logger")

	for logParts := range channel {
		fmt.Printf("%s\n", logParts["content"])
	}

	server.Wait()
	klog.Infof("Stopping logger")

}
