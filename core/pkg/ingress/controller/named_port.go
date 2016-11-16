/*
Copyright 2015 The Kubernetes Authors.

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

package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	podutil "k8s.io/kubernetes/pkg/api/pod"
	"k8s.io/kubernetes/pkg/labels"

	"k8s.io/ingress/core/pkg/ingress/annotations/service"
)

// checkSvcForUpdate verifies if one of the running pods for a service contains
// named port. If the annotation in the service does not exists or is not equals
// to the port mapping obtained from the pod the service must be updated to reflect
// the current state
func (ic *GenericController) checkSvcForUpdate(svc *api.Service) error {
	// get the pods associated with the service
	// TODO: switch this to a watch
	pods, err := ic.cfg.Client.Pods(svc.Namespace).List(api.ListOptions{
		LabelSelector: labels.Set(svc.Spec.Selector).AsSelector(),
	})

	if err != nil {
		return fmt.Errorf("error searching service pods %v/%v: %v", svc.Namespace, svc.Name, err)
	}

	if len(pods.Items) == 0 {
		return nil
	}

	namedPorts := map[string]string{}

	// we need to check only one pod searching for named ports
	pod := &pods.Items[0]
	glog.V(4).Infof("checking pod %v/%v for named port information", pod.Namespace, pod.Name)
	for i := range svc.Spec.Ports {
		servicePort := &svc.Spec.Ports[i]

		_, err := strconv.Atoi(servicePort.TargetPort.StrVal)
		if err != nil {
			portNum, err := podutil.FindPort(pod, servicePort)
			if err != nil {
				glog.V(4).Infof("failed to find port for service %s/%s: %v", portNum, svc.Namespace, svc.Name, err)
				continue
			}

			if servicePort.TargetPort.StrVal == "" {
				continue
			}

			namedPorts[servicePort.TargetPort.StrVal] = fmt.Sprintf("%v", portNum)
		}
	}

	if svc.ObjectMeta.Annotations == nil {
		svc.ObjectMeta.Annotations = map[string]string{}
	}

	curNamedPort := svc.ObjectMeta.Annotations[service.NamedPortAnnotation]
	if len(namedPorts) > 0 && !reflect.DeepEqual(curNamedPort, namedPorts) {
		data, _ := json.Marshal(namedPorts)

		newSvc, err := ic.cfg.Client.Services(svc.Namespace).Get(svc.Name)
		if err != nil {
			return fmt.Errorf("error getting service %v/%v: %v", svc.Namespace, svc.Name, err)
		}

		if newSvc.ObjectMeta.Annotations == nil {
			newSvc.ObjectMeta.Annotations = map[string]string{}
		}

		newSvc.ObjectMeta.Annotations[service.NamedPortAnnotation] = string(data)
		glog.Infof("updating service %v with new named port mappings", svc.Name)
		_, err = ic.cfg.Client.Services(svc.Namespace).Update(newSvc)
		if err != nil {
			return fmt.Errorf("error syncing service %v/%v: %v", svc.Namespace, svc.Name, err)
		}

		return nil
	}

	return nil
}
