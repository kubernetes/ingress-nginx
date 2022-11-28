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

package grpcclient

import (
	"k8s.io/klog/v2"
)

func (c *Client) EventService() {

	stream, err := c.EventClient.PublishEvent(c.ctx)
	if err != nil {
		klog.Errorf("error creating event client: %s", err)
		return
	}

	for msg := range c.EventCh {
		message := &msg
		message.Backend = c.Backendname
		klog.Infof("sending message %+v", message)
		if err := stream.Send(message); err != nil {
			klog.Errorf("error sending message: %v", err)
			return
		}
	}
}
